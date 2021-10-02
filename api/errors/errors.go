package errors

import (
	"github.com/pkg/errors"
)

var (
	ErrJamLabelMalformed     = errors.New("malformed JamLabel")
	ErrJamLabelMissing       = errors.New("missing JamLabel")
	ErrJamSessionNotFound    = errors.New("no JamSession found")
	ErrSearchResultMalformed = errors.New("malformed search result")
	ErrSearchTypeInvalid     = errors.New("invalid search type")
	ErrSessionCouldNotSave   = errors.New("could not save session")
	ErrSessionMalformed      = errors.New("malformed session")
	ErrSessionMissing        = errors.New("missing session")
	ErrTokenInvalid          = errors.New("invalid token")
	ErrTokenMalformed        = errors.New("malformed token")
	ErrTokenMissing          = errors.New("missing token")
	ErrTokenMismatch         = errors.New("state mismatch")
	ErrUserTypeInvalid       = errors.New("invalid user type")
	ErrUserTypeMalformed     = errors.New("malformed user type")
	ErrUserTypeMissing       = errors.New("missing user type")
	ErrNoUser                = errors.New("no user for request")
	ErrInvalidVotingType     = errors.New("invalid voting type")
	ErrOriginMissing         = errors.New("missing origin")
	ErrOriginMalformed       = errors.New("malformed origin")
	ErrIdentifierMissing     = errors.New("missing identifier")
	ErrIdentifierMalformed   = errors.New("malformed identifier")
	ErrAlreadyMember         = errors.New("already member")
	ErrQueueEmpty            = errors.New("queue empty")
	ErrNoDevice              = errors.New("no playback device")
	ErrOnlyOneHost           = errors.New("only one host allowed")
	ErrBadRight              = errors.New("bad right")
)
