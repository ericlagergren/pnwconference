package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"

	"../"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	useful.LogFile.Init()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Hello, world!")
	})

	myHandler := useful.NewUsefulHandler(handler)

	http.Handle("/", myHandler)
	server := http.Server{
		Addr:    ":1234",
		Handler: nil,
	}
	server.ListenAndServe()
}
