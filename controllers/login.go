package controllers

import (
	"net/http"

	"github.com/EricLagerg/pnwconference/auth"
	dt "github.com/EricLagerg/pnwconference/controllers/datatypes"
	"github.com/EricLagerg/pnwconference/database"
	"github.com/EricLagerg/pnwconference/paths"
	"github.com/EricLagerg/pnwconference/reload"
	"github.com/EricLagerg/pnwconference/views"

	"github.com/julienschmidt/httprouter"
)

// LoginViewHandler handles GET requests to "/login/"
func LoginViewHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Specific headers for our login page.
	w.Header().Set("pragma", "no-cache")
	w.Header().Set("cache-control", "no-cache, no-store")
	w.Header().Set("expires", "Mon, 01-Jan-1990 00:00:00 GMT")

	session, validAuth, httperr := auth.CheckSession(r)
	if httperr != nil {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, httperr)
	}

	// If the user is logged in then just redirect to the dashboard.
	// This is why the logic may look a little backwards.
	if validAuth {
		http.Redirect(w, r, paths.DashboardPath, http.StatusFound)
		return
	}

	ss := auth.GetSetSession(w, r, session)
	if ss == nil {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, database.ErrInternalServerError)
		return
	}

	views.RenderTemplate(w, r, reload.Login, http.StatusOK, &dt.LoginData{r.Host, ss.CSRFToken, "", ""})
}

// LoginActionHandler handles POST requests to "/login/"
func LoginActionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// The CSRF token check is inside AuthenticateUser.
	status, data, err := auth.AuthenticateUser(w, r)

	// Errors from AuthenticateUser or a status of InvalidAuth indicate that
	// the user is not authenticated, and we should handle the response
	// accordingly.
	if err != nil {
		switch err.Err {
		// Re-render with error information.
		case auth.ErrBadUsername, auth.ErrBadPassword, auth.ErrInvalidLogin,
			auth.ErrTooManyRequests, auth.ErrInvalidCSRFToken:
			views.RenderTemplate(w, r, reload.Login, err.Status, data)
		default:
			views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, err)
		}

		return
	}

	if status == auth.ValidAuth {
		// It's okay to send the user to their original destination.
		http.Redirect(w, r, data.Redir, http.StatusFound)
		return
	} else {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, database.ErrInternalServerError)
		return
	}

	http.Redirect(w, r, paths.LoginPath, http.StatusOK)
}
