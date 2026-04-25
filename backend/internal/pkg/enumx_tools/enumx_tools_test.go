package enumx_tools

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// ── isEnumMarked ──────────────────────────────────────────────────

func TestIsEnumMarked_Nil(t *testing.T) {
	if isEnumMarked(nil) {
		t.Fatal("nil CommentGroup should return false")
	}
}

func TestIsEnumMarked_NoMarker(t *testing.T) {
	cg := commentGroup("// just a comment", "// another line")
	if isEnumMarked(cg) {
		t.Fatal("comment without enumx:enum should return false")
	}
}

func TestIsEnumMarked_HasMarker(t *testing.T) {
	cg := commentGroup("//enumx:enum")
	if !isEnumMarked(cg) {
		t.Fatal("comment with enumx:enum should return true")
	}
}

func TestIsEnumMarked_MarkerAmongOthers(t *testing.T) {
	cg := commentGroup("// doc comment", "//enumx:enum", "// another")
	if !isEnumMarked(cg) {
		t.Fatal("should return true when any comment contains enumx:enum")
	}
}

func TestIsEnumMarked_SpaceBetweenSlashAndMarker(t *testing.T) {
	cg := commentGroup("// enumx:enum")
	if !isEnumMarked(cg) {
		t.Fatal("should return true with space before enumx:enum")
	}
}

// ── parseTag ──────────────────────────────────────────────────────

func TestParseTag_Nil(t *testing.T) {
	if got := parseTag(nil, "label"); got != "" {
		t.Fatalf("nil CommentGroup: want \"\", got %q", got)
	}
}

func TestParseTag_TagAbsent(t *testing.T) {
	cg := commentGroup("// some comment")
	if got := parseTag(cg, "label"); got != "" {
		t.Fatalf("missing tag: want \"\", got %q", got)
	}
}

func TestParseTag_LabelTag(t *testing.T) {
	cg := commentGroup("//enumx:label=Active")
	got := parseTag(cg, "label")
	if got != "Active" {
		t.Fatalf("want %q, got %q", "Active", got)
	}
}

func TestParseTag_CodeTag(t *testing.T) {
	cg := commentGroup("//enumx:code=ACT")
	got := parseTag(cg, "code")
	if got != "ACT" {
		t.Fatalf("want %q, got %q", "ACT", got)
	}
}

func TestParseTag_MultipleComments_CorrectTagPicked(t *testing.T) {
	cg := commentGroup("// description", "//enumx:label=Pending", "//enumx:code=PND")
	if got := parseTag(cg, "label"); got != "Pending" {
		t.Errorf("label: want %q, got %q", "Pending", got)
	}
	if got := parseTag(cg, "code"); got != "PND" {
		t.Errorf("code: want %q, got %q", "PND", got)
	}
}

// ── fileNameFromType ──────────────────────────────────────────────

func TestFileNameFromType(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"aggregateTypeVal", "aggregate_type"},
		{"UserStatusType", "user_status_type"},
		{"Simple", "simple"},
		{"AVal", "a"},
		{"CostMethodVal", "cost_method"},
		{"LedgerAccountTypeVal", "ledger_account_type"},
	}
	for _, c := range cases {
		got := fileNameFromType(c.input)
		if got != c.want {
			t.Errorf("fileNameFromType(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

// ── mapToSortedSlice ──────────────────────────────────────────────

func TestMapToSortedSlice_Empty(t *testing.T) {
	result := mapToSortedSlice(map[string]*EnumDef{})
	if len(result) != 0 {
		t.Fatalf("want empty slice, got %v", result)
	}
}

func TestMapToSortedSlice_Single(t *testing.T) {
	m := map[string]*EnumDef{
		"Z": {TypeName: "Z"},
	}
	result := mapToSortedSlice(m)
	if len(result) != 1 || result[0].TypeName != "Z" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestMapToSortedSlice_SortedAlphabetically(t *testing.T) {
	m := map[string]*EnumDef{
		"Charlie": {TypeName: "Charlie"},
		"Alpha":   {TypeName: "Alpha"},
		"Beta":    {TypeName: "Beta"},
	}
	result := mapToSortedSlice(m)
	want := []string{"Alpha", "Beta", "Charlie"}
	for i, def := range result {
		if def.TypeName != want[i] {
			t.Errorf("index %d: want %q, got %q", i, want[i], def.TypeName)
		}
	}
}

// ── validate ─────────────────────────────────────────────────────

func TestValidate_PanicsOnMissingAlias(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing alias")
		}
	}()
	validate(map[string]*EnumDef{
		"Foo": {TypeName: "Foo", Alias: "", Consts: map[string]EnumConst{"A": {}}},
	})
}

func TestValidate_PanicsOnEmptyConsts(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty consts")
		}
	}()
	validate(map[string]*EnumDef{
		"Foo": {TypeName: "Foo", Alias: "FooAlias", Consts: map[string]EnumConst{}},
	})
}

func TestValidate_NoPanicWhenValid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	validate(map[string]*EnumDef{
		"Foo": {TypeName: "Foo", Alias: "FooAlias", Consts: map[string]EnumConst{"A": {Name: "A"}}},
	})
}

