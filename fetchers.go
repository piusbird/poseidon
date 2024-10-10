package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	readability "github.com/go-shiori/go-readability"
	"piusbird.space/poseidon/nuparser"
)

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

//
// build_request Builds a client side http request for the fetcher

func build_http_request(url string, user_agent string) (*http.Request, error) {

	req, err := http.NewRequest("GET", url, nil)
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
	return req, nil
}

// FIXME: This code is basically a pile of dung
// Templates render where they shouldn't, miniweb should be deprecated etc, etc
// Shpuld also move this into it's own file

func fetch(fetchurl string, user_agent string, parser_select bool, original *http.Request) (*http.Response, error) {

	tpl, err := pongo2.FromString(Header)
	if err != nil {
		return nil, err

	}
	jar, _ := cookiejar.New(nil)

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
	client.Jar = jar

	req, err := build_http_request(fetchurl, user_agent)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Printf("Request Status %v", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		client.Do(req)
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
		article.Text = raw_article.TextContent
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
	if strings.HasPrefix(original.Header.Get("User-Agent"), "curl") {
		prettyBody := fmt.Sprintf("%s By %s\n %s\n ", article.Title, article.Byline, article.Text)
		resp.Body = io.NopCloser(strings.NewReader(prettyBody))
		return resp, err
	}
	resp.Body = io.NopCloser(strings.NewReader(out))

	return resp, err

}
