package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"regexp"

	"golang.org/x/time/rate"
)

// Add rate limitin per treehouse

// The code i am about to write shouldn't go here.. But
// but this is really the cleanest way to do it given the runaway costs
func rateLimitIndex(next func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateBurst), rateMax)
	agents, _ := regexp.Compile(blocList)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		incoming := request.Header.Get("User-Agent")
		if !limiter.Allow() || agents.MatchString(incoming) {
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

func encodeCookie(c OurCookie) (string, error) {
	first, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	output := base64.URLEncoding.EncodeToString(first)
	return output, nil
}

func decodeCookie(cookieValue string) (OurCookie, error) {
	decodedJson, err := base64.URLEncoding.DecodeString(cookieValue)
	var oc OurCookie
	if err != nil {
		return oc, err
	}
	err = json.Unmarshal([]byte(decodedJson), &oc)
	if err != nil {
		return oc, err
	}
	return oc, nil
}
