package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/flosch/pongo2/v6"
	readability "github.com/go-shiori/go-readability"
	"golang.org/x/time/rate"
)

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
func postFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		http.Error(w, "Method not allowed "+r.Method, http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	log.Println(r.Form)
	target_url := r.Form.Get("target_url")
	rd := r.Form["readability"]
	log.Println(rd)
	ua := r.Form.Get("target_ua")
	if !validUserAgent(ua) {
		http.Error(w, "Agent not allowed "+ua, http.StatusForbidden)
	}
	var vb = false
	if len(rd) != 0 {
		vb = true
	}
	ckMstr := OurCookie{ua, vb}

	encoded_ua, err := encodeCookie(ckMstr)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

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

func fetch(fetchurl string, user_agent string, rdbl bool) (*http.Response, error) {

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
	if rdbl {
		client = &http.Client{}
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
	if rdbl {
		var tmp2 bytes.Buffer
		io.Copy(&tmp2, resp.Body)
		publishUrl, err := url.Parse(fetchurl)
		if err != nil {
			return resp, err
		}

		article, err := readability.FromReader(&tmp2, publishUrl)
		resp.Body = ioutil.NopCloser(strings.NewReader(article.Content))
	}
	return resp, err

}

var tpl = pongo2.Must(pongo2.FromFile("index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		http.Error(w, "I am not an owl", http.StatusTeapot)
		return
	}

	if r.URL.Path == "/" {
		err := tpl.ExecuteWriter(pongo2.Context{"useragents": UserAgents, "version": version}, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	curl_mode := r.Header.Get("X-BP-Target-UserAgent")

	if curl_mode != "" {
		urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
		if !validUserAgent(curl_mode) {
			http.Error(w, "Agent not allowed "+curl_mode, http.StatusForbidden)
		}
		if len(urlparts) < 2 {
			return
		}

		var mozreader = false

		if r.Header.Get("X-BP-MozReader") != "" {
			mozreader = true
		}
		remurl := urlparts[0] + "//" + urlparts[1]
		_, err := validateURL(remurl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := fetch(remurl, curl_mode, mozreader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
		return

	}
	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}

	remurl := urlparts[0] + "//" + urlparts[1]
	encoded_ua, err := encodeCookie(defaultCookie)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
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
	decagent, err := decodeCookie(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = validateURL(remurl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Header.Get("X-Forwarded-For") != "" {
		log.Printf("%s: %s", r.Header.Get("X-Forwarded-For"), remurl)
	} else {
		log.Printf("%v: %s", r.RemoteAddr, remurl)
	}

	resp, err := fetch(remurl, decagent.UserAgent, decagent.Readability)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)

}

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
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", postFormHandler)
	mux.HandleFunc("/redirect/", postFormHandler)
	mux.HandleFunc("/", rateLimitIndex(indexHandler))

	http.ListenAndServe(":"+port, mux)
}
