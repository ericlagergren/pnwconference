package controllers

import (
	"net/http"

	"github.com/EricLagerg/pnwconference/auth"
	"github.com/EricLagerg/pnwconference/paths"
	"github.com/EricLagerg/pnwconference/reload"
	"github.com/EricLagerg/pnwconference/views"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

// LogoutActionHandler handles POST requests to "/logout/"
func LogoutActionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, validAuth, httperr := auth.CheckSession(r)
	if httperr != nil {
		views.RenderTemplate(w, r, reload.ErrorPage, httperr.Status, httperr)
		return
	}

	if !validAuth || !auth.ValidCSRF(r, session, false) {
		http.Redirect(w, r, paths.LoginPath, http.StatusFound)
		return
	}

	if !auth.DestroySession(w, r) {
		glog.Errorln(auth.ErrUnableToLogOut)
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, auth.ErrUnableToLogOut)
		return
	}

	http.Redirect(w, r, paths.IndexPath, http.StatusFound)
}
