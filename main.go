package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"net/http/pprof"

	"github.com/flosch/pongo2/v6"
	readability "github.com/go-shiori/go-readability"

	"golang.org/x/time/rate"
	"piusbird.space/poseidon/nuparser"
)

var homeURL string = "http://localhost:3000"

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
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(r.Form)
	target_url := r.Form.Get("target_url")
	rd := r.Form["readability"]
	log.Println(rd)
	ua := r.Form.Get("target_ua")
	if !validUserAgent(ua) {
		http.Error(w, "Agent not allowed "+ua, http.StatusForbidden)
	}
	vb := ArcParser
	if len(rd) != 0 {
		vb = NUParser
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
	http.Redirect(w, req, final, http.StatusFound)

}
func gmiFetch(fetchurl string) (*http.Response, error) {
	resp := http.Response{
		Body: io.NopCloser(bytes.NewBufferString("Gemini Contenet")),
	}
	tpl, err := pongo2.FromString(Header)
	if err != nil {
		return nil, err
	}
	rawhtml, err := gmiGet(fetchurl, 0)
	if err != nil {
		return nil, err
	}
	uu, err := url.Parse(fetchurl)
	if err != nil {
		return nil, err
	}
	tmpbuf := strings.NewReader(rawhtml)
	filteredContent, err := RewriteLinks(tmpbuf, homeURL)
	inbuf := strings.NewReader(filteredContent)
	if err != nil {
		log.Println("Failed filter pass on " + fetchurl)
		inbuf = strings.NewReader(rawhtml)
	}

	article, err := readability.FromReader(inbuf, uu)
	if err != nil {
		return nil, err
	}
	out, err := tpl.Execute(pongo2.Context{"article": article, "url": fetchurl})
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(strings.NewReader(out))
	return &resp, nil

}

// FIXME: This code is basically a pile of dung
// Templates render where they shouldn't, miniweb should be deprecated etc, etc
// Shpuld also move this into it's own file
func fetch(fetchurl string, user_agent string, parser_select bool, original *http.Request) (*http.Response, error) {

	tpl, err := pongo2.FromString(Header)
	if err != nil {
		return nil, err

	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	u, err := url.Parse(original.RequestURI)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newQueryString := u.Query()
	origQuery, _ := url.ParseQuery(original.URL.RawQuery)
	if _, ok := origQuery["engine"]; !ok {
		newQueryString.Set("engine", "1")
	} else {
		newQueryString.Del("engine")
	}
	u.RawQuery = newQueryString.Encode()
	lightswitch := u.String()

	client := &http.Client{}
	client.Transport = tr

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
	defer resp.Body.Close()

	var tmp bytes.Buffer
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		log.Println("Yes we gziped")

		gz, _ := gzip.NewReader(resp.Body)

		contentSize := resp.ContentLength

		if contentSize > maxBodySize {
			return nil, errors.New("response body to large")
		}

		decompBuffMax := maxBodySize * 2
		log.Println("dezipping")

		for {
			var bytesRead int64 = 0
			n, err := io.CopyN(&tmp, gz, 4096)
			if errors.Is(err, io.EOF) {
				break
			}
			bytesRead += n
			if bytesRead > decompBuffMax {
				return nil, errors.New("decompression failed")
			}

		}

		err = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(&tmp)

	}
	var tmp2 bytes.Buffer
	_, err = io.Copy(&tmp2, resp.Body)
	if err != nil {
		return nil, err
	}
	if tmp.Len() < 1 {
		return nil, errors.New("watson this is weird")
	}

	publishUrl, err := url.Parse(fetchurl)
	if err != nil {
		return resp, err
	}

	var article GenaricArticle

	if parser_select {
		raw_article, err := readability.FromReader(&tmp2, publishUrl)
		if err != nil {
			return nil, err
		}
		article = GenaricArticle{}
		article.Byline = raw_article.Byline
		article.Content = raw_article.Content
		article.Title = raw_article.Title
		article.Length = raw_article.Length
		article.Image = raw_article.Image
	} else {
		raw_article, err := nuparser.FromReader(&tmp2)
		if err != nil {
			return nil, err
		}
		article = GenaricArticle{}
		article.Byline = raw_article.Byline
		article.Content = raw_article.Content
		article.Title = raw_article.Title
		article.Length = raw_article.Length
		article.Image = raw_article.Image
	}

	tmp_content := strings.NewReader(article.Content)
	backupContent := strings.Clone(article.Content)
	filteredContent, err := RewriteLinks(tmp_content, homeURL)

	article.Content = filteredContent
	if err != nil {
		log.Println("failed filter pass " + fetchurl)
		article.Content = backupContent
	}

	out, err := tpl.Execute(pongo2.Context{"article": article, "url": fetchurl, "switchurl": lightswitch})
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(strings.NewReader(out))

	return resp, err

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fakeCookie := http.Cookie{
		Name:     "blueProxyUserAgent",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	if r.Method == http.MethodPost {
		http.Error(w, "I am not an owl", http.StatusTeapot)
		return
	}
	var tpl = pongo2.Must(pongo2.FromFile("index.html"))

	if r.URL.Path == "/" {
		err := tpl.ExecuteWriter(pongo2.Context{"useragents": UserAgents, "version": version}, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	queryParams, _ := url.ParseQuery(r.URL.RawQuery)
	var err error
	homeURL = "http://" + r.Host
	log.Println(homeURL)
	if err != nil {
		http.Error(w, "Lost my soul", http.StatusInternalServerError)
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
		// Confusing part needed to hook up gemini starts here
		// Basically we skip validation if it's a gemini uri and
		// do our own thing with it

		remurl := urlparts[0] + "//" + urlparts[1]
		ur, err := url.Parse(remurl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Honk!")
		log.Println(ur.String())
		if ur.Scheme == "gemini" {
			remurl += r.URL.RawQuery
			resp, err := gmiFetch(remurl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer resp.Body.Close()
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		_, err = validateURL(remurl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := fetch(remurl, curl_mode, mozreader, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	}
	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}

	remurl := urlparts[0] + "//" + urlparts[1]
	encoded_ua, err := encodeCookie(defaultCookie)
	fakeCookie.Value = encoded_ua

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	ur, err := url.Parse(remurl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Honk!")
	log.Println(ur.String())
	if ur.Scheme == "gemini" {
		remurl += r.URL.RawQuery
		resp, err := gmiFetch(remurl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_, err = validateURL(remurl)
	if err != nil {
		http.Error(w, err.Error()+" "+remurl, http.StatusInternalServerError)
		return
	}
	var cookie *http.Cookie
	cookie, err = r.Cookie("blueProxyUserAgent")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			cookie = &fakeCookie

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

	if r.Header.Get("X-Forwarded-For") != "" {
		log.Printf("%s: %s", r.Header.Get("X-Forwarded-For"), remurl)
	} else {
		log.Printf("%v: %s", r.RemoteAddr, remurl)
	}
	var parser_select bool
	if _, ok := queryParams["engine"]; !ok {
		parser_select = !bool(decagent.Parser)
	} else {
		parser_select = bool(decagent.Parser)
	}

	resp, err := fetch(remurl, decagent.UserAgent, parser_select, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

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
	mux.HandleFunc("/", rateLimitIndex(indexHandler))

	if debugmode != "" {

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	srv.Handler = mux

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
