package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var default_agent = "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
var ourProxy = "http://localhost:8090"

func postFormHandler(w http.ResponseWriter, r *http.Request) {
	//if r.Method != http.MethodPost {

	//http.Error(w, "Method not allowed"+r.Method, http.StatusInternalServerError)
	//	return
	//}
	r.ParseForm()
	log.Println(r.Form)
	target_url := r.Form.Get("target_url")
	ua := r.Form.Get("target_ua")

	final := r.URL.Hostname() + "/" + target_url
	cookie := http.Cookie{
		Name:     "blueProxyUserAgent",
		Value:    ua,
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Header.Get("X-Target-User-Agent"))
	if r.URL.Path == "/" {
		w.Write(template)
		return
	}
	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}
	agent := ""
	remurl := urlparts[0] + "//" + urlparts[1]
	cookie, err := r.Cookie("blueProxyUserAgent")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			cookie := http.Cookie{
				Name:     "blueProxyUserAgent",
				Value:    default_agent,
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
	agent = cookie.Value
	resp, err := fetch(remurl, agent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)

}

func muggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/muggle/" {
		w.Write(template)
		log.Println("It Works!")
		return
	}
	pattern := regexp.MustCompile(`/muggle/`)
	res := pattern.ReplaceAllString(r.URL.Path, "")

	log.Println("Hello!")

	log.Println(res)

	urlparts := strings.SplitN(res, "/", 2)
	if len(urlparts) < 2 {
		return
	}
	remurl := urlparts[0] + "//" + urlparts[1]
	resp, err := fetch(remurl, "")
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
	mux.HandleFunc("/muggle/", muggleHandler)
	mux.HandleFunc("/", indexHandler)

	http.ListenAndServe(":"+port, mux)
}

var template = []byte(
	`<title> Blue DabaDeeDabaProxy </title>
	<body><style>body { font-size: 2em; }</style>
<p>hello there
<form action="/redirect" method="post">
<p>URL: <input name="target_url"> <br/>
<label for="target_ua">Agent?</label>

<select name="target_ua" id="cars">
  <option value="Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0">Desktop</option>
  <option value="Twitterbot/1.0">Twitter</option>
  <option value="Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)">Google</option>
</select> <br><br>
<input type="submit" value="Go">
</form>
`)
