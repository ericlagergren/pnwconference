package auth

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// validLoginInput is a quick check to see if a user gave us valid data.
// It will return two booleans; the first is true if the string is empty,
// and the second is true if the ID is an acceptable length.
func validLoginInput(id string) (bool, bool) {
	// TODO: Should this return errors instead of two bools?
	n := utf8.RuneCountInString(id)
	return n != 0, n > 0 && n <= 255
}

// CheckValidPassword will check to see if the password meets the minimum
// password requirements. It'll return an error if applicable. Nil implies
// a valid password.
func CheckValidPassword(p1, p2 string) error {
	// Passwords have to match.
	if p1 != p2 {
		return ErrPassDoesntMatch
	}

	// Is it long enough?
	if n := utf8.RuneCountInString(p1); n < minPasswordLength {
		return ErrPassTooShort
	}

	// Is it a common password?
	lower := strings.ToLower(p1)
	if _, ok := passwordMap[lower]; ok {
		return ErrCommonPassword
	}

	return nil
}

// sameOrigin checks that the origins are the same.
func sameOrigin(o1, o2 *url.URL) bool {
	return o1.Scheme == o2.Scheme && o1.Host == o2.Host
}

// clearSlice will overwrite an entire slice with null bytes.
func clearSlice(buf []byte) {
	for i := range buf {
		buf[i] = '0'
	}
}
