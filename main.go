package main

import (
	"net/http"
	"os"
	"net/url"
	"io/ioutil"
	"log"
	"crypto/tls"
	"compress/gzip"
	"bytes"
	"io"
	"strings"
)

var ourProxy = "http://localhost:8090"
func fetch(fetchurl string)  (*http.Response, error) {

	proxyURL, err := url.Parse(ourProxy)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
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

var hello = []byte(
	`<body><style>body { font-size: 2em; }</style>
<p>hello there
<form action="javascript:void(0);" onsubmit="go()">
<p>URL: <input id="urlbox">
<script>
function go() {
    var inp = document.getElementById("urlbox")
    var val = inp.value
    window.location.href = window.location.href + val
}
</script>
<button onclick="go()">GO</button>
</form>
`)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Write(hello)
		return
	}
	urlparts := strings.SplitN(r.URL.Path[1:], "/", 2)
	if len(urlparts) < 2 {
		return
	}
	remurl := urlparts[0] + "//" + urlparts[1]
	resp, err := fetch(remurl)
	if err != nil {
		log.Println(err)
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

	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}


