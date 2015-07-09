package auth

import (
	"bytes"
	"net/http"
	"time"

	dt "github.com/EricLagerg/pnwconference/controllers/datatypes"
	"github.com/EricLagerg/pnwconference/database"
	"github.com/EricLagerg/pnwconference/paths"
	"github.com/EricLagerg/pnwconference/tokens"

	"github.com/golang/glog"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

const (
	cookieName = "pnw_conf"

	authToken = "AUTHID"
	authDate  = "AUTHDA"
	csrfToken = "CSRFTK"

	minPasswordLength = 8

	nsConv    = 1000000000          // second -> nanosecond conversion
	allowance = (3600 * nsConv) * 8 // 8 hours
)

var (
	store = sessions.NewFilesystemStore(
		"store",
		[]byte("2CA434288F3FAB93963CBA1A5B836EEF"),
		[]byte("0655A28CAAEB0448132026D863771C5F"),
	)
)

func init() {
	// Secure cookie options.
	store.Options = &sessions.Options{
		// Domain:   "localhost", // Set this when we have a domain name.
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		// Secure:   true, // Set this when TLS is set up.
	}
}

// Enum for AuthenticateUser's return values.
type AuthStatus uint8

const (
	InvalidAuth AuthStatus = iota // Bad user/password.
	ValidAuth                     // Auth is 100% valid; proceed.
)

// AuthenticateUser attempts to authenticate a user from the login page.
//
// It returns:
//     - An an enum indicating the user's authentication status.
//     - A pointer to dt.LoginData which holds information to be displayed.
//     - An HTTPError if applicable.
//
// Errors displayed to the user _must_ use proper punctuation.
func AuthenticateUser(w http.ResponseWriter, r *http.Request) (AuthStatus, *dt.LoginData, *HTTPError) {
	// Gather information about the user.
	session, validAuth, httperr := CheckSession(r)
	if httperr != nil {
		glog.Errorln(httperr.Err)
		return InvalidAuth, &dt.LoginData{}, &HTTPError{
			http.StatusInternalServerError, httperr.Err,
		}
	}

	ss, err := GetSession(session.ID)
	if err != nil {
		glog.Errorln(err)
	}

	// Technically this should never be true because if the cookie is
	// valid the user won't have to re-authenticate.
	if validAuth {
		glog.V(2).Infof("redirected to %s\n", paths.DashboardPath)

		return ValidAuth, &dt.LoginData{
			CSRF:  ss.CSRFToken,
			Redir: paths.DashboardPath,
		}, nil
	}

	// Now we can begin to gather form information.
	if err := r.ParseForm(); err != nil {
		glog.Errorln(err)
		return InvalidAuth, &dt.LoginData{
			CSRF: ss.CSRFToken,
		}, &HTTPError{http.StatusInternalServerError, err}
	}

	if !ValidCSRF(r, session, false) {
		return InvalidAuth, &dt.LoginData{
			CSRF:  ss.CSRFToken,
			Error: "Invalid CSRF token.",
		}, &HTTPError{http.StatusOK, ErrInvalidCSRFToken}
	}

	// Max email address/username is 255 characters, so if it's too long bail.
	// There's no reason to _not_ have a hard limit for email addresses.
	formID := r.PostFormValue("username")
	length, ok := validLoginInput(formID)

	// Username with a length of 0.
	if !length {
		return InvalidAuth, &dt.LoginData{
			CSRF:  ss.CSRFToken,
			Error: "Enter your username.",
		}, &HTTPError{http.StatusOK, ErrInvalidLogin}
	}

	if !ok {
		return InvalidAuth, &dt.LoginData{
			CSRF:  ss.CSRFToken,
			Error: "Enter a valid username.",
		}, &HTTPError{http.StatusOK, ErrInvalidLogin}
	}

	// No need to validate size. POSTs larger than client_max_body_size will
	// be discarded, and there's absolutely no reason why we should _ever_
	// have a limit on password length.
	formPassword := r.PostFormValue("password")

	// Now that the form data isn't blatantly invalid, create a new user
	// object to test the form data given to us.
	user, exists, err := database.CheckUser(formID)
	if err != nil {
		glog.Errorln(err)
	}

	// User does not exist. Package the error and display an error message.
	if !exists {

		session.AddFlash(ErrBadUsername.Error(), "_errors")
		if err := session.Save(r, w); err != nil {
			glog.Errorln(err)
			return InvalidAuth, &dt.LoginData{CSRF: ss.CSRFToken}, &HTTPError{
				http.StatusInternalServerError, err,
			}
		}

		// NOTE: It's _not_ a security risk to let the user know the given
		// username does not exist. If there's a public sign up page, then
		// they already have the ability to check if a username exists,
		// thus any "coy" language like, "The username/password are invalid"
		// just creates a bad user experience.
		// See: http://blog.codinghorror.com/the-god-login/
		return InvalidAuth, &dt.LoginData{
			CSRF:  ss.CSRFToken,
			Error: "Username does not exist.",
		}, &HTTPError{http.StatusOK, ErrBadUsername}
	}

	// Securely compare hashes.
	err = ComparePassword(user, []byte(formPassword), true)

	if err != nil {
		glog.Errorln(err)

		session.AddFlash(ErrBadPassword.Error(), "_errors")
		if err := session.Save(r, w); err != nil {
			glog.Errorln(err)
			return InvalidAuth, &dt.LoginData{
					CSRF: ss.CSRFToken,
				}, &HTTPError{
					http.StatusInternalServerError, ErrBadPassword,
				}
		}

		// bcrypt.CompareHashAndPassword will return one of two named errors:
		//     - ErrMismatchedHashAndPasword if the password doesn't match
		//       the hash.
		//     - ErrHashTooShort if the hash is too short to be a bcrypt hash
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return InvalidAuth, &dt.LoginData{
				CSRF:  ss.CSRFToken,
				Error: "The password you entered is incorrect.",
			}, &HTTPError{http.StatusOK, ErrBadPassword}
		} else {
			glog.Errorln(err)
			return InvalidAuth, &dt.LoginData{
					CSRF: ss.CSRFToken}, &HTTPError{
					http.StatusInternalServerError, err,
				}
		}
	}

	// Default end date is now (ns) + our allowance, which is usually like
	// 8 hours or so.
	endDate := time.Now().UnixNano() + allowance

	// If the user selected the "remember me" checkbox, set the expiration
	// date to as high as it will go.
	remember := r.PostFormValue("remember")
	if remember == "true" {
		endDate = int64(^uint64(0) >> 1)
	}

	// Check if the CSRF/ID are set. It might not be set if the user didn't
	// have any cookies when they visited the login page.
	if session.ID == "" {
		session.ID = tokens.NewSessionID()
	}

	csrf, _ := getCSRF(session)

	ss.AuthToken = tokens.NewAuthToken()
	ss.CSRFToken = csrf
	ss.Email = user.Email
	ss.Date = endDate
	ss.School = user.School

	// Store the session and if it doesn't work try to remove it just
	// to be safe.
	//
	// (Note: this does *not* save the session in the user's
	// browser. While a more generic 'SaveSession' would be nice, it'd
	// also be mixing our auth and SQL 'modules' which is a bit sloppy.
	// I'd rather have to explicitly perform each step rather than
	// wade through a function that abstracts away arguably the most
	// important logic in the app.)
	err = ss.StoreSession(session.ID)
	if err != nil {
		RemoveSession(session.ID)

		glog.Errorln(err)
		return InvalidAuth, &dt.LoginData{
			CSRF: ss.CSRFToken,
		}, &HTTPError{http.StatusInternalServerError, err}
	}

	// Store some relevant values.
	session.Values[authToken] = ss.AuthToken
	session.Values[authDate] = ss.Date
	session.Values[csrfToken] = ss.CSRFToken
	session.Values["user"] = ss.Email
	session.Values["school"] = ss.School

	// Set cookie termination date for responsible clients.
	// Adjust ns to seconds because MaxAge assumes seconds.
	session.Options.MaxAge = int(endDate / nsConv)

	// Now we save the session in the user's browser.
	err = session.Save(r, w)
	if err != nil {
		glog.Errorln(err)
		return InvalidAuth, &dt.LoginData{
			CSRF: ss.CSRFToken,
		}, &HTTPError{http.StatusInternalServerError, err}
	}

	// Send the user to the dashboard.
	return ValidAuth, &dt.LoginData{
		CSRF:  ss.CSRFToken,
		Redir: paths.DashboardPath,
	}, nil
}

