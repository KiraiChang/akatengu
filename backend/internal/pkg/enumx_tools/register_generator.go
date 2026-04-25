package enumx_tools

import (
	"fmt"
	"sort"

	"golang.org/x/tools/go/packages"
)

type RegistryTemplateData struct {
	Package string
	Imports []Import
	Enums   []EnumRender
}

type Import struct {
	Alias string
	Path  string
}

type EnumRender struct {
	Alias            string
	PkgName          string
	RegisterFuncName string
}

func runRegistry(defs []*EnumDef) error {
	pkgs := loadPackage(".")
	targetPkg := pkgs[0]
	data := buildRegistryData(targetPkg, defs)

	out, err := renderTemplate("templates/registry.tmpl", data)
	if err != nil {
		return err
	}

	return writeFile("enums_registry_gen.go", out)
}

func buildRegistryData(targetPkg *packages.Package, defs []*EnumDef) RegistryTemplateData {
	importMap := map[string]string{}
	var enums []EnumRender

	for _, d := range defs {
		external := d.PkgPath != targetPkg.PkgPath
		registFuncName := ""
		if external {
			importMap[d.PkgPath] = d.PkgName
			registFuncName = d.PkgName + ".All" + d.Alias
		} else {
			registFuncName = "All" + d.Alias
		}

		enums = append(enums, EnumRender{
			Alias:            d.Alias,
			PkgName:          d.PkgName,
			RegisterFuncName: registFuncName,
		})
		fmt.Println("def.PkgName : "+d.PkgName, "def.Alias : "+d.Alias)
	}

	var imports []Import
	for path, alias := range importMap {
		imports = append(imports, Import{Alias: alias, Path: path})
	}

	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path < imports[j].Path
	})

	return RegistryTemplateData{
		Package: targetPkg.Name,
		Imports: imports,
		Enums:   enums,
	}
}
