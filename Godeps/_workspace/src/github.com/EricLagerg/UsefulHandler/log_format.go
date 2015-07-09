package useful

import (
	"fmt"
	"io"
)

// LogFmt is the interface implemented by log types to print an
// ApacheLogRecord in the desired format.
type LogFmt interface {
	Print(w io.Writer, r *ApacheLogRecord) int
}

const (
	timeFormat    = "02/Jan/2006 03:04:05"
	requestFormat = "%s %s %s"
)

// timeRequest returns the formatted time of the request and the request line.
func (r *ApacheLogRecord) formattedTimeRequest() (string, string) {
	return r.time.Format(timeFormat), fmt.Sprintf(requestFormat, r.method, r.uri, r.protocol)
}

func (l commonLog) Print(w io.Writer, r *ApacheLogRecord) int {
	timeFormatted, requestLine := r.formattedTimeRequest()

	n, _ := fmt.Fprintf(w, l.Format, r.ip, timeFormatted,
		requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())

	return n
}

func (l commonLogWithVHost) Print(w io.Writer, r *ApacheLogRecord) int {
	timeFormatted, requestLine := r.formattedTimeRequest()

	n, _ := fmt.Fprintf(w, l.Format, r.ip, timeFormatted,
		requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())

	return n
}

func (l ncsaLog) Print(w io.Writer, r *ApacheLogRecord) int {
	timeFormatted, requestLine := r.formattedTimeRequest()

	n, _ := fmt.Fprintf(w, l.Format, r.ip,
		timeFormatted, requestLine, r.status, r.responseBytes,
		r.elapsedTime.Seconds(), r.referer, r.agent)

	return n
}

func (l refererLog) Print(w io.Writer, r *ApacheLogRecord) int {
	n, _ := fmt.Fprintf(w, l.Format, r.referer, r.uri)
	return n
}

func (l agentLog) Print(w io.Writer, r *ApacheLogRecord) int {
	n, _ := fmt.Fprintf(w, l.Format, r.agent)
	return n
}