// CheckSession checks to see if a session exists for a user.
// It returns three values:
//     - A pointer to a session if available
//     - A bool which indicates whether the session is authenticated or not
//     - An error if store.Get fails
//
// The structure is this:
// First, we check the user's cookies for our cookie. If store.Get
// gives us an error, we return a nil session, false for authentication,
// and the error.
//
// Theoretically that should never happen because store.Get will return
// a new session if the user did not provide a valid cookie, but we need
// to check for it anyway.
//
// We then test the session to see if it's new. If it is, we jump to our
// error label and return the new session and return false for
// authentication.
//
// We then test the user's cookie for the authToken and the authDate.
// If neither exist, or they cannot be cast to their proper types, []byte
// and int64 respectively, then we jump to our error label.
//
// We then query database (usually memcache/redis, falling back to
// PostgreSQL) for the session that matches 'session.ID' (the unique
// ID attached to each session that's inside our cookie).
//
// The session's values are compared to those provided by the user.
// If -- and only if -- the session's values match those provided by
// the user, we then return the session, true for authentication, and
// nil for any errors.
func CheckSession(r *http.Request) (*sessions.Session, bool, *HTTPError) {
	session, err := store.Get(r, cookieName)
	if err != nil {
		if err == securecookie.ErrMacInvalid {
			return session, false, nil
		}

		glog.Errorln(err)
		return nil, false, &HTTPError{http.StatusInternalServerError,
			database.ErrInternalServerError}
	}

	var (
		userTokenIf interface{} // Raw user auth token before cast
		utExists    bool        // Used in the map check
		userToken   []byte      // Auth token from the user's cookie
		utCastOk    bool        // Can we correctly cast the interface?

		ss *ServerSession

		userDateIf interface{} // Raw user session end date before cast
		udExists   bool        // Used in the map check
		userDate   int64       // User session end date after cast
		udCastOk   bool        // Can we correctly cast the interface?

		now int64 // Used to check cookie expiry dates
	)

	// New sessions are obviously not authenticated.
	if session.IsNew {
		goto inval
	}

	// Grab the token from the user's cookie.
	userTokenIf, utExists = session.Values[authToken]
	if !utExists {
		goto inval
	}

	// Try to cast it to a byte slice.
	userToken, utCastOk = userTokenIf.([]byte)
	if !utCastOk {
		goto inval
	}

	// Grab the expiry date from the user's cookie.
	userDateIf, udExists = session.Values[authDate]
	if !udExists {
		goto inval
	}

	// Try to cast it to an int64.
	userDate, udCastOk = userDateIf.(int64)
	if !udCastOk {
		goto inval
	}

	// Grab our user from the database.
	ss, err = GetSession(session.ID)
	if err != nil {
		goto inval
	}

	// Side-channel timing attacks are irrelevant here. Basically,
	// if the user's able to break our encrypted and authenticated cookie,
	// as well as insert a auth token into, or read one from, our database,
	// as well as figure out the correct key in the map for our token,
	// we have more serious problems than a side-channel attack.
	//
	// Since this is called for each and every page visited (nearly
	// every page requires authentication), it's not worth comparing
	// _all_ of the 64/32 bytes of the token.
	if !bytes.Equal(ss.AuthToken, userToken) {
		goto inval
	}

	now = time.Now().UnixNano()
	if ss.Date != userDate ||
		(ss.Date < now || userDate < now) ||
		(ss.Date < 0 || userDate < 0) {
		goto inval
	}

	return session, true, nil

inval:
	return session, false, nil
}

