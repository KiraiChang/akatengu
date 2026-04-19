package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

const version = "enumx-gen v3.1"

func Run() error {
	fmt.Println(fmt.Sprintf("%s scanning...", version))

	pkgs := loadPackage("./...")
	allDefs := []*EnumDef{}
	for _, pkg := range pkgs {

		aliases := collectAliases(pkg)
		defs := collectEnums(pkg, aliases)

		if len(defs) == 0 {
			continue
		}

		for _, d := range defs {
			d.PkgPath = pkg.PkgPath // 🔥 關鍵
			d.PkgName = pkg.Name    // enums / domain
		}

		validate(defs)

		generate(pkg, defs)
		allDefs = append(allDefs, mapToSortedSlice(defs)...)
	}

	return runRegistry(allDefs)
}

func mapToSortedSlice(m map[string]*EnumDef) []*EnumDef {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := make([]*EnumDef, 0, len(m))
	for _, k := range keys {
		result = append(result, m[k])
	}

	return result
}

func generate(pkg *packages.Package, defs map[string]*EnumDef) {
	for _, def := range defs {
		out, err := renderTemplate("templates/enums.tmpl", def)
		if err != nil {
			return
		}

		file := fileNameFromType(def.TypeName) + "_gen.go"

		full := filepath.Join(pkg.Dir, file)
		writeFile(full, out)
	}
}

func fileNameFromType(typeName string) string {
	// aggregateTypeVal → aggregate_type
	name := strings.TrimSuffix(typeName, "Val")

	var out []rune
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}
