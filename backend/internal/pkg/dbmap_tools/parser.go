package dbmap_tools

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// collectStructDefs scans pkg for types annotated with //dbmap:sqlcdb=TypeName.
func collectStructDefs(pkg *packages.Package) []*StructDef {
	var defs []*StructDef

	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			targets := extractDbmapTargets(gen.Doc)
			if len(targets) == 0 {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				def := &StructDef{
					TypeName:    ts.Name.Name,
					SqlcTargets: targets,
					PkgName:     pkg.Name,
					PkgPath:     pkg.PkgPath,
					Dir:         pkg.Dir,
				}

				for _, field := range st.Fields.List {
					if len(field.Names) == 0 {
						continue
					}
					tag := extractDbTag(field.Tag)
					if tag == "" || tag == "-" {
						continue
					}
					typ := typeExprToString(field.Type)
					for _, name := range field.Names {
						def.Fields = append(def.Fields, FieldDef{
							GoName: name.Name,
							GoType: typ,
							DbTag:  tag,
						})
					}
				}

				defs = append(defs, def)
			}
		}
	}

	return defs
}

// collectSqlcStructs reads all struct types from the sqlcdb package, indexed by type name.
// Each entry maps db-tag → FieldDef for fast lookup during field matching.
func collectSqlcStructs(pkg *packages.Package) map[string]map[string]FieldDef {
	result := map[string]map[string]FieldDef{}

	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				byTag := map[string]FieldDef{}
				for _, field := range st.Fields.List {
					if len(field.Names) == 0 {
						continue
					}
					tag := extractDbTag(field.Tag)
					if tag == "" || tag == "-" {
						continue
					}
					typ := typeExprToString(field.Type)
					for _, name := range field.Names {
						byTag[tag] = FieldDef{
							GoName: name.Name,
							GoType: typ,
							DbTag:  tag,
						}
					}
				}
				result[ts.Name.Name] = byTag
			}
		}
	}

	return result
}

// extractDbmapTargets collects every //dbmap:sqlcdb=TypeName line in the comment group.
// Multiple lines on the same struct are all collected, enabling one-to-many mapping.
func extractDbmapTargets(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	const marker = "dbmap:sqlcdb="
	var targets []string
	for _, c := range cg.List {
		if idx := strings.Index(c.Text, marker); idx >= 0 {
			if t := strings.TrimSpace(c.Text[idx+len(marker):]); t != "" {
				targets = append(targets, t)
			}
		}
	}
	return targets
}

func extractDbTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	s := strings.Trim(tag.Value, "`")
	return reflect.StructTag(s).Get("db")
}

func typeExprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + typeExprToString(e.X)
	case *ast.SelectorExpr:
		return typeExprToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeExprToString(e.Elt)
	case *ast.MapType:
		return "map[" + typeExprToString(e.Key) + "]" + typeExprToString(e.Value)
	}
	return "any"
}