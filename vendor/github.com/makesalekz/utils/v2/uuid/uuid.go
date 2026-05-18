package uuid

import "github.com/google/uuid"

// NewFromActorID
// this uses actor id to generate uuid
// WARNING: It is not safe to use as it uses default google uuid library, which panics in case of internal errors
func NewFromActorID(actorId int64) uuid.UUID {
	uid := uuid.Must(uuid.NewV7())

	uid[7] = byte(actorId >> 56)
	uid[9] = byte(actorId >> 48)
	uid[10] = byte(actorId >> 40)
	uid[11] = byte(actorId >> 32)
	uid[12] = byte(actorId >> 24)
	uid[13] = byte(actorId >> 16)
	uid[14] = byte(actorId >> 8)
	uid[15] = byte(actorId)

	return uid
}
