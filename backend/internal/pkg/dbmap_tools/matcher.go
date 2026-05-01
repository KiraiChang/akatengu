package dbmap_tools

import "fmt"

// buildMappings pairs source fields with sqlcdb fields matched by db tag, and
// determines a Go expression for each conversion direction.
// Returns the mappings and whether dbmapconv helpers are referenced.
func buildMappings(srcFields []FieldDef, tgtByTag map[string]FieldDef) ([]FieldMapping, bool) {
	needsConv := false
	var mappings []FieldMapping

	for _, sf := range srcFields {
		tf, ok := tgtByTag[sf.DbTag]
		if !ok {
			continue // no matching field in sqlcdb
		}

		toExpr, fromExpr, conv := resolveConversion(sf, tf)
		if conv {
			needsConv = true
		}

		mappings = append(mappings, FieldMapping{
			SrcField: sf.GoName,
			TgtField: tf.GoName,
			ToExpr:   toExpr,
			FromExpr: fromExpr,
		})
	}

	return mappings, needsConv
}

// resolveConversion returns (toExpr, fromExpr, needsConvPkg).
// toExpr converts "m.SrcField" to the target type.
// fromExpr converts "s.TgtField" to the source type.
func resolveConversion(src, tgt FieldDef) (toExpr, fromExpr string, needsConvPkg bool) {
	s := src.GoType
	t := tgt.GoType
	mf := "m." + src.GoName
	sf := "s." + tgt.GoName

	switch {
	case s == t:
		return mf, sf, false

	// T → *T
	case "*"+s == t:
		return fmt.Sprintf("dbmapconv.Ptr(%s)", mf), fmt.Sprintf("dbmapconv.Deref(%s)", sf), true

	// *T → T
	case s == "*"+t:
		return fmt.Sprintf("dbmapconv.Deref(%s)", mf), fmt.Sprintf("dbmapconv.Ptr(%s)", sf), true

	// time.Time ↔ string
	case s == "time.Time" && t == "string":
		return fmt.Sprintf("dbmapconv.TimeToStr(%s)", mf), fmt.Sprintf("dbmapconv.StrToTime(%s)", sf), true
	case s == "string" && t == "time.Time":
		return fmt.Sprintf("dbmapconv.StrToTime(%s)", mf), fmt.Sprintf("dbmapconv.TimeToStr(%s)", sf), true

	// *time.Time ↔ *string
	case s == "*time.Time" && t == "*string":
		return fmt.Sprintf("dbmapconv.PtrTimeToStr(%s)", mf), fmt.Sprintf("dbmapconv.PtrStrToTime(%s)", sf), true
	case s == "*string" && t == "*time.Time":
		return fmt.Sprintf("dbmapconv.PtrStrToTime(%s)", mf), fmt.Sprintf("dbmapconv.PtrTimeToStr(%s)", sf), true

	// bool ↔ int64 (SQLite stores booleans as INTEGER)
	case s == "bool" && t == "int64":
		return fmt.Sprintf("dbmapconv.BoolToInt(%s)", mf), fmt.Sprintf("dbmapconv.IntToBool(%s)", sf), true
	case s == "int64" && t == "bool":
		return fmt.Sprintf("dbmapconv.IntToBool(%s)", mf), fmt.Sprintf("dbmapconv.BoolToInt(%s)", sf), true

	default:
		// Types differ in a way we don't auto-handle; emit a TODO so the file
		// still compiles only when the user fills in the conversion.
		return fmt.Sprintf("/* TODO: convert %s→%s */ %s", s, t, mf),
			fmt.Sprintf("/* TODO: convert %s→%s */ %s", t, s, sf),
			false
	}
}