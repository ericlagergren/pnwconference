package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/EricLagerg/pnwconference/cleanup"
	"github.com/EricLagerg/pnwconference/controllers"
	"github.com/EricLagerg/pnwconference/paths"

	"github.com/golang/glog"
	"github.com/unrolled/secure"

	useful "github.com/EricLagerg/UsefulHandler"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

var secureOpts *secure.Secure

func init() {
	// Check the github repo and set these options properly when the site
	// goes live.
	secureOpts = secure.New(secure.Options{
		AllowedHosts:         []string{"localhost"},
		SSLRedirect:          false,
		SSLHost:              "ssl.example.com",
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		// ContentSecurityPolicy: "default-src 'self'",
		IsDevelopment: true,
	})

	// Logging (in conjunction with compression).
	opts := useful.DefaultOptions()
	opts.LogDestination = useful.File
	opts.LogName = filepath.Join("log", "access.log")
	opts.ArchiveDir = filepath.Join("log", "archives")
	useful.LogFile.Init(opts)

	cleanup.Register("glog", glog.Flush)             // Make sure logs flush before we exit.
	cleanup.Register("useful", useful.LogFile.Close) // Close our log file.

}

func main() {
	fmt.Println("Compilation complete.")

	flag.Parse()

	// Theoretically we should catch most signals.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case s := <-ch:
			if glog.V(2) {
				glog.Infof("Caught signal %s.", s.String())
			}

			// This is ran when a signal similar to ctrl+c is caught. It allows
			// us to do nice things like flush logs, close fds, etc. It's not
			// 100% necessary, but nice and easy to do.
			cleanup.RunAndQuit(s)
		}
	}()

	r := controllers.Router

	// Serve our static files, e.g. CSS, JS.
	dir, err := os.Getwd()
	if err != nil {
		glog.Fatalln(err)
	}
	dir = filepath.Join(dir, "/static/")
	r.ServeFiles("/static/*filepath", http.Dir(dir))

	handler := useful.NewUsefulHandler(secureOpts.Handler(r))

	go func() {
		if err := http.ListenAndServeTLS(paths.PQDN+":443", "keys/cert.pem", "keys/server.key", handler); err != nil {
			glog.Fatalln(err)
		}
	}()

	if err := http.ListenAndServe(paths.PQDN+":80", http.HandlerFunc(redir)); err != nil {
		glog.Fatalln(err)
	}
}

func redir(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+paths.PQDN+r.RequestURI, http.StatusMovedPermanently)
}
