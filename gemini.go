package main

/* Gemini Client for poseidon */

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	gemini "git.sr.ht/~adnano/go-gemini"
	"github.com/flosch/pongo2/v6"
	"github.com/vincent-petithory/dataurl"
)

// This will assume the url has already been validated

var currentContext *url.URL

func gmiGet(remote_url string, redirs int) (string, error) {
	var err error

	working, err := url.Parse(remote_url)
	currentContext = working
	if err != nil {
		return "", err
	}
	client := gemini.Client{TrustCertificate: nil}
	target_url, err := url.Parse(remote_url)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	req, err := gemini.NewRequest(remote_url)

	if err != nil {
		return err.Error(), err
	}

	if err != nil {
		return "", err
	}
	resp, err := client.Do(ctx, req)
	if resp.Status == gemini.StatusNotFound {
		return "not found", nil
	}

	log.Println("Metadata " + resp.Meta)
	if itemIn(SupportedImagesTypes, resp.Meta) {

		return genImagePage(resp, remote_url)

	}

	if err != nil {
		return err.Error(), err

	}
	defer resp.Body.Close()
	switch resp.Status.Class() {
	case gemini.StatusInput:
		return "", errors.New("input Unsupported")
	case gemini.StatusRedirect:
		if redirs > 3 {
			return "", errors.New("too Many Rediects")
		}
		redurl, _ := url.Parse(resp.Meta)
		if redurl.IsAbs() {
			rdrresponse, err := gmiGet(resp.Meta, redirs+1)
			currentContext = working
			if err != nil {
				return "", err
			}
			return rdrresponse, nil
		}
		desturl, err := url.JoinPath(target_url.Host, resp.Meta)
		if err != nil {
			return "", err
		}
		rdrresponse, err := gmiGet(desturl, redirs+1)
		currentContext = working
		if err != nil {
			return "", err

		}
		return rdrresponse, nil

	}
	output := bytes.NewBufferString("")
	hw := HTMLWriter{out: output}
	gemini.ParseLines(resp.Body, hw.Handle)
	return output.String(), nil

}
func getCanonicalUrl(current url.URL, urlfrag string) (string, error) {

	canurl, err := url.Parse(urlfrag)
	if err != nil {
		return "gemini://" + current.Hostname(), err
	}
	if canurl.IsAbs() {
		return canurl.String(), nil
	} else if strings.HasPrefix(urlfrag, "/") {
		newurl := url.URL(current)
		newurl.Path = urlfrag
		rv := newurl.String()

		return rv, nil
	} else {
		newurl := url.URL(current)
		cwd := filepath.Dir(newurl.Path)
		newurl.Path, err = url.JoinPath(cwd, urlfrag)
		if err != nil {
			return "gemini://" + current.Hostname(), err
		}
		return newurl.String(), nil

	}

}

// Credit adnano ISC license
type HTMLWriter struct {
	out  io.Writer
	pre  bool
	list bool
}

func (h *HTMLWriter) Handle(line gemini.Line) {
	if _, ok := line.(gemini.LineListItem); ok {
		if !h.list {
			h.list = true
			fmt.Fprint(h.out, "<ul>\n")
		}
	} else if h.list {
		h.list = false
		fmt.Fprint(h.out, "</ul>\n")
	}
	switch line := line.(type) {
	case gemini.LineLink:
		realurl, _ := getCanonicalUrl(*currentContext, html.EscapeString(line.URL))
		url := realurl
		name := html.EscapeString(line.Name)
		if name == "" {
			name = url
		}
		fmt.Fprintf(h.out, "<p><a href='%s'>%s</a></p>\n", url, name)
	case gemini.LinePreformattingToggle:
		h.pre = !h.pre
		if h.pre {
			fmt.Fprint(h.out, "<pre>\n")
		} else {
			fmt.Fprint(h.out, "</pre>\n")
		}
	case gemini.LinePreformattedText:
		fmt.Fprintf(h.out, "%s\n", html.EscapeString(string(line)))
	case gemini.LineHeading1:
		fmt.Fprintf(h.out, "<h1>%s</h1>\n", html.EscapeString(string(line)))
	case gemini.LineHeading2:
		fmt.Fprintf(h.out, "<h2>%s</h2>\n", html.EscapeString(string(line)))
	case gemini.LineHeading3:
		fmt.Fprintf(h.out, "<h3>%s</h3>\n", html.EscapeString(string(line)))
	case gemini.LineListItem:
		fmt.Fprintf(h.out, "<li>%s</li>\n", html.EscapeString(string(line)))
	case gemini.LineQuote:
		fmt.Fprintf(h.out, "<blockquote>%s</blockquote>\n", html.EscapeString(string(line)))
	case gemini.LineText:
		if line == "" {
			fmt.Fprint(h.out, "<br>\n")
		} else {
			fmt.Fprintf(h.out, "<p>%s</p>\n", html.EscapeString(string(line)))
		}
	}
}

func (h *HTMLWriter) Finish() {
	if h.pre {
		fmt.Fprint(h.out, "</pre>\n")
	}
	if h.list {
		fmt.Fprint(h.out, "</ul>\n")
	}
}

// End

func genImagePage(response *gemini.Response, remote_url string) (string, error) {

	imgtmpl, err := pongo2.FromString(ImageProlog)
	if err != nil {
		return err.Error(), err
	}
	ourimg := ImgData{}
	ourimg.Name = remote_url
	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, response.Body)
	if err != nil {
		return err.Error(), err
	}

	du := dataurl.New(buf.Bytes(), response.Meta)

	ourimg.Image = du.String()
	out, err := imgtmpl.Execute(pongo2.Context{"imgdata": ourimg})
	if err != nil {
		return err.Error(), err
	}
	return out, nil

}
