package main

import (
	"fmt"
	"sort"

	"golang.org/x/tools/go/packages"
)

const version = "enumx-gen v3.1."

func main() {
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

	pkgs = loadPackage(".")
	targetPkg := pkgs[0]
	generateRegistry(targetPkg, allDefs)
}

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
