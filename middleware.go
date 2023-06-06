package main

import (
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

// Add rate limitin per treehouse
func rateLimitIndex(next func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateBurst), rateMax)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !limiter.Allow() {
			http.Error(writer, "Enhance your calm", 420)
			return
		} else {
			next(writer, request)
		}
	})
}

func LoggingWrapper(log *os.File, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
		t := time.Now()
		log.WriteString(r.RemoteAddr + " " + t.Format(time.UnixDate) + " " + r.RequestURI + "\n")
	}
}
