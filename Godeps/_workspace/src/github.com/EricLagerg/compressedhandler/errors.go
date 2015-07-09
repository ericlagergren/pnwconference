package compressedhandler

import (
	"errors"
	"fmt"
)

// ErrEmptyContentCoding indicates that the Accept-Encoding header was
// empty.
var ErrEmptyContentCoding = errors.New("Empty Accept-Encoding")

// KeyError is a Key/Value struct that matches the error with the
// string that caused the error.
type KeyError struct {
	Key string // The "key", which is the ill-formatted string
	Err error  // the "value", which is the error from strconv
}

func (k *KeyError) Error() string {
	return fmt.Sprintf("'%s' caused error: '%v'", k.Key, k.Err.Error())
}

// ErrorList is a slice of KeyErrors that allows individual errors
// to be extracted.
type ErrorList []KeyError // Slice of returned errors

func (e *ErrorList) Error() string {
	return fmt.Sprintf("%d errors", len(*e))
}
