package main

/* Gemini Client for poseidon */

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"

	gemini "git.sr.ht/~adnano/go-gemini"
	"git.sr.ht/~adnano/go-gemini/tofu"
)

// This will assume the url has already been validated

var hostKeys tofu.KnownHosts

func gmiGet(remote_url string) (string, error) {

	client := gemini.Client{TrustCertificate: nil}
	ctx := context.Background()
	req, err := gemini.NewRequest(remote_url)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(ctx, req)
	if err != nil {
		return "Gemini Error", err

	}
	defer resp.Body.Close()
	switch resp.Status.Class() {
	case gemini.StatusInput:
		return "", errors.New("input Unsupported")
	case gemini.StatusRedirect:
		return "", errors.New("Redirects no worky it's on the list")

	}
	output := bytes.NewBufferString("")
	hw := HTMLWriter{out: output}
	gemini.ParseLines(resp.Body, hw.Handle)
	return output.String(), nil

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
		url := html.EscapeString(line.URL)
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
