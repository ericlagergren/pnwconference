package auth

// For internal use inside MarshalJSON only.
import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/EricLagerg/pnwconference/helpers"

	"github.com/golang/glog"
)

// These need to match the names in the ServerSession struct.
var (
	auth   = []byte("auth_token")
	csrf   = []byte("csrf_token")
	email  = []byte("email")
	school = []byte("school")
	date   = []byte("expiry")

	ErrCantMarshal   = errors.New("Cannot marshal ServerSession.")
	ErrCantUnmarshal = errors.New("Cannot unmarshal byte slice.")
)

// MarshalJSON implements the Marshaler interface for expedient JSON marshalling.
// We implement it ourself, as ugly and as brittle as it is, because of the
// speedup we receive. As long as this is updated whenever the ServerSession struct
// is updated, we shouldn't have any issues.
//
// See https://play.golang.org/p/FWubrFGNx2 for the benchmark code.
// The results of the benchmark on my laptop were:
//
// eric@archbox ~/sermodigital/sermocrm/test $ go test marshal_test.go -bench=.
// testing: warning: no tests to run
// PASS
// BenchmarkMarshal	          300000	      4460 ns/op
// BenchmarkCustomMarshal	 2000000	       889 ns/op
// ok  	command-line-arguments	43.620s
func (s *ServerSession) MarshalJSON() (buf []byte, err error) {

	// TODO: Does this need to be removed?
	defer func() {
		_, ok := recover().(error)
		if ok {
			err = ErrCantMarshal
			glog.Errorln(err)
		}
		return
	}()

	buf = make([]byte,
		2+ // brackets
			18+ // quotes
			5+ // colons
			4+ // commas
			1+ // 0-indexed
			len(auth)+ // "auth_token"
			len(csrf)+ // "csrf_token"
			len(email)+ // "email"
			len(school)+ // "school"
			len(date)+ // "date"
			len(s.AuthToken)+len(s.CSRFToken)+len(s.Email)+len(s.School)+helpers.NumWidth(s.Date))

	buf[0] = '{'
	buf[1] = '"'
	n := copy(buf[2:], auth) + 2 // Key
	buf[n] = '"'
	n++
	buf[n] = ':'
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], s.AuthToken) // Val
	buf[n] = '"'
	n++
	buf[n] = ','
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], csrf) // Key
	buf[n] = '"'
	n++
	buf[n] = ':'
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], s.CSRFToken) // Val
	buf[n] = '"'
	n++
	buf[n] = ','
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], email) // Key
	buf[n] = '"'
	n++
	buf[n] = ':'
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], s.Email) // Val
	buf[n] = '"'
	n++
	buf[n] = ','
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], school) // Key
	buf[n] = '"'
	n++
	buf[n] = ':'
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], s.School) // Val
	buf[n] = '"'
	n++
	buf[n] = ','
	n++
	buf[n] = '"'
	n++
	n += copy(buf[n:], date) // Key
	buf[n] = '"'
	n++
	buf[n] = ':'
	n++
	n += copy(buf[n:], helpers.FormatInt(s.Date)) // Val
	buf[n] = '}'

	// Sometimes we'll end up with null bytes at the end of our slice.
	// I'm not sure where they come from, but they were causing some errors
	// when Go would validate the return of our MarshalJSON interface.
	if i := bytes.IndexByte(buf, '\x00'); i != -1 {
		buf = buf[:i]
	}

	return
}

// UnmarshalJSON implements the Marshaler interface for expedient JSON
// unmarshalling of ServerSession structures. See ServerSession.MarshalJSON
// for more information on why it's done by hand.
//
// An additional note is that the standard library's json.Unmarshal
// base64-encodes byte slices, and we don't want to waste the time
// un-encoding those when we get the cached ServerSession from redis.
func (s *ServerSession) UnmarshalJSON(b []byte) (err error) {
	parts := bytes.Split(b, []byte(","))

	if len(parts[0]) == 0 {
		return ErrCantUnmarshal
	}

	s.AuthToken = parts[0][2+len(auth)+3 : len(parts[0])-1]
	s.CSRFToken = parts[1][1+len(csrf)+3 : len(parts[1])-1]
	s.Email = string(parts[2][1+len(email)+3 : len(parts[2])-1])
	s.School = string(parts[3][1+len(school)+3 : len(parts[3])-1])
	s.Date, err = strconv.ParseInt(
		string(parts[4][1+len(date)+2:len(parts[4])-1]), 10, 64)

	return err
}

// Ensure that our ServerSession implements the Marshaler and Unmarshaler
// interfaces.
var (
	_ json.Marshaler   = (*ServerSession)(nil)
	_ json.Unmarshaler = (*ServerSession)(nil)
)
