package types

type MemberRights string

const (
	RightsGuest MemberRights = "Guest"
	RightHost   MemberRights = "Host"
)

var (
	ValidRights = []MemberRights{RightsGuest, RightHost}
)
