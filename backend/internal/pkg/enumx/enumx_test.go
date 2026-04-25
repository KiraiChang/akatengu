package enumx

import (
	"encoding/json"
	"os"
	"testing"
)

// ── 測試用型別 ────────────────────────────────────────────────────
// 每個型別只在 TestMain 裡 Register 一次，避免重複 panic。

type colorVal string

const (
	Red   colorVal = "red"
	Blue  colorVal = "blue"
	Green colorVal = "green"
)

type statusVal string

const (
	Active   statusVal = "active"
	Inactive statusVal = "inactive"
)

// 永遠不 Register，用於測試「未註冊」路徑
type neverRegisteredVal string

// 只在 TestRegister_DuplicateType_Panics 裡用到
type dupVal string

func TestMain(m *testing.M) {
	Register([]colorVal{Red, Blue, Green})
	Register([]statusVal{Active, Inactive})
	os.Exit(m.Run())
}

// ── Register ──────────────────────────────────────────────────────

func TestRegister_EmptySlice_Panics(t *testing.T) {
	defer expectPanic(t, "Register with empty slice")
	Register([]neverRegisteredVal{})
}

func TestRegister_DuplicateType_Panics(t *testing.T) {
	Register([]dupVal{"x"}) // 第一次成功
	defer expectPanic(t, "duplicate Register")
	Register([]dupVal{"x"}) // 第二次應 panic
}

// ── Parse / New ───────────────────────────────────────────────────

func TestParse_ValidValue(t *testing.T) {
	e, err := Parse[colorVal]("red")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Val() != Red {
		t.Errorf("Val: want %q, got %q", Red, e.Val())
	}
}

func TestParse_InvalidValue_ReturnsError(t *testing.T) {
	_, err := Parse[colorVal]("purple")
	if err == nil {
		t.Fatal("expected error for invalid value")
	}
}

func TestParse_UnregisteredType_ReturnsError(t *testing.T) {
	_, err := Parse[neverRegisteredVal]("anything")
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestNew_ValidValue(t *testing.T) {
	e, err := New[colorVal](Blue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Val() != Blue {
		t.Errorf("Val: want %q, got %q", Blue, e.Val())
	}
}

// ── Must ──────────────────────────────────────────────────────────

func TestMust_ValidValue_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	e := Must[colorVal](Green)
	if e.Val() != Green {
		t.Errorf("Val: want %q, got %q", Green, e.Val())
	}
}

func TestMust_InvalidValue_Panics(t *testing.T) {
	defer expectPanic(t, "Must with invalid value")
	Must[colorVal]("yellow")
}

// ── MustRegistered ────────────────────────────────────────────────

func TestMustRegistered_Registered_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	MustRegistered[colorVal]()
}

func TestMustRegistered_Unregistered_Panics(t *testing.T) {
	defer expectPanic(t, "MustRegistered for unregistered type")
	MustRegistered[neverRegisteredVal]()
}

// ── IsValid ───────────────────────────────────────────────────────

func TestIsValid_ValidValue_ReturnsTrue(t *testing.T) {
	if !IsValid[colorVal]("red") {
		t.Fatal("expected true for valid value")
	}
}

func TestIsValid_InvalidValue_ReturnsFalse(t *testing.T) {
	if IsValid[colorVal]("purple") {
		t.Fatal("expected false for invalid value")
	}
}

func TestIsValid_UnregisteredType_ReturnsFalse(t *testing.T) {
	if IsValid[neverRegisteredVal]("anything") {
		t.Fatal("expected false for unregistered type")
	}
}

// ── Enum methods ──────────────────────────────────────────────────

func TestEnum_Val(t *testing.T) {
	e := mustColor(t, Red)
	if e.Val() != Red {
		t.Errorf("want %q, got %q", Red, e.Val())
	}
}

func TestEnum_String(t *testing.T) {
	e := mustColor(t, Blue)
	if e.String() != "blue" {
		t.Errorf("want %q, got %q", "blue", e.String())
	}
}

