package servegit

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/gitservice"
)

type Serve struct {
	Addr   string
	Root   string
	Logger log.Logger
}

func (s *Serve) Start() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Wrap(err, "listen")
	}

	// Update Addr to what listener actually used.
	s.Addr = ln.Addr().String()

	s.Logger.Info("serving git repositories", log.String("url", "http://"+s.Addr), log.String("root", s.Root))

	srv := &http.Server{Handler: s.handler()}

	// We have opened the listener, now start serving connections in the
	// background.
	go func() {
		if err := srv.Serve(ln); err == http.ErrServerClosed {
			s.Logger.Info("http serve closed")
		} else {
			s.Logger.Error("http serve failed", log.Error(err))
		}
	}()

	// Also listen for shutdown signals in the background. We don't need
	// graceful shutdown since this only runs in app and the only clients of
	// the server will also be shutdown at the same time.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		<-c
		if err := srv.Close(); err != nil {
			s.Logger.Error("failed to Close http serve", log.Error(err))
		}
	}()

	return nil
}

var indexHTML = template.Must(template.New("").Parse(`<html>
<head><title>src serve-git</title></head>
<body>
<h2>src serve-git</h2>
<pre>
{{.Explain}}
<ul>{{range .Links}}
<li><a href="{{.}}">{{.}}</a></li>
{{- end}}
</ul>
</pre>
</body>
</html>`))

type Repo struct {
	Name      string
	URI       string
	ClonePath string
}

func (s *Serve) handler() http.Handler {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := indexHTML.Execute(w, map[string]interface{}{
			"Explain": explainAddr(s.Addr),
			"Links": []string{
				"/v1/list-repos",
				"/repos/",
			},
		})
		if err != nil {
			s.Logger.Debug("failed to return / response", log.Error(err))
		}
	})

	mux.HandleFunc("/v1/list-repos", func(w http.ResponseWriter, r *http.Request) {
		repos, err := s.Repos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := struct {
			Items []Repo
		}{
			Items: repos,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(&resp)
	})

	fs := http.FileServer(http.Dir(s.Root))
	svc := &gitservice.Handler{
		Dir: func(name string) string {
			return filepath.Join(s.Root, filepath.FromSlash(name))
		},
		Trace: func(ctx context.Context, svc, repo, protocol string) func(error) {
			start := time.Now()
			return func(err error) {
				s.Logger.Debug("git service", log.String("svc", svc), log.String("protocol", protocol), log.String("repo", repo), log.Duration("duration", time.Since(start)), log.Error(err))
			}
		},
	}
	mux.Handle("/repos/", http.StripPrefix("/repos/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use git service if git is trying to clone. Otherwise show http.FileServer for convenience
		for _, suffix := range []string{"/info/refs", "/git-upload-pack"} {
			if strings.HasSuffix(r.URL.Path, suffix) {
				svc.ServeHTTP(w, r)
				return
			}
		}
		fs.ServeHTTP(w, r)
	})))

	return http.HandlerFunc(mux.ServeHTTP)
}

// Checks if git thinks the given path is a valid .git folder for a repository
func isBareRepo(path string) bool {
	c := exec.Command("git", "--git-dir", path, "rev-parse", "--is-bare-repository")
	c.Dir = path
	out, err := c.CombinedOutput()

	if err != nil {
		return false
	}

	return string(out) != "false\n"
}

// Check if git thinks the given path is a proper git checkout
func isGitRepo(path string) bool {
	// Executing git rev-parse --git-dir in the root of a worktree returns .git
	c := exec.Command("git", "rev-parse", "--git-dir")
	c.Dir = path
	out, err := c.CombinedOutput()

	if err != nil {
		return false
	}

	return string(out) == ".git\n"
}

// Repos returns a slice of all the git repositories it finds.
func (s *Serve) Repos() ([]Repo, error) {
	var repos []Repo
	var reposRootIsRepo bool

	root, err := filepath.EvalSymlinks(s.Root)
	if err != nil {
		s.Logger.Warn("ignoring error searching", log.String("path", root), log.Error(err))
		return nil, nil
	}

	err = filepath.WalkDir(root, func(path string, fi fs.DirEntry, fileErr error) error {
		if fileErr != nil {
			s.Logger.Warn("ignoring error searching", log.String("path", path), log.Error(fileErr))
			return nil
		}
		if !fi.IsDir() {
			return nil
		}

		// Previously we recursed into bare repositories which is why this check was here.
		// Now we use this as a sanity check to make sure we didn't somehow stumble into a .git dir.
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// Check whether a particular directory is a repository or not.
		//
		// Valid paths are either bare repositories or git worktrees.
		isBare := isBareRepo(path)
		isGit := isGitRepo(path)

		if !isGit && !isBare {
			s.Logger.Debug("not a repository root", log.String("path", path))
			return nil
		}

		subpath, err := filepath.Rel(root, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			return errors.Wrapf(err, "filepath.Walk returned %s which is not relative to %s", path, root)
		}

		name := filepath.ToSlash(subpath)
		reposRootIsRepo = reposRootIsRepo || name == "."
		cloneURI := pathpkg.Join("/repos", name)
		clonePath := cloneURI

		// Regular git repos won't clone without the full path to the .git directory.
		if isGit {
			clonePath += "/.git"
		}

		repos = append(repos, Repo{
			Name:      name,
			URI:       cloneURI,
			ClonePath: clonePath,
		})

		// At this point we know the directory is either a git repo or a bare git repo,
		// we don't need to recurse further to save time.
		// TODO: Look into whether it is useful to support git submodules
		return filepath.SkipDir
	})

	if err != nil {
		return nil, err
	}

	if !reposRootIsRepo {
		return repos, nil
	}

	// Update all names to be relative to the parent of reposRoot. This is to
	// give a better name than "." for repos root
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, errors.Errorf("failed to get the absolute path of reposRoot: %w", err)
	}
	rootName := filepath.Base(abs)
	for i := range repos {
		repos[i].Name = pathpkg.Join(rootName, repos[i].Name)
	}

	return repos, nil
}

func explainAddr(addr string) string {
	return fmt.Sprintf(`Serving the repositories at http://%s.

See https://docs.sourcegraph.com/admin/external_service/src_serve_git for
instructions to configure in Sourcegraph.
`, addr)
}
