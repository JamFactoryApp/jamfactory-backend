package types

type SessionType string

const (
	SessionTypeHost  SessionType = "Host"
	SessionTypeGuest SessionType = "Guest"
	SessionTypeNew    SessionType = "New"
)
