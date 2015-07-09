package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EricLagerg/pnwconference/auth"
	"github.com/EricLagerg/pnwconference/database"
	"github.com/EricLagerg/pnwconference/paths"
	"github.com/EricLagerg/pnwconference/reload"
	"github.com/EricLagerg/pnwconference/views"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

type Signup struct {
	Email  string
	First  string
	Last   string
	School string
	State  string
}

func (s *Signup) Store() error {
	_, err := database.DB.Exec(`WITH new_values (
			email, first, last, school, state) as (
		  		values ($1::text,
		  				$2::text,
		  				$3::text,
		  				$4::text,
		  				$5::text,
		  			)
		),
		upsert as
		( 
		    UPDATE sessions m
				SET first  = nv.first,
		            last   = nv.last,
		            school = nv.school,
		    FROM new_values nv
		    WHERE m.session_id = nv.session_id
		    RETURNING m.*
		)
		INSERT INTO sessions (email, first, last, school, state)
		SELECT email, first, last, school, state
		FROM new_values
		WHERE NOT EXISTS (SELECT 1
		                  FROM upsert up
		                  WHERE up.email = new_values.email)`,
		s.Email, s.First, s.Last, s.School, s.State)

	if err != nil {
		glog.Errorln(err)
	}
	return err
}

func (s *Signup) Remove() error {
	_, err := database.DB.Exec(`DELETE
		FROM signups
		WHERE email = $1`, s.Email)
	if err != nil {
		glog.Errorln(err)
	}
	return err
}

type SignupData struct {
	CSRF  []byte
	Error string
}

func SignupViewHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	session, _, httperr := auth.CheckSession(r)
	if httperr != nil {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, httperr)
	}

	ss := auth.GetSetSession(w, r, session)
	if ss == nil {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, database.ErrInternalServerError)
		return
	}

	views.RenderTemplate(w, r, reload.Signup, http.StatusOK, &SignupData{ss.CSRFToken, ""})
}

func SignupActionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, validAuth, httperr := auth.CheckSession(r)
	if !validAuth || !auth.ValidCSRF(r, session, false) || httperr != nil {
		http.Redirect(w, r, paths.SignupPath, http.StatusUnauthorized)
		return
	}

	ss := auth.GetSetSession(w, r, session)
	if ss == nil {
		views.RenderTemplate(w, r, reload.ErrorPage, http.StatusInternalServerError, database.ErrInternalServerError)
		return
	}

	reg := &Signup{
		First:  r.PostFormValue("_fname"),
		Last:   r.PostFormValue("_lname"),
		Email:  r.PostFormValue("_email"),
		School: r.PostFormValue("_school"),
		State:  r.PostFormValue("_state"),
	}

	if err := reg.validate(); err != nil {
		views.RenderTemplate(w, r, reload.Signup, http.StatusOK,
			&SignupData{
				ss.CSRFToken,
				err.Error(),
			})
		return
	}

	reg.Store()

	http.Redirect(w, r, paths.ThankYouPath, http.StatusFound)
}

// meh.
func (s *Signup) validate() error {
	var errors []string

	if s.First == "" {
		errors = append(errors, "First name")
	}

	if s.Last == "" {
		errors = append(errors, "Last name")
	}

	if s.Email == "" {
		errors = append(errors, "Email")
	}

	if s.School == "" {
		errors = append(errors, "School")
	}

	if s.State == "" {
		errors = append(errors, "State")
	}

	if errors != nil {
		fields := strings.Join(errors, ", ")
		return fmt.Errorf("Please fill out field(s) %s", fields)
	}
	return nil
}
