package event

import "github.com/google/uuid"

func NewUUIDv7() string {
	return uuid.Must(uuid.NewV7()).String()
}
