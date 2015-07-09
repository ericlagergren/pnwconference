package auth

import (
	"errors"
	"fmt"
)

var (
	ErrBadUsername      = errors.New("Incorrect username.")
	ErrBadPassword      = errors.New("Incorrect password.")
	ErrInvalidTFACode   = errors.New("Invalid Two-Factor Authentication code.")
	ErrInvalidCSRFToken = errors.New("Invalid CSRF token.")

	// Generic error for when a user isn't logged in.
	ErrUnauthorized = errors.New("Unauthorized access.")

	// Means email/username was too long.
	ErrInvalidLogin = errors.New("Invalid login submitted.")

	ErrTooManyRequests = errors.New("Too many invalid login attempts.")
	ErrUnableToLogOut  = errors.New("Unable to destroy tokens.")

	ErrPassDoesntMatch = errors.New("Provided passwords do not match.")
	ErrCommonPassword  = errors.New("Password is too common.")
	ErrPassTooShort    = errors.New("Password is too short.")

	ErrNoReferer = errors.New("Referer header is empty.")
)

// httpError encapsulates an error and an HTTP status code.
// (E.g, 200 OK, 302 Found)
type HTTPError struct {
	Status int
	Err    error
}

func (err *HTTPError) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return fmt.Sprintf("Status %d", err.Status)
}