// DestroySession terminates a session for a given user.
// It returns a boolean indicating whether or not the termination of
// the cookie was successful.
func DestroySession(w http.ResponseWriter, r *http.Request) bool {

	// TODO: Would using an error as a return value be useful here?

	// We don't care if it's valid. Just try to delete everything anyway.
	session, _, httperr := CheckSession(r)
	if httperr != nil {
		return false
	}

	// Set it to have expired before now.
	// This is for responsible clients only.
	session.Options.MaxAge = -1

	// Delete the values from the map.
	delete(session.Values, authToken)
	delete(session.Values, authDate)

	// Remove the session from the database.
	RemoveSession(session.ID)

	err := session.Save(r, w)
	if err != nil {
		glog.Errorln(err)
	}
	return err == nil && session.Options.MaxAge == -1
}

// GetSetSession will either get or set a ServerSession depending on
// whether or not the session exists in the database.
//
// If a session cannot be successfully retrieved, one will be created.
// However, until properly authenticated, the
func GetSetSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) *ServerSession {

	// Attempt to get an existing session.
	ss, err := GetSession(session.ID)
	if err != nil {
		// No session in DB, so create a new one.
		if err == ErrNoSession {
			return createSession(w, r, session)
		}
		glog.Errorln(err)
		return nil
	}
	return ss
}

func createSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) *ServerSession {

	// Each session needs a unique ID in order to be saved.
	if session.ID == "" {
		session.ID = tokens.NewSessionID()
	}

	ss := &ServerSession{
		CSRFToken: tokens.NewCSRFToken(session.ID),
	}

	// Attempt to store the session. Remove the session if it's not stored
	// correctly.
	if err := ss.StoreSession(session.ID); err != nil {
		RemoveSession(session.ID)
		glog.Fatalln(err)
	}

	// Similarly, save it in our FS storage and set the user's cookie.
	if err := session.Save(r, w); err != nil {
		RemoveSession(session.ID)
		glog.Fatalln(err)
	}

	return ss
}

// ComparePassword compares a *User's password with the given password.
// It uses bcrypt's constant time comparison to securely compare the
// passwords.
func ComparePassword(u *database.User, pass []byte, clear bool) (err error) {
	err = bcrypt.CompareHashAndPassword(u.Password, pass)

	// Perhaps this isn't 100% necessary, but typically it's not a bad idea
	// to clear out buffers that hold "secure" information like passwords
	// or SSNs. Clear it with a Goroutine so nothing blocks.
	if clear {
		go clearSlice(pass)
		go clearSlice(u.Password)
	}

	return
}
