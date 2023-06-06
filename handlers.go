package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/flosch/pongo2/v6"
)

var homeURL string = "http://localhost:3000"

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
		vb = ArcParser
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
		Secure:   false,
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fakeCookie := http.Cookie{
		Name:     "blueProxyUserAgent",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
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

	requesterUserAgent := r.Header.Get("User-Agent")

	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}

	remurl := urlparts[0] + "//" + urlparts[1]
	encoded_ua, err := encodeCookie(defaultCookie)
	fakeCookie.Value = encoded_ua
	if strings.HasPrefix(requesterUserAgent, "curl") {
		_, err = validateURL(remurl)

		if err != nil {
			http.Error(w, err.Error()+" "+remurl, http.StatusTeapot)
			return
		}
		ur, _ := url.Parse(remurl)
		if ur.Scheme == "gemini" {
			http.Error(w, "Gemini not supported through curl", http.StatusBadGateway)
			return
		}
		a, err := fetch(remurl, default_agent, bool(ArcParser), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.Copy(w, a.Body)
		return

	}

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
		http.Error(w, err.Error()+" "+remurl, http.StatusTeapot)
		return
	}
	var cookie *http.Cookie
	cookie, err = r.Cookie("blueProxyUserAgent")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			cookie = &fakeCookie
			http.SetCookie(w, cookie)
			http.Redirect(w, r, r.RequestURI, http.StatusSeeOther)

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
