package uuidx

import "github.com/google/uuid"

var eventNamespace = uuid.MustParse("7e0f4be0-1a2b-7000-8000-a0a000beef01")

// NewFromEvent generates a deterministic UUID (v5 / SHA1) from an event UUID and a qualifier string.
// This is used to create stable sub-entity UUIDs that survive projection replay:
// given the same eventUUID and qualifier, the output is always identical.
func NewFromEvent(eventUUID string, qualifier string) string {
	return uuid.NewSHA1(eventNamespace, []byte(eventUUID+":"+qualifier)).String()
}
