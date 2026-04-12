package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

type EnumDef struct {
	TypeName string
	Alias    string
	BaseType string

	PkgName string // e.g. enums / domain
	PkgPath string // e.g. your_project/internal/enums

	Consts map[string]EnumConst
}

type EnumConst struct {
	Name  string
	Label string
	Code  string
}

func collectEnums(pkg *packages.Package, aliases map[string]string) map[string]*EnumDef {
	result := map[string]*EnumDef{}
	typeMarked := map[string]bool{}
	typeBase := map[string]string{}

	// 找 enum type
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			if !isEnumMarked(gen.Doc) {
				continue
			}

			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)

				typeMarked[ts.Name.Name] = true

				if ident, ok := ts.Type.(*ast.Ident); ok {
					typeBase[ts.Name.Name] = ident.Name
				}
			}
		}
	}

	// 收集 const
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}

			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)

				if vs.Type == nil {
					continue
				}

				t, ok := vs.Type.(*ast.Ident)
				if !ok || !typeMarked[t.Name] {
					continue
				}

				def := result[t.Name]
				if def == nil {
					def = &EnumDef{
						TypeName: t.Name,
						Alias:    aliases[t.Name],
						BaseType: typeBase[t.Name],
						Consts:   map[string]EnumConst{},
					}
					result[t.Name] = def
				}

				for _, n := range vs.Names {
					def.Consts[n.Name] = EnumConst{
						Name:  n.Name,
						Label: parseTag(vs.Comment, "label"),
						Code:  parseTag(vs.Comment, "code"),
					}
				}
			}
		}
	}

	return result
}

func collectAliases(pkg *packages.Package) map[string]string {
	aliases := map[string]string{}

	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {

			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)

				if !ts.Assign.IsValid() {
					continue
				}

				idx, ok := ts.Type.(*ast.IndexExpr)
				if !ok {
					continue
				}

				sel, ok := idx.X.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Enum" {
					continue
				}

				arg, ok := idx.Index.(*ast.Ident)
				if !ok {
					continue
				}

				valueType := arg.Name
				alias := ts.Name.Name

				if old, exists := aliases[valueType]; exists {
					panic(fmt.Sprintf("duplicate alias for %s: %s / %s", valueType, old, alias))
				}

				aliases[valueType] = alias
			}
		}
	}

	fmt.Println("alias map:", aliases)

	return aliases
}

func validate(defs map[string]*EnumDef) {
	for _, d := range defs {

		if d.Alias == "" {
			panic("enumx: missing alias for " + d.TypeName)
		}

		if len(d.Consts) == 0 {
			panic("enumx: no consts for " + d.TypeName)
		}
	}
}

func isEnumMarked(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, "enumx:enum") {
			return true
		}
	}
	return false
}

func parseTag(cg *ast.CommentGroup, s string) string {
	if cg == nil {
		return ""
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, fmt.Sprintf("enumx:%s=", s)) {
			parts := strings.Split(c.Text, "=")
			return parts[len(parts)-1]
		}
	}
	return ""
}
