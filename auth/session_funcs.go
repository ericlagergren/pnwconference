package auth

import (
	"database/sql"
	"errors"

	"github.com/EricLagerg/pnwconference/database"

	"github.com/golang/glog"
)

var ErrNoSession = errors.New("No sessions")

// ServerSession is the data-view of a user's session found both in redis
// and our database.
type ServerSession struct {
	ID        []byte `redis:"-"` // Session ID
	AuthToken []byte `redis:"auth_token"`
	CSRFToken []byte `redis:"csrf_token"`
	Email     string `redis:"email"`  // User's email
	School    string `redis:"school"` // User's school
	Date      int64  `redis:"date"`   // Session expiry date
}

// StoreSession stores a ServerSession in both memcached/redis and the
// backing PostgreSQL database.
func (s *ServerSession) StoreSession(id string) error {
	conn := database.Pool.Get()
	defer conn.Close()

	obj, err := s.MarshalJSON()
	if err != nil {
		glog.Errorln(err)
		return err
	}

	reply, err := conn.Do("SET", id, obj)
	if err != nil {
		glog.Errorln(err, reply)
		return err
	}

	_, err = database.DB.Exec(`WITH new_values (
			session_id, auth_token, csrf_token, email, school, date) as (
		  		values ($1::text,
		  				$2::bytea,
		  				$3::bytea,
		  				$4::text,
		  				$5::text,
		  				$6::bigint
		  			)
		),
		upsert as
		( 
		    UPDATE sessions m
				SET auth_token = nv.auth_token,
		            csrf_token = nv.csrf_token,
		            email      = nv.email,
		            school     = nv.school,
		            date       = nv.date
		    FROM new_values nv
		    WHERE m.session_id = nv.session_id
		    RETURNING m.*
		)
		INSERT INTO sessions (session_id, auth_token, csrf_token, email, school, date)
		SELECT session_id, auth_token, csrf_token, email, school, date
		FROM new_values
		WHERE NOT EXISTS (SELECT 1
		                  FROM upsert up
		                  WHERE up.session_id = new_values.session_id)`,
		id, s.AuthToken, s.CSRFToken, s.Email, s.School, s.Date)

	if err != nil {
		glog.Fatal(err)
	}

	return err
}

// GetSession returns a SeverSession from the database.
// It checks memcached/redis first, and then falls back on the
// database.
func GetSession(id string) (*ServerSession, error) {
	conn := database.Pool.Get()
	defer conn.Close()

	var ss ServerSession

	reply, err := conn.Do("GET", id)
	if reply != nil && err == nil {
		b, ok := reply.([]byte)

		if ok {
			err = ss.UnmarshalJSON(b)
			if err != nil {
				glog.Errorln(err)
			}
		}
	} else {
		err = database.DB.QueryRow(`SELECT *
			FROM sessions
			WHERE session_id=$1`, id).
			Scan(&ss.ID, &ss.AuthToken, &ss.CSRFToken, &ss.Email, &ss.School, &ss.Date)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoSession
		} else {
			glog.Errorln(err)
		}
	}

	return &ss, err
}

// RemoveSession removes a ServerSession from the database.
// It checks memcached/redis first, and even if the session does
// not exist in memcached/redis, it still removes it from the database.
// It returns two errors, one for removing the session from redis and one
// from removing the session from PostgreSQL. The errors are nil if nothing
// went wrong.
func RemoveSession(id string) (error, error) {
	conn := database.Pool.Get()
	defer conn.Close()

	// "DEL" returns an integer number of keys that were removed. However,
	// we don't need to validate whether or not the session was removed from
	// redis because we don't actually know if the session was *in* redis in
	// the first place.
	_, er := conn.Do("DEL", id)

	_, ed := database.DB.Exec(`DELETE
		FROM sessions
		WHERE session_id = $1`, id)

	return er, ed
}
