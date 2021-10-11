package errors

import (
	"github.com/pkg/errors"
)

var (
	ErrJamSessionNotFound    = errors.New("no JamSession found")
	ErrSearchResultMalformed = errors.New("malformed search result")
	ErrSearchTypeInvalid     = errors.New("invalid search type")
	ErrSessionCouldNotSave   = errors.New("could not save session")
	ErrSessionMalformed      = errors.New("malformed session")
	ErrSessionMissing        = errors.New("missing session")
	ErrTokenInvalid          = errors.New("invalid token")
	ErrTokenMismatch         = errors.New("state mismatch")
	ErrUserTypeInvalid       = errors.New("invalid user type")
	ErrOriginMissing         = errors.New("missing origin")
	ErrOriginMalformed       = errors.New("malformed origin")
	ErrIdentifierMissing     = errors.New("missing identifier")
	ErrIdentifierMalformed   = errors.New("malformed identifier")
	ErrAlreadyMember         = errors.New("already member")
	ErrQueueEmpty            = errors.New("queue empty")
	ErrNoDevice              = errors.New("no playback device")
	ErrOnlyOneHost           = errors.New("only one host allowed")
	ErrBadRight              = errors.New("bad right")
	ErrWrongMemberCount = errors.New("wrong member count")
	ErrMissingMember = errors.New("member missing")
)
