package controllers

import (
	"net/http"

	"github.com/EricLagerg/pnwconference/reload"
	"github.com/EricLagerg/pnwconference/views"

	"github.com/julienschmidt/httprouter"
)

// HomeViewHandler handles GET requests to "/"
func HomeViewHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	views.RenderTemplate(w, r, reload.Index, http.StatusOK, nil)
}
