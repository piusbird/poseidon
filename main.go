package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"net/http/pprof"
)

func main_wrap() int {
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	logfile, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Println("Error opening log")
		return 1
	}
	defer logfile.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	srv.Addr = ":" + port

	fs := http.FileServer(http.Dir("assets"))
	mux := http.NewServeMux()
	debugmode := os.Getenv("DEBUG")
	mux.HandleFunc("/redirect", postFormHandler)
	mux.HandleFunc("/redirect/", postFormHandler)
	mux.HandleFunc("/", LoggingWrapper(logfile, rateLimitIndex(indexHandler)))

	if debugmode != "" {

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	srv.Handler = mux

	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}
func main() {
	os.Exit(main_wrap())
}
