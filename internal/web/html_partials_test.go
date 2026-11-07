package web

import (
	"strings"
	"testing"
	"testing/fstest"

	"notediscovery/internal/config"
)

func TestExpandHTMLIncludesSimple(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html":         {Data: []byte("<main>\n<!-- @include partials/body.html -->\n</main>\n")},
		"frontend/partials/body.html": {Data: []byte("<p>hello</p>\n")},
	}

	got, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err != nil {
		t.Fatalf("expandHTMLIncludes() error = %v", err)
	}
	want := "<main>\n<p>hello</p>\n</main>\n"
	if got != want {
		t.Errorf("expandHTMLIncludes() =\n%q\nwant\n%q", got, want)
	}
}

func TestExpandHTMLIncludesNested(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html":        {Data: []byte("A\n<!-- @include partials/one.html -->\nZ\n")},
		"frontend/partials/one.html": {Data: []byte("B\n<!-- @include partials/two.html -->\nY\n")},
		"frontend/partials/two.html": {Data: []byte("C\n")},
	}

	got, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err != nil {
		t.Fatalf("expandHTMLIncludes() error = %v", err)
	}
	want := "A\nB\nC\nY\nZ\n"
	if got != want {
		t.Errorf("expandHTMLIncludes() =\n%q\nwant\n%q", got, want)
	}
}

func TestExpandHTMLIncludesPreservesCRLF(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html":      {Data: []byte("A\r\n<!-- @include partials/x.html -->\r\nZ\r\n")},
		"frontend/partials/x.html": {Data: []byte("B\r\n")},
	}

	got, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err != nil {
		t.Fatalf("expandHTMLIncludes() error = %v", err)
	}
	want := "A\r\nB\r\nZ\r\n"
	if got != want {
		t.Errorf("expandHTMLIncludes() =\n%q\nwant\n%q", got, want)
	}
}

func TestExpandHTMLIncludesMissingFile(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html": {Data: []byte("<!-- @include partials/nope.html -->\n")},
	}

	_, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err == nil || !strings.Contains(err.Error(), "nope.html") {
		t.Fatalf("expandHTMLIncludes() error = %v, want missing include error", err)
	}
}

func TestExpandHTMLIncludesRejectsTraversal(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html": {Data: []byte("<!-- @include ../secrets.html -->\n")},
	}

	_, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err == nil || !strings.Contains(err.Error(), "escapes frontend") {
		t.Fatalf("expandHTMLIncludes() error = %v, want traversal error", err)
	}
}

func TestExpandHTMLIncludesDetectsCycle(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html":      {Data: []byte("<!-- @include partials/a.html -->\n")},
		"frontend/partials/a.html": {Data: []byte("<!-- @include partials/b.html -->\n")},
		"frontend/partials/b.html": {Data: []byte("<!-- @include partials/a.html -->\n")},
	}

	_, err := expandHTMLIncludes(fsys, string(mustRead(t, fsys, "frontend/index.html")))
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expandHTMLIncludes() error = %v, want cycle error", err)
	}
}

func TestBuildIndexExpandsPartialsAndPlaceholders(t *testing.T) {
	fsys := fstest.MapFS{
		"frontend/index.html":         {Data: []byte("<title>__APP_NAME__</title>\n<!-- @include partials/body.html -->\n")},
		"frontend/partials/body.html": {Data: []byte("<p>__APP_NAME__</p>\n")},
		"locales/pt-BR.json":          {Data: []byte(`{"hello":"olá"}`)},
	}

	cfg := &config.Config{
		App: config.AppConfig{Name: "My Vault", Version: "1.0.0"},
		UI:  config.UIConfig{DefaultTheme: "dark"},
	}
	asset, err := buildIndex(fsys, cfg)
	if err != nil {
		t.Fatalf("buildIndex() error = %v", err)
	}
	body := string(asset.data)
	if !strings.Contains(body, "<title>My Vault</title>") {
		t.Errorf("app name in shell was not replaced: %q", body)
	}
	if !strings.Contains(body, "<p>My Vault</p>") {
		t.Errorf("app name inside partial was not replaced: %q", body)
	}
	if strings.Contains(body, "@include") {
		t.Errorf("include directive leaked into the rendered shell: %q", body)
	}
}

func mustRead(t *testing.T, fsys fstest.MapFS, name string) []byte {
	t.Helper()
	file, ok := fsys[name]
	if !ok {
		t.Fatalf("missing fixture %s", name)
	}
	return file.Data
}
