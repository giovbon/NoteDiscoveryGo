// Package notediscovery embeds the web assets (frontend, themes, locales) so the
// binary can run standalone, while still allowing a `web/` directory next to the
// working directory to override what was compiled in (edit assets without rebuild).
package notediscovery

import (
	"embed"
	"io/fs"
	"os"
)

// Runtime assets only. The frontend source tree also holds dev-only files
// (src/, test/, e2e/, package.json, node_modules/, Playwright output); a bare
// `all:web` would embed all of them and make the build fail whenever Playwright
// rewrites test-results/ mid-compile. Listing the shipped files keeps the
// binary lean and the build deterministic.
//
//go:embed all:web/frontend/index.html
//go:embed all:web/frontend/login.html
//go:embed all:web/frontend/app.js
//go:embed all:web/frontend/app.js.map
//go:embed all:web/frontend/sw.js
//go:embed all:web/frontend/editor-markdown-continue.js
//go:embed all:web/frontend/tailwind.css
//go:embed all:web/frontend/markdown-preview.css
//go:embed all:web/frontend/manifest.json
//go:embed all:web/frontend/favicon.svg
//go:embed all:web/frontend/logo.svg
//go:embed all:web/frontend/partials
//go:embed all:web/frontend/vendor
//go:embed all:web/themes
//go:embed all:web/locales
var embeddedWeb embed.FS

// AssetFS returns the filesystem holding web/, themes/ and locales/.
// A web/ directory on disk wins over the embedded copy.
func AssetFS() fs.FS {
	if st, err := os.Stat("web"); err == nil && st.IsDir() {
		return os.DirFS("web")
	}
	sub, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		return embeddedWeb
	}
	return sub
}