func TestEnum_IsZero_ZeroValue(t *testing.T) {
	var e Enum[colorVal]
	if !e.IsZero() {
		t.Fatal("zero value should be zero")
	}
}

func TestEnum_IsZero_NonZero(t *testing.T) {
	e := mustColor(t, Red)
	if e.IsZero() {
		t.Fatal("non-zero value should not be zero")
	}
}

func TestEnum_Is_Match(t *testing.T) {
	e := mustColor(t, Red)
	if !e.Is(Red) {
		t.Fatal("Is(Red) should be true")
	}
}

func TestEnum_Is_NoMatch(t *testing.T) {
	e := mustColor(t, Red)
	if e.Is(Blue) {
		t.Fatal("Is(Blue) should be false for Red")
	}
}

func TestEnum_In_ContainsValue(t *testing.T) {
	e := mustColor(t, Green)
	if !e.In(Red, Blue, Green) {
		t.Fatal("In should return true when value is in list")
	}
}

func TestEnum_In_NotContainsValue(t *testing.T) {
	e := mustColor(t, Red)
	if e.In(Blue, Green) {
		t.Fatal("In should return false when value is not in list")
	}
}

// ── JSON ──────────────────────────────────────────────────────────

func TestEnum_MarshalJSON_NonEmpty(t *testing.T) {
	e := mustColor(t, Red)
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"red"` {
		t.Errorf("want %q, got %q", `"red"`, string(b))
	}
}

func TestEnum_MarshalJSON_ZeroValue_ReturnsNull(t *testing.T) {
	var e Enum[colorVal]
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != "null" {
		t.Errorf("want %q, got %q", "null", string(b))
	}
}

func TestEnum_UnmarshalJSON_ValidValue(t *testing.T) {
	var e Enum[colorVal]
	if err := json.Unmarshal([]byte(`"blue"`), &e); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if e.Val() != Blue {
		t.Errorf("want %q, got %q", Blue, e.Val())
	}
}

func TestEnum_UnmarshalJSON_Null_SetsZero(t *testing.T) {
	e := mustColor(t, Red)
	if err := json.Unmarshal([]byte("null"), &e); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !e.IsZero() {
		t.Fatal("after null unmarshal, value should be zero")
	}
}

func TestEnum_UnmarshalJSON_InvalidValue_ReturnsError(t *testing.T) {
	var e Enum[colorVal]
	if err := json.Unmarshal([]byte(`"purple"`), &e); err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}

func TestEnum_RoundTrip_JSON(t *testing.T) {
	original := mustColor(t, Green)
	b, _ := json.Marshal(original)

	var decoded Enum[colorVal]
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if decoded.Val() != original.Val() {
		t.Errorf("round-trip mismatch: want %q, got %q", original.Val(), decoded.Val())
	}
}

// ── SQL Value / Scan ──────────────────────────────────────────────

func TestEnum_Value_NonEmpty(t *testing.T) {
	e := mustColor(t, Red)
	v, err := e.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != "red" {
		t.Errorf("want %q, got %v", "red", v)
	}
}

func TestEnum_Value_ZeroValue_ReturnsNil(t *testing.T) {
	var e Enum[colorVal]
	v, err := e.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != nil {
		t.Errorf("want nil, got %v", v)
	}
}

func TestEnum_Scan_String(t *testing.T) {
	var e Enum[colorVal]
	if err := e.Scan("blue"); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if e.Val() != Blue {
		t.Errorf("want %q, got %q", Blue, e.Val())
	}
}

func TestEnum_Scan_Bytes(t *testing.T) {
	var e Enum[colorVal]
	if err := e.Scan([]byte("green")); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if e.Val() != Green {
		t.Errorf("want %q, got %q", Green, e.Val())
	}
}

func TestEnum_Scan_Nil_SetsZero(t *testing.T) {
	e := mustColor(t, Red)
	if err := e.Scan(nil); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !e.IsZero() {
		t.Fatal("after nil scan, value should be zero")
	}
}

func TestEnum_Scan_UnsupportedType_ReturnsError(t *testing.T) {
	var e Enum[colorVal]
	if err := e.Scan(42); err == nil {
		t.Fatal("expected error for unsupported scan type")
	}
}

func TestEnum_Scan_InvalidValue_ReturnsError(t *testing.T) {
	var e Enum[colorVal]
	if err := e.Scan("purple"); err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}

// ── NullEnum JSON ─────────────────────────────────────────────────

func TestNullEnum_MarshalJSON_Valid(t *testing.T) {
	n := NullEnum[colorVal]{Enum: mustColor(t, Red), Valid: true}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"red"` {
		t.Errorf("want %q, got %q", `"red"`, string(b))
	}
}

