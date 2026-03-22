package enumx

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// EnumVal 是每個具名型別 embed 的核心結構
// T 是具名字串型別本身，例如 JournalStatus
type EnumVal[T ~string] struct {
	value T
	valid []T
}

// New 建立一個 EnumVal，同時驗證初始值
func New[T ~string](s string, valid []T) (EnumVal[T], error) {
	for _, v := range valid {
		if string(v) == s {
			return EnumVal[T]{value: v, valid: valid}, nil
		}
	}
	return EnumVal[T]{}, fmt.Errorf("invalid value %q, allowed: [%s]",
		s, joinEnum(valid))
}

// Must 同 New，但無效值直接 panic（適合測試或常數初始化）
func Must[T ~string](s string, valid []T) EnumVal[T] {
	v, err := New(s, valid)
	if err != nil {
		panic(err)
	}
	return v
}

func (e EnumVal[T]) Val() T         { return e.value }
func (e EnumVal[T]) String() string { return string(e.value) }

// UnmarshalJSON 實作 json.Unmarshaler
func (e *EnumVal[T]) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 補上和 Scan 一樣的 registry 查詢
	if len(e.valid) == 0 {
		var zero T
		key := fmt.Sprintf("%T", zero)
		for _, v := range registry[key] {
			e.valid = append(e.valid, T(v))
		}
	}

	parsed, err := New(s, e.valid)
	if err != nil {
		return err
	}
	e.value = parsed.value
	return nil
}

// MarshalJSON 實作 json.Marshaler
func (e EnumVal[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(e.value))
}

func joinEnum[T ~string](vals []T) string {
	ss := make([]string, len(vals))
	for i, v := range vals {
		ss[i] = string(v)
	}
	return strings.Join(ss, ", ")
}

// enumx.go 補充：Registry 讓零值也能驗證

var registry = map[string][]string{}

func Register[T ~string](valid []T) {
	var zero T
	key := fmt.Sprintf("%T", zero)
	for _, v := range valid {
		registry[key] = append(registry[key], string(v))
	}
}

// Scan 實作 sql.Scanner
func (e *EnumVal[T]) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("EnumVal.Scan: expected string, got %T", src)
	}
	// 如果 valid 清單是空的，從 registry 補回來
	if len(e.valid) == 0 {
		var zero T
		key := fmt.Sprintf("%T", zero)
		for _, v := range registry[key] {
			e.valid = append(e.valid, T(v))
		}
	}
	parsed, err := New(s, e.valid)
	if err != nil {
		return err
	}
	e.value = parsed.value
	return nil
}

// Value 實作 driver.Valuer，讓 sqlx 知道怎麼寫入 DB
func (e EnumVal[T]) Value() (driver.Value, error) {
	if e.value == "" {
		return nil, nil
	}
	return string(e.value), nil
}
