package virtual

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	statusCodeToErrorWrapper = map[int]func(err error, serverID string) error{
		410: NewBlacklistedActivationError,
	}

	// Make sure it implements interface.
	_ HTTPError = NewBlacklistedActivationError(errors.New("n/a"), "n/a").(HTTPError)
)

// HTTPError is the interface implemented by errors that map to a specific
// status code. It should be used in conjunction with statusCodeToErrorWrapper
// so that the status code is automatically set on the server, and the status
// code is automatically translated back into the appropriate error wrapped by
// the client.
type HTTPError interface {
	HTTPStatusCode() int
}

// BlacklistedActivationErr indicates that the actor activation has been
// blacklisted on this specific server temporarily (usually due to resource
// usage or balancing reasons).
type BlacklistedActivationErr struct {
	err      error
	serverID string
}

// NewBlacklistedActivationError creates a new BlacklistedActivationErr.
func NewBlacklistedActivationError(err error, serverID string) error {
	return BlacklistedActivationErr{err: err, serverID: serverID}
}

func (b BlacklistedActivationErr) Error() string {
	return fmt.Sprintf(
		"BlacklistedActivationError(ServerID:%s): %s",
		b.serverID, b.err.Error())
}

func (b BlacklistedActivationErr) Is(target error) bool {
	if target == nil {
		return false
	}

	_, ok1 := target.(*BlacklistedActivationErr)
	_, ok2 := target.(BlacklistedActivationErr)
	return ok1 || ok2
}

func (b BlacklistedActivationErr) HTTPStatusCode() int {
	return http.StatusGone
}

func (b BlacklistedActivationErr) ServerID() string {
	return b.serverID
}

// IsBlacklistedActivationError returns a boolean indicating whether the error
// was caused by the actor being blacklisted from being activated on the server.
func IsBlacklistedActivationError(err error) bool {
	return errors.Is(err, BlacklistedActivationErr{})
}
