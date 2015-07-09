package auth

import (
	"crypto/subtle"
	"mime"
	"net/http"
	"net/url"

	"github.com/EricLagerg/pnwconference/tokens"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

// getCSRF returns the CSRF from a session or creates a new CSRF token if
// the value does not exist inside the session.
func getCSRF(session *sessions.Session) (csrf []byte, ok bool) {
	csrfIf, ok := session.Values[csrfToken]
	if !ok {
		csrf = tokens.NewCSRFToken(session.ID)
	} else {
		if csrf, ok = csrfIf.([]byte); !ok {
			csrf = tokens.NewCSRFToken(session.ID)
		}
	}
	return
}

// ValidCSRF returns true if the given CSRF token matches the CSRF token
// for a specific session.
func ValidCSRF(r *http.Request, session *sessions.Session, ws bool) bool {
	var err error

	// MITM check.
	if r.URL.Scheme == "https" {

		referer, err := url.Parse(r.Header.Get("Referer"))
		if err != nil || referer.String() == "" {
			glog.Errorln(err, ErrNoReferer)
			return false
		}

		if !sameOrigin(referer, r.URL) {
			return false
		}
	}

	ss, err := GetSession(session.ID)
	if err != nil {
		return false
	}

	var mac1 []byte

	// If we have an input field inside a multipart form, then our CSRF
	// token will be found inside Form. Otherwise, it'll be in
	// PostForm. We'd like to default to PostFormValue because FormValue
	// accepts URL query parameters, and we'll never send CSRF tokens
	// that way.
	d, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if d == "multipart/form-data" || ws {
		mac1 = []byte(r.FormValue("_csrf"))
	} else {
		mac1 = []byte(r.PostFormValue("_csrf"))
	}

	mac2 := ss.CSRFToken

	return len(mac1) == len(mac2) && subtle.ConstantTimeCompare(mac1, mac2) == 1
}
