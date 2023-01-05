package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func postFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		http.Error(w, "Method not allowed"+r.Method, http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	log.Println(r.Form)
	target_url := r.Form.Get("target_url")
	ua := r.Form.Get("target_ua")
	encoded_ua := base64.URLEncoding.EncodeToString([]byte(ua))

	final := r.URL.Hostname() + "/" + target_url
	cookie := http.Cookie{
		Name:     "blueProxyUserAgent",
		Value:    encoded_ua,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	log.Println(final)
	req, err := http.NewRequest("GET", final, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	req.Header.Add("X-Target-User-Agent", ua)
	w.Header().Set("X-Target-User-Agent", ua)
	http.Redirect(w, req, final, 302)

}

func fetch(fetchurl string, user_agent string) (*http.Response, error) {

	proxyURL, err := url.Parse(ourProxy)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", fetchurl, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("User-Agent", default_agent)
	if user_agent != "" {
		req.Header.Set("User-Agent", user_agent)
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		log.Println("dezipping")
		var tmp bytes.Buffer
		gz, _ := gzip.NewReader(resp.Body)
		io.Copy(&tmp, gz)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(&tmp)
	}
	return resp, err

}

var tpl = pongo2.Must(pongo2.FromFile("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Header.Get("X-Target-User-Agent"))
	if r.URL.Path == "/" {
		err := tpl.ExecuteWriter(pongo2.Context{"useragents": UserAgents, "version": version}, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}

	remurl := urlparts[0] + "//" + urlparts[1]
	encoded_ua := base64.URLEncoding.EncodeToString([]byte(default_agent))
	cookie, err := r.Cookie("blueProxyUserAgent")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			cookie := http.Cookie{
				Name:     "blueProxyUserAgent",
				Value:    encoded_ua,
				Path:     "/",
				MaxAge:   3600,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, &cookie)
			http.Error(w, "Try Again", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
	decagent, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := fetch(remurl, string(decagent))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", postFormHandler)
	mux.HandleFunc("/redirect/", postFormHandler)
	mux.HandleFunc("/", indexHandler)

	http.ListenAndServe("127.0.0.1"+":"+port, mux)
}
