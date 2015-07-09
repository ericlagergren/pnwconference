package views

import (
	"errors"
	"net/http"

	"github.com/EricLagerg/pnwconference/reload"

	"github.com/golang/glog"
)

const (
	htmlMIME = "text/html; charset=utf-8"
	jsonMIME = "application/json; charset=utf-8"
	textMIME = "text/plain; charset=utf-8"
)

var ErrTemplateDoesntExist = errors.New("Template does not exist.")

// RenderTemplate accepts the name of a template and a piece of data (usually
// a structure) and executes it into the request's handler.
//
// RenderTemplate will write an error page if an error occurs. This breaks
// from what other functions typically do, but it saves having to constantly
// type 'if renderTemplate(...) != nil { /* process error */ }' all over the
// place.
func RenderTemplate(w http.ResponseWriter, r *http.Request, name reload.TmplName, status int, data interface{}) (err error) {

	defer func() {
		e, ok := recover().(error)
		if ok {
			err = ErrTemplateDoesntExist
			glog.Errorf("Error %v caused %s", e, err)
		}
	}()

	w.Header().Set("Content-Type", htmlMIME)
	w.WriteHeader(status)

	// Dat stutter.
	tmpl := reload.Tmpls.Templates[name]
	if err = tmpl.ExecuteTemplate(w, "base", data); err != nil {
		glog.Errorln(err)
	}

	return
}
