package controllers

import (
	"github.com/EricLagerg/pnwconference/paths"

	"github.com/julienschmidt/httprouter"
)

var Router = NewRouter()

// NewRouter returns a new router.
func NewRouter() *httprouter.Router {
	r := httprouter.New()

	r.GET("/", HomeViewHandler)

	r.GET(paths.LoginPath, LoginViewHandler)
	r.POST(paths.LoginPath, LoginActionHandler)

	r.POST(paths.LogoutPath, LogoutActionHandler)

	r.GET(paths.SignupPath, SignupViewHandler)
	r.POST(paths.SignupPath, SignupActionHandler)

	r.GET(paths.AboutPath, AboutViewHandler)

	return r
}
