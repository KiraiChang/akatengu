package cache

import (
	"golang.org/x/sync/singleflight"
)

type FlightGroup struct {
	group singleflight.Group
}

func (f *FlightGroup) Do(
	key string,
	fn func() ([]byte, error),
) ([]byte, error) {
	v, err, _ := f.group.Do(key, func() (any, error) {
		return fn()
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}
