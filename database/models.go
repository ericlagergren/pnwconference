package database

import (
	"encoding/gob"
	"time"
)

func init() {
	// Register the types we'll need when we add them to our auth
	// cookies.
	gob.RegisterName("main.User", &User{})
}

type User struct {
	Email           string    // User email address
	Name            string    // Username
	School          string    // Organization
	Password        []byte    // Bcrypt hashed pasword
	PasswordChanged time.Time // Last time password was changed

	// User's misc. data stored as XML. []byte instead of string because of
	// this issue:
	// https://github.com/go-sql-driver/mysql/wiki/Examples#ignoring-null-values
	Data []byte
}

type FailedAttempt struct {
	IP       string    // Address where attempt originated
	User     string    // Account of attempt
	Attempts int       // Number of attempts
	Last     time.Time // Time of last attempt
}
