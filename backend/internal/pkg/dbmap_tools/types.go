package dbmap_tools

// StructDef describes a model/db struct annotated with one or more //dbmap:sqlcdb= lines.
type StructDef struct {
	TypeName    string     // Go type name, e.g. "Account"
	SqlcTargets []string   // one target per //dbmap:sqlcdb= line, e.g. ["Account","CreateAccountParams","UpdateAccountParams"]
	PkgName     string     // source package name, e.g. "projection"
	PkgPath     string     // source package import path
	Dir         string     // filesystem directory of source package
	Fields      []FieldDef // ordered list of source fields
}

// FieldDef describes one field in a struct.
type FieldDef struct {
	GoName string // Go field name, e.g. "AccountId"
	GoType string // Go type string, e.g. "*string", "time.Time"
	DbTag  string // db struct-tag value, e.g. "account_id"
}

// FieldMapping pairs a source field with its matching sqlcdb field.
type FieldMapping struct {
	SrcField string // source Go field name
	TgtField string // target sqlcdb Go field name
	ToExpr   string // expression producing the sqlcdb value (source var is "m")
	FromExpr string // expression producing the model value (target var is "s")
}

// MapperData holds template data for one struct conversion pair.
type MapperData struct {
	SrcTypeName string
	TgtTypeName string
	Mappings    []FieldMapping
}

// FileTemplateData is the top-level data fed to mapper.tmpl for one output file.
// Generated code lives in the same package as the model, so no source import is needed.
type FileTemplateData struct {
	PkgName   string       // package declaration name, e.g. "projection"
	NeedsConv bool         // true when dbmapconv helpers are referenced
	Mappers   []MapperData
}