package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/EricLagerg/pnwconference/cleanup"
	"github.com/EricLagerg/pnwconference/helpers"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"

	_ "github.com/lib/pq"
)

const (
	getUserStmt = `SELECT email,username,password
		FROM users
		WHERE email = $1 OR username = $1`
)

// 	cacheDuration = 5 * time.Minute
// 	deadDuration  = 12 * time.Hour
// 	cacheNumber   = 5
// 	cachePrefix   = "userdata:"
// )

// // CacheStatus is an enum for the return type of *UserAccessMap.cacheable
// type CacheStatus uint8

// const (
// 	ColdAccess   CacheStatus = iota // Isn't accessed enough
// 	HotAccess                       // Accessed enough to cache
// 	CachedAccess                    // Means it *was* cached
// 	DeadAccess                      // Not accessed in a long time
// )

var (
	// ErrInternalServerError is for occasions when we goof something up
	// but need a pretty error to return to the user. For example,
	// there's a function where we return an error to the user. However,
	// that error could be the result of a bad SQL statement, and
	// the user doesn't need to know how much we suck at SQL. So, instead,
	// we just show this.
	ErrInternalServerError = errors.New("Internal server error.")

	ErrNoUser = errors.New("No users.")

	ErrNoUsernameInCookie = errors.New("No user identifier inside of cookie.")
)

const (
	DBUser    = "postgres"
	DBPasword = "password"
	DBName    = "pnw_conf"

	redisSocket = "/var/run/redis/redis.sock"
)

var (
	DB *sql.DB

	Pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("unix", redisSocket)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
)

func init() {
	var err error // because of scope issues.

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DBUser, DBPasword, DBName)

	DB, err = sql.Open("postgres", dbinfo)
	if err != nil {
		glog.FatalDepth(10, err)
	}
	DB.SetMaxOpenConns(150)

	cleanup.Register("postgres", DB.Close) // Close the database connection.
	cleanup.Register("redis", Pool.Close)  // Close redis connection pool.
}

// CheckUser returns a user's basic information, a bool indicating
// whether the user exists, and an error if one occurred.
func CheckUser(id string) (*User, bool, error) {
	u, err := getUserByID(id)
	return u, err == nil && &u != nil, err
}

// GetAllUsernames returns a slice of all the usernames in the database.
func GetAllUsernames() ([]string, error) {
	var names []string

	rows, err := DB.Query(`SELECT email
		FROM users`)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	if err = rows.Err(); err != nil {
		glog.Errorln(err)
	}

	return names, err
}

// getUserByID returns a user's basic information, sans custom user data,
// via the user's ID (i.e., name or email.)
// The err will describe any database issues.
func getUserByID(id string) (*User, error) {
	conn := Pool.Get()
	defer conn.Close()

	var u User

	// uam.increment(id)

	reply, err := conn.Do("GET", id)
	if reply != nil && err == nil {
		b, ok := reply.([]byte)

		if ok {
			err = json.Unmarshal(b, &u)
			if err != nil {
				glog.Errorln(err)
			}
		}
	} else {
		err = DB.QueryRow(getUserStmt, id).
			Scan(&u.Email, &u.Name, &u.Password)
	}

	if err != nil {
		glog.Errorln(err)

		if err == sql.ErrNoRows {
			err = ErrNoUser
		}
	}

	return &u, err
}

// GetUser returns a user's basic information, sans custom user data.
// The err will describe any database issues.
func GetUser(session *sessions.Session) (*User, error) {
	name, ok := helpers.GetUsername(session)
	if !ok {
		return nil, ErrNoUsernameInCookie
	}

	u, err := getUserByID(name)
	return u, err
}

// GetUserEmail is a more lightweight way to get a user's email that should
// usually skip a database call. It relies on having the user's email inside
// the session, so make sure to validate the session beforehand!
func GetUserEmail(session *sessions.Session) (string, error) {
	emailIf, ok := session.Values["user"]
	if ok {
		email, ok := emailIf.(string)
		if ok {
			return email, nil
		}
	}

	user, err := GetUser(session)
	return user.Email, err
}

// GetUserData simply returns a user's custom data, as well as any database
// errors that occurred.
func GetUserData(session *sessions.Session) (string, error) {
	var data string

	name, ok := helpers.GetUsername(session)
	if !ok {
		return "", ErrInternalServerError
	}

	// uam.increment(name)

	err := DB.QueryRow(`SELECT user_data
		FROM users
		WHERE email = $1 OR username = $1`,
		name).Scan(&data)

	if err != nil {
		glog.Errorln(err)

		if err == sql.ErrNoRows {
			err = ErrNoUser
		}
	}

	return data, err
}

// GetUserAndData returns the entire row of data for a user. It's similar to
// calling GetUser and GetUserData, but more efficient.
func GetUserAndData(session *sessions.Session) (*User, error) {
	var u User

	// Yes, SELECT * because we want the entire row.
	err := DB.QueryRow(`SELECT *
		FROM users
		WHERE email = $1 OR username = $1`,
		session.Values["user"]).
		Scan(&u.Email, &u.Name, &u.Password)

	if err != nil {
		glog.Errorln(err)

		if err == sql.ErrNoRows {
			err = ErrNoUser
		}
	}

	return &u, err
}
