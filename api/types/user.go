package types

type UserType string

const (
	UserTypeHost  UserType = "Host"
	UserTypeGuest UserType = "Guest"
	UserTypeNew   UserType = "New"
)
