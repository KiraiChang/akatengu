package dbmap_tools

import (
	"bytes"
	"embed"
	"go/format"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/go/packages"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

func loadPackage(pattern string) []*packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedModule,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		panic("no packages found for: " + pattern)
	}
	return pkgs
}

func renderTemplate(path string, data any) ([]byte, error) {
	tmpl, err := template.New(filepath.Base(path)).
		ParseFS(templatesFS, path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatSource(src []byte) []byte {
	formatted, err := format.Source(src)
	if err != nil {
		return src // return unformatted; compilation will surface the real error
	}
	return formatted
}

func writeFile(filename string, content []byte) error {
	return os.WriteFile(filename, content, 0666)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}