package kvdb

type TTLState int

const (
	TTLKeyNotFound TTLState = iota + 1
	TTLPersistent
	TTLExpiring
)
