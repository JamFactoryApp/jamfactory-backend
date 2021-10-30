package types

type Permission string

const (
	RightsGuest Permission = "Guest"
	RightHost   Permission = "Host"
)

var (
	ValidPermissions = []Permission{RightsGuest, RightHost}
)