// ── buildRegistryData ─────────────────────────────────────────────

func TestBuildRegistryData_SamePackage_NoImport(t *testing.T) {
	pkg := &packages.Package{Name: "enums", PkgPath: "mymod/internal/enums"}
	defs := []*EnumDef{
		{TypeName: "StatusVal", Alias: "Status", PkgName: "enums", PkgPath: "mymod/internal/enums"},
	}

	data := buildRegistryData(pkg, defs)

	if data.Package != "enums" {
		t.Errorf("Package: want %q, got %q", "enums", data.Package)
	}
	if len(data.Imports) != 0 {
		t.Errorf("Imports: want none, got %v", data.Imports)
	}
	if len(data.Enums) != 1 {
		t.Fatalf("Enums: want 1, got %d", len(data.Enums))
	}
	if data.Enums[0].RegisterFuncName != "AllStatus" {
		t.Errorf("RegisterFuncName: want %q, got %q", "AllStatus", data.Enums[0].RegisterFuncName)
	}
}

func TestBuildRegistryData_ExternalPackage_AddsImport(t *testing.T) {
	pkg := &packages.Package{Name: "registry", PkgPath: "mymod/internal/registry"}
	defs := []*EnumDef{
		{TypeName: "StatusVal", Alias: "Status", PkgName: "enums", PkgPath: "mymod/internal/enums"},
	}

	data := buildRegistryData(pkg, defs)

	if len(data.Imports) != 1 {
		t.Fatalf("Imports: want 1, got %d", len(data.Imports))
	}
	if data.Imports[0].Path != "mymod/internal/enums" {
		t.Errorf("Import.Path: want %q, got %q", "mymod/internal/enums", data.Imports[0].Path)
	}
	if data.Enums[0].RegisterFuncName != "enums.AllStatus" {
		t.Errorf("RegisterFuncName: want %q, got %q", "enums.AllStatus", data.Enums[0].RegisterFuncName)
	}
}

func TestBuildRegistryData_ImportsAreSorted(t *testing.T) {
	pkg := &packages.Package{Name: "reg", PkgPath: "mymod/reg"}
	defs := []*EnumDef{
		{Alias: "Z", PkgName: "z", PkgPath: "mymod/z"},
		{Alias: "A", PkgName: "a", PkgPath: "mymod/a"},
	}

	data := buildRegistryData(pkg, defs)

	if len(data.Imports) < 2 {
		t.Fatalf("want 2 imports, got %d", len(data.Imports))
	}
	if data.Imports[0].Path >= data.Imports[1].Path {
		t.Errorf("imports not sorted: %q >= %q", data.Imports[0].Path, data.Imports[1].Path)
	}
}

// ── renderTemplate ────────────────────────────────────────────────

func TestRenderTemplate_EnumsTmpl(t *testing.T) {
	def := &EnumDef{
		TypeName: "colorVal",
		Alias:    "Color",
		BaseType: "string",
		PkgName:  "enums",
		Consts: map[string]EnumConst{
			"Red":  {Name: "Red", Label: "Red", Code: "RED"},
			"Blue": {Name: "Blue", Label: "Blue", Code: "BLU"},
		},
	}

	out, err := renderTemplate("templates/enums.tmpl", def)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}

	body := string(out)
	for _, want := range []string{
		"func (v colorVal ) Enum()",
		"func AllColor()",
		"func ColorLabels()",
		"func ParseColor(",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, body)
		}
	}
}

func TestRenderTemplate_RegistryTmpl(t *testing.T) {
	data := RegistryTemplateData{
		Package: "enums",
		Imports: []Import{
			{Alias: "domain", Path: "mymod/internal/domain"},
		},
		Enums: []EnumRender{
			{RegisterFuncName: "AllStatus"},
			{RegisterFuncName: "domain.AllKind"},
		},
	}

	out, err := renderTemplate("templates/registry.tmpl", data)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}

	body := string(out)
	for _, want := range []string{
		"package enums",
		`"mymod/internal/domain"`,
		"enumx.Register(AllStatus())",
		"enumx.Register(domain.AllKind())",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("output missing %q\n--- output ---\n%s", want, body)
		}
	}
}

// ── writeFile ─────────────────────────────────────────────────────

func TestWriteFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.go")
	content := []byte("package test\n")

	if err := writeFile(path, content); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch:\n got  %q\n want %q", got, content)
	}
}

func TestWriteFile_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.go")
	_ = os.WriteFile(path, []byte("old content"), 0666)

	newContent := []byte("new content")
	if err := writeFile(path, newContent); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != string(newContent) {
		t.Errorf("want %q, got %q", newContent, got)
	}
}

// ── helpers ───────────────────────────────────────────────────────

func commentGroup(texts ...string) *ast.CommentGroup {
	var list []*ast.Comment
	for _, text := range texts {
		list = append(list, &ast.Comment{Slash: token.NoPos, Text: text})
	}
	return &ast.CommentGroup{List: list}
}