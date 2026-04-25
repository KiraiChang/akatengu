package enumx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	registry = map[reflect.Type]map[string]struct{}{}
	mu       sync.RWMutex

	// 可選：若你保證只在 init 註冊，可以設 false 提升效能
	threadSafe = true
)

func getType[T any]() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

func lock() {
	if threadSafe {
		mu.Lock()
	}
}
func unlock() {
	if threadSafe {
		mu.Unlock()
	}
}
func rlock() {
	if threadSafe {
		mu.RLock()
	}
}
func runlock() {
	if threadSafe {
		mu.RUnlock()
	}
}

// Register 註冊
func Register[T ~string](valid []T) {
	if len(valid) == 0 {
		panic("enumx: Register with empty valid set")
	}

	t := getType[T]()

	lock()
	defer unlock()

	if _, exists := registry[t]; exists {
		panic("enumx: duplicate register for type " + t.String())
	}

	m := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		m[string(v)] = struct{}{}
	}

	registry[t] = m
}

func mustRegistered[T ~string]() {
	t := getType[T]()

	rlock()
	defer runlock()

	if _, ok := registry[t]; !ok {
		panic("enumx: type not registered: " + t.String())
	}
}

func parse[T ~string](s string) (T, error) {
	t := getType[T]()

	rlock()
	m, ok := registry[t]
	runlock()

	if !ok {
		return "", fmt.Errorf("enumx: type not registered: %s", t.String())
	}

	if _, exists := m[s]; !exists {
		return "", fmt.Errorf("enumx: invalid value %q for %s", s, t.String())
	}

	return T(s), nil
}

type Enum[T ~string] struct {
	value T
}

func Parse[T ~string](s string) (Enum[T], error) {
	return New[T](T(s))
}

func New[T ~string](s T) (Enum[T], error) {
	v, err := parse[T](string(s))
	if err != nil {
		return Enum[T]{}, err
	}
	return Enum[T]{value: v}, nil
}

func Must[T ~string](s T) Enum[T] {
	v, err := New[T](s)
	if err != nil {
		panic(err)
	}
	return v
}

func (e Enum[T]) Val() T {
	return e.value
}

func (e Enum[T]) String() string {
	return string(e.value)
}

func (e Enum[T]) IsZero() bool {
	return e.value == ""
}

func (e Enum[T]) Is(v T) bool {
	return e.value == v
}

func (e Enum[T]) In(vals ...T) bool {
	for _, v := range vals {
		if e.value == v {
			return true
		}
	}
	return false
}

// -------------------------
// JSON
// -------------------------

func (e Enum[T]) MarshalJSON() ([]byte, error) {
	if e.value == "" {
		return []byte("null"), nil
	}
	return json.Marshal(string(e.value))
}

func (e *Enum[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.value = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	v, err := parse[T](s)
	if err != nil {
		return err
	}

	e.value = v
	return nil
}

// -------------------------
// SQL
// -------------------------

func (e Enum[T]) Value() (driver.Value, error) {
	if e.value == "" {
		return nil, nil
	}
	return string(e.value), nil
}

func (e *Enum[T]) Scan(src any) error {
	if src == nil {
		e.value = ""
		return nil
	}

	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("enumx: cannot scan %T", src)
	}

	val, err := parse[T](s)
	if err != nil {
		return err
	}

	e.value = val
	return nil
}

// -------------------------
// Nullable（重要：DB + JSON 常用）
// -------------------------

type NullEnum[T ~string] struct {
	Enum  Enum[T]
	Valid bool
}

func (n NullEnum[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return n.Enum.MarshalJSON()
}

func (n *NullEnum[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Enum = Enum[T]{}
		return nil
	}
	n.Valid = true
	return n.Enum.UnmarshalJSON(data)
}

func (n NullEnum[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Enum.Value()
}

func (n *NullEnum[T]) Scan(src any) error {
	if src == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	return n.Enum.Scan(src)
}

// -------------------------
// Helper（可選）
// -------------------------

func IsValid[T ~string](s string) bool {
	_, err := parse[T](s)
	return err == nil
}

func MustRegistered[T ~string]() {
	mustRegistered[T]()
}

func From[T ~string](v T) Enum[T] {
	return Must[T](v)
}

// -------------------------
// Domain Error（可選）
// -------------------------

var ErrInvalidEnum = errors.New("invalid enum value")
