package repos

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/regexp"
	bytesize "github.com/inhies/go-bytesize"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

// excludeFunc takes either a generic object and returns true if the repo should be excluded. In
// the case of repo sourcing it will take a repository name, ID, or the repo itself as input.
type excludeFunc func(input any) bool

// excludeBuilder builds an excludeFunc.
type excludeBuilder struct {
	exact    map[string]struct{}
	patterns []*regexp.Regexp
	generic  []excludeFunc
	err      error
}

// Exact will case-insensitively exclude the string name.
func (e *excludeBuilder) Exact(name string) {
	if e.exact == nil {
		e.exact = map[string]struct{}{}
	}
	if name == "" {
		return
	}
	e.exact[strings.ToLower(name)] = struct{}{}
}

// Pattern will exclude strings matching the regex pattern.
func (e *excludeBuilder) Pattern(pattern string) {
	if pattern == "" {
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		e.err = err
		return
	}
	e.patterns = append(e.patterns, re)
}

// Generic registers the passed in exclude function that will be used to determine whether a repo
// should be excluded.
func (e *excludeBuilder) Generic(ef excludeFunc) {
	if ef == nil {
		return
	}
	e.generic = append(e.generic, ef)
}

// Build will return an excludeFunc based on the previous calls to Exact, Pattern, and
// Generic.
func (e *excludeBuilder) Build() (excludeFunc, error) {
	return func(input any) bool {
		if inputString, ok := input.(string); ok {
			if _, ok := e.exact[strings.ToLower(inputString)]; ok {
				return true
			}

			for _, re := range e.patterns {
				if re.MatchString(inputString) {
					return true
				}
			}
		} else {
			for _, ef := range e.generic {
				if ef(input) {
					return true
				}
			}
		}

		return false
	}, e.err
}

func ParseGitHubExcludeRule(rule *schema.ExcludedGitHubRepo) (excludeFunc, error) {
	starsConstraint, err := buildStarsConstraintsExcludeFn(rule.Stars)
	if err != nil {
		return nil, err
	}

	sizeConstraint, err := buildSizeConstraintsExcludeFn(rule.Size)
	if err != nil {
		return nil, err
	}

	return func(repo any) bool {
		githubRepo, ok := repo.(github.Repository)
		if !ok {
			return false
		}

		st := starsConstraint(githubRepo)
		si := sizeConstraint(githubRepo)
		fmt.Printf("st=%+v, si=%+v\n", st, si)
		return st && si
	}, nil
}

type gitHubExcludeFunc func(github.Repository) bool

func githubExcludeNoop(repo github.Repository) bool { return false }

// TODO: Put these in the schema?
var starsConstraintRegex = regexp.MustCompile(`([<>=]{1,2})\s*(\d+)`)
var sizeConstraintRegex = regexp.MustCompile(`([<>=]{1,2})\s*(\d+\s*\w+)`)

func buildStarsConstraintsExcludeFn(constraint string) (gitHubExcludeFunc, error) {
	if constraint == "" {
		return githubExcludeNoop, nil
	}

	matches := starsConstraintRegex.FindStringSubmatch(constraint)
	if matches == nil {
		return nil, fmt.Errorf("invalid stars constraint format: %q", constraint)
	}

	operator := matches[1]
	count, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	var fn gitHubExcludeFunc = nil
	switch operator {
	case "<":
		fn = func(r github.Repository) bool { fmt.Println("yo1"); return r.StargazerCount < count }
	case ">":
		fn = func(r github.Repository) bool { fmt.Println("yo2"); return r.StargazerCount > count }
	case "<=":
		fn = func(r github.Repository) bool { fmt.Println("yo3"); return r.StargazerCount <= count }
	case ">=":
		fn = func(r github.Repository) bool { fmt.Println("yo4"); return r.StargazerCount >= count }
	default:
		return nil, fmt.Errorf("invalid operator %q for stars constraint", operator)
	}

	return fn, nil
}

func buildSizeConstraintsExcludeFn(constraint string) (gitHubExcludeFunc, error) {
	if constraint == "" {
		return githubExcludeNoop, nil
	}

	sizeMatch := sizeConstraintRegex.FindStringSubmatch(constraint)
	if sizeMatch == nil {
		return nil, fmt.Errorf("invalid size constraint format: %q", constraint)
	}

	operator := sizeMatch[1]
	size, err := bytesize.Parse(sizeMatch[2])
	if err != nil {
		return nil, err
	}

	fmt.Printf("customer size input: %s\n", size)

	var fn gitHubExcludeFunc = nil
	switch operator {
	case "<":
		fn = func(r github.Repository) bool {
			repoSize, err := bytesize.Parse(fmt.Sprintf("%d KB", r.SizeKibiBytes))
			if err != nil {
				panic(err)
			}
			return repoSize < size
		}
	case ">":
		fn = func(r github.Repository) bool {
			repoSize, err := bytesize.Parse(fmt.Sprintf("%d KB", r.SizeKibiBytes))
			if err != nil {
				panic(err)
			}
			return repoSize > size
		}
	case "<=":
		fn = func(r github.Repository) bool {
			repoSize, err := bytesize.Parse(fmt.Sprintf("%d KB", r.SizeKibiBytes))
			if err != nil {
				panic(err)
			}
			return repoSize <= size
		}
	case ">=":
		fn = func(r github.Repository) bool {
			kilobyte := float64(r.SizeKibiBytes)
			str := fmt.Sprintf("%f KB", kilobyte)
			repoSize, err := bytesize.Parse(str)
			if err != nil {
				panic(err)
			}
			fmt.Printf("size kibibytes=%d, str=%q, kilobytes=%+v, reposize=%+v, comparison=%+v\n",
				r.SizeKibiBytes,
				str,
				kilobyte,
				repoSize,
				repoSize >= size,
			)
			return repoSize >= size
		}
	default:
		return nil, fmt.Errorf("invalid operator %q for stars constraint", operator)
	}

	return fn, nil
}
