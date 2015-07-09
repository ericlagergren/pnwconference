package helpers

import (
	"github.com/gorilla/sessions"
)

const (
	digits   = "0123456789abcdefghijklmnopqrstuvwxyz"
	digits01 = "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	digits10 = "0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"
)

func FormatInt(n int64) []byte {
	u := uint64(n)
	var a [64 + 1]byte // +1 for sign of 64bit value in base 2
	i := len(a)

	neg := false
	if n < 0 {
		neg = true
		u = -u
	}

	// common case: use constants for / and % because
	// the compiler can optimize it into a multiply+shift,
	// and unroll loop
	for u >= 100 {
		i -= 2
		q := u / 100
		j := uintptr(u - q*100)
		a[i+1] = digits01[j]
		a[i+0] = digits10[j]
		u = q
	}
	if u >= 10 {
		i--
		q := u / 10
		a[i] = digits[uintptr(u-q*10)]
		u = q
	}

	// u < base
	i--
	a[i] = digits[uintptr(u)]

	if neg {
		i--
		a[i] = '-'
	}

	return a[i:]
}

func NumWidth(n int64) int {
	width := 0
	minWidth := 1

	for ; 10 < n; n /= 10 {
		width++
	}

	if width < minWidth {
		width = minWidth
	}

	return int(width)
}

func GetUsername(session *sessions.Session) (string, bool) {
	nameIf := session.Values["user"]
	name, ok := nameIf.(string)
	return name, ok
}
