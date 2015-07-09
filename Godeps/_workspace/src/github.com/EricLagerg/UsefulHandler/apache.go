package useful

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ErrUnHijackable indicates an unhijackable connection. I.e., (one of)
// the underlying http.ResponseWriter(s) doesn't support the http.Hijacker
// interface.
var ErrUnHijackable = errors.New("A(n) underlying ResponseWriter doesn't support the http.Hijacker interface")

// These format strings correspond with the log formats described in
// https://httpd.apache.org/docs/2.2/mod/mod_log_config.html
const (
	// CommonLog is "%h %l %u %t \"%r\" %>s %b"
	commonLogFmt = "%s - - [%s] \"%s %d %d\" %f\n"

	// CommonLogWithVHost is "%v %h %l %u %t \"%r\" %>s %b"
	commonLogWithVHostFmt = "%s %s - - [%s] \"%s %d %d\" %f\n"

	// NCSALog is
	// "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-agent}i\""
	ncsaLogFmt = "%s - - [%s] \"%s %d %d\" %f\n \"%s\" \"%s\""

	// RefererLog is "%{Referer}i -> %U"
	refererLogFmt = "%s -> %s"

	// AgentLog is "%{User-agent}i"
	agentLogFmt = "%s"
)

type (
	// LogPrinter is a struct with a string format. An example use
	// of the format would be inside the Print method.
	// This allows more extensibility by providing the ability
	// to change the pre-defined formats.
	// LogPrinter         struct{ Format string }
	commonLog          struct{ Format string }
	commonLogWithVHost struct{ Format string }
	ncsaLog            struct{ Format string }
	refererLog         struct{ Format string }
	agentLog           struct{ Format string }
)

// Log format types.
var (
	CommonLog          = commonLog{commonLogFmt}
	CommonLogWithVHost = commonLogWithVHost{commonLogWithVHostFmt}
	NCSALog            = ncsaLog{ncsaLogFmt}
	RefererLog         = refererLog{refererLogFmt}
	AgentLog           = agentLog{agentLogFmt}
)

// ApacheLogRecord is a structure containing the necessary information
// to write a proper log in the ApacheFormatPattern.
type ApacheLogRecord struct {
	http.ResponseWriter
	LogFmt

	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
	referer, agent        string
}

// Hijack implements the http.Hijacker interface to allow connection
// hijacking.
func (a *ApacheLogRecord) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	hj, ok := a.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrUnHijackable
	}
	return hj.Hijack()
}

// Log will log an entry to the io.Writer specified by LogDestination.
func (r *ApacheLogRecord) Log(out io.Writer) {

	n := r.Print(out, r)

	if LogFile.size+int64(n) >= LogFile.Opts.MaxFileSize {
		LogFile.Rotate()
	}

	LogFile.size += int64(n)
}

// Write fulfills the Write method of the http.ResponseWriter interface.
func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(n)
	return n, err
}

// WriteHeader fulfills the WriteHeader method of the http.ResponseWriter
// interface.
func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// ServeHTTP fulfills the ServeHTTP method of the http.Handler interface.
func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	record := &ApacheLogRecord{
		ResponseWriter: rw,
		LogFmt:         LogFile.Opts.LogFormat,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
		referer:        r.Referer(),
		agent:          r.UserAgent(),
	}

	startTime := time.Now()
	h.handler.ServeHTTP(record, r)
	finishTime := time.Now()

	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)

	LogFile.Lock()
	defer LogFile.Unlock()
	record.Log(LogFile.out)
}
