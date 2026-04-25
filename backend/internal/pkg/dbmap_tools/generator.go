package dbmap_tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/packages"
)

const (
	sqlcdbPattern   = "./internal/database/sqlcdb"
	modelDbPattern  = "./internal/model/db/..."
	dbmapconvImport = "akatengu/internal/pkg/dbmapconv"
)

// Run is the entry point for the dbmap code generator.
// It locates the module root automatically, so it can be invoked from any
// subdirectory via //go:generate.
func Run() error {
	root, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("find module root: %w", err)
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("chdir to module root %s: %w", root, err)
	}

	fmt.Println("dbmap-gen scanning...")

	sqlcPkgs := loadPackage(sqlcdbPattern)
	sqlcStructs := collectSqlcStructs(sqlcPkgs[0])

	modelPkgs := loadPackage(modelDbPattern)

	type pkgGroup struct {
		pkg  *packages.Package
		defs []*StructDef
	}

	var groups []pkgGroup

	for _, pkg := range modelPkgs {
		defs := collectStructDefs(pkg)
		if len(defs) == 0 {
			continue
		}
		groups = append(groups, pkgGroup{pkg: pkg, defs: defs})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].pkg.PkgPath < groups[j].pkg.PkgPath
	})

	for _, g := range groups {
		if err := generateForPackage(g.pkg, g.defs, sqlcStructs); err != nil {
			return err
		}
	}

	fmt.Println("dbmap-gen done.")
	return nil
}

func generateForPackage(
	pkg *packages.Package,
	defs []*StructDef,
	sqlcStructs map[string]map[string]FieldDef,
) error {
	data, needsConv := buildFileData(pkg, defs, sqlcStructs)
	if len(data.Mappers) == 0 {
		return nil
	}
	data.NeedsConv = needsConv

	raw, err := renderTemplate("templates/mapper.tmpl", data)
	if err != nil {
		return fmt.Errorf("render template for %s: %w", pkg.Name, err)
	}

	// Output lives alongside the model files — same package, same directory.
	outFile := filepath.Join(pkg.Dir, "dbmap_gen.go")
	if err := writeFile(outFile, formatSource(raw)); err != nil {
		return fmt.Errorf("write %s: %w", outFile, err)
	}

	fmt.Printf("  wrote %s\n", outFile)
	return nil
}

// findModuleRoot walks up from the current working directory until it finds
// a directory that contains a go.mod file.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found (started from %s)", dir)
		}
		dir = parent
	}
}

func buildFileData(
	pkg *packages.Package,
	defs []*StructDef,
	sqlcStructs map[string]map[string]FieldDef,
) (FileTemplateData, bool) {
	needsConv := false
	var mappers []MapperData

	for _, def := range defs {
		for _, target := range def.SqlcTargets {
			tgtByTag, ok := sqlcStructs[target]
			if !ok {
				fmt.Printf("  WARNING: sqlcdb.%s not found (referenced by %s.%s)\n",
					target, def.PkgName, def.TypeName)
				continue
			}

			mappings, conv := buildMappings(def.Fields, tgtByTag)
			if conv {
				needsConv = true
			}

			mappers = append(mappers, MapperData{
				SrcTypeName: def.TypeName,
				TgtTypeName: target,
				Mappings:    mappings,
			})
		}
	}

	return FileTemplateData{
		PkgName: pkg.Name,
		Mappers: mappers,
	}, needsConv
}