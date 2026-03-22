package cache

type CachePolicy int

const (
	ReadOnly CachePolicy = iota
	WriteBack
)