func TestNullEnum_MarshalJSON_Invalid_ReturnsNull(t *testing.T) {
	var n NullEnum[colorVal]
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != "null" {
		t.Errorf("want %q, got %q", "null", string(b))
	}
}

func TestNullEnum_UnmarshalJSON_ValidValue_SetsValid(t *testing.T) {
	var n NullEnum[colorVal]
	if err := json.Unmarshal([]byte(`"green"`), &n); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !n.Valid {
		t.Fatal("Valid should be true")
	}
	if n.Enum.Val() != Green {
		t.Errorf("want %q, got %q", Green, n.Enum.Val())
	}
}

func TestNullEnum_UnmarshalJSON_Null_ClearsValid(t *testing.T) {
	n := NullEnum[colorVal]{Enum: mustColor(t, Red), Valid: true}
	if err := json.Unmarshal([]byte("null"), &n); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if n.Valid {
		t.Fatal("Valid should be false after null")
	}
}

func TestNullEnum_RoundTrip_JSON(t *testing.T) {
	original := NullEnum[colorVal]{Enum: mustColor(t, Blue), Valid: true}
	b, _ := json.Marshal(original)

	var decoded NullEnum[colorVal]
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !decoded.Valid || decoded.Enum.Val() != original.Enum.Val() {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

// ── NullEnum SQL ──────────────────────────────────────────────────

func TestNullEnum_Value_Valid(t *testing.T) {
	n := NullEnum[colorVal]{Enum: mustColor(t, Red), Valid: true}
	v, err := n.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != "red" {
		t.Errorf("want %q, got %v", "red", v)
	}
}

func TestNullEnum_Value_Invalid_ReturnsNil(t *testing.T) {
	var n NullEnum[colorVal]
	v, err := n.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != nil {
		t.Errorf("want nil, got %v", v)
	}
}

func TestNullEnum_Scan_NonNil_SetsValid(t *testing.T) {
	var n NullEnum[colorVal]
	if err := n.Scan("blue"); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !n.Valid {
		t.Fatal("Valid should be true after non-nil scan")
	}
	if n.Enum.Val() != Blue {
		t.Errorf("want %q, got %q", Blue, n.Enum.Val())
	}
}

func TestNullEnum_Scan_Nil_ClearsValid(t *testing.T) {
	n := NullEnum[colorVal]{Enum: mustColor(t, Red), Valid: true}
	if err := n.Scan(nil); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if n.Valid {
		t.Fatal("Valid should be false after nil scan")
	}
}

// ── From ──────────────────────────────────────────────────────────

func TestFrom_ValidValue(t *testing.T) {
	e := From[colorVal](Red)
	if e.Val() != Red {
		t.Errorf("want %q, got %q", Red, e.Val())
	}
}

func TestFrom_InvalidValue_Panics(t *testing.T) {
	defer expectPanic(t, "From with invalid value")
	From[colorVal]("yellow")
}

// ── helpers ───────────────────────────────────────────────────────

func mustColor(t *testing.T, v colorVal) Enum[colorVal] {
	t.Helper()
	e, err := New[colorVal](v)
	if err != nil {
		t.Fatalf("New[colorVal](%q): %v", v, err)
	}
	return e
}

func expectPanic(t *testing.T, label string) {
	t.Helper()
	if r := recover(); r == nil {
		t.Fatalf("%s: expected panic but did not panic", label)
	}
}