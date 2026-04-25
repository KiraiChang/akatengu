package enumx_tools

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"golang.org/x/tools/go/packages"
)

//go:embed  templates/*.tmpl
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
		panic("package not found")
	}

	return pkgs
}

func renderTemplate(path string, data any) ([]byte, error) {
	funcMap := template.FuncMap{}
	tmpl, err := template.New(filepath.Base(path)).
		Funcs(funcMap).
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

func writeFile(filename string, content []byte) error {
	//formatted, err := format.Source(content)
	//if err != nil {
	//	return fmt.Errorf("source fail: %w", err)
	//}
	//return os.WriteFile(filename, formatted, 0644)

	return os.WriteFile(filename, content, 0666)
}
