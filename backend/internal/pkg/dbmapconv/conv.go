package dbmapconv

import "time"

// Ptr returns a pointer to a copy of v.
func Ptr[T any](v T) *T { return &v }

// Deref dereferences v, returning the zero value if v is nil.
func Deref[T any](v *T) (zero T) {
	if v != nil {
		return *v
	}
	return zero
}

// TimeToStr formats t as RFC3339.
func TimeToStr(t time.Time) string {
	return t.Format(time.RFC3339)
}

// StrToTime parses s as RFC3339, returning zero time on error.
func StrToTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// PtrTimeToStr converts *time.Time to *string (nil-safe).
func PtrTimeToStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

// PtrStrToTime converts *string to *time.Time (nil-safe).
func PtrStrToTime(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

// IntToBool converts a SQLite INTEGER (0/1) to bool.
func IntToBool(v int64) bool { return v != 0 }

// BoolToInt converts bool to a SQLite INTEGER (0/1).
func BoolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}