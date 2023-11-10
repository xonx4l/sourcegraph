package dist

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

//go:embed *
var assetsFS embed.FS

var (
	webBuildManifestOnce sync.Once
	webBuildManifest     *assets.WebBuildManifest
	webBuildManifestErr  error
)

func init() {
	// Sets the global assets provider.
	assets.Provider = Provider{}
}

type Provider struct{}

func (p Provider) LoadWebBuildManifest() (*assets.WebBuildManifest, error) {
	webBuildManifestOnce.Do(func() {
		f, err := assetsFS.Open("vite-manifest.json")
		if err != nil {
			webBuildManifestErr = errors.Wrap(err, "read manifest file")
			return
		}
		defer f.Close()

		manifestContent, err := io.ReadAll(f)
		if err != nil {
			webBuildManifestErr = errors.Wrap(err, "read manifest file")
			return
		}

		if err := json.Unmarshal(manifestContent, &webBuildManifest); err != nil {
			webBuildManifestErr = errors.Wrap(err, "unmarshal manifest json")
			return
		}
	})
	return webBuildManifest, webBuildManifestErr
}

var providerAssets = http.FS(assetsFS)

func (p Provider) Assets() http.FileSystem {
	return providerAssets
}
