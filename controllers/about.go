package controllers

import (
	"net/http"

	"github.com/EricLagerg/pnwconference/reload"
	"github.com/EricLagerg/pnwconference/views"

	"github.com/julienschmidt/httprouter"
)

// AboutViewHandler handles GET requests to "/about/"
func AboutViewHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	views.RenderTemplate(w, r, reload.About, http.StatusOK, nil)
}
