package permissions

type Permission string

type Permissions []Permission

const (
	Guest  Permission = "Guest"
	Host              = "Host"
	Listen            = "Listen"
)

var valid = map[Permission]struct{}{
	Guest:  {},
	Host:   {},
	Listen: {},
}

func (p Permission) Valid() bool {
	_, ok := valid[p]
	return ok
}

func (p Permissions) Valid() bool {
	for _, toCheck := range p {
		if !toCheck.Valid() {
			return false
		}
	}
	return true
}
