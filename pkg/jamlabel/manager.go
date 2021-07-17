package jamlabel

import (
	"github.com/pkg/errors"
)

var (
	ErrJamLabelNotFound = errors.New("could not find JamLabel")
)

// Manager manages a collection of unique JamLabels
type Manager interface {
	// List returns all JamLabels managed by this Manager
	List() []string
	// Create creates a new unique JamLabel
	Create() string
	// Delete deletes a JamLabel from this Manager
	Delete(jamLabel string) error
}
