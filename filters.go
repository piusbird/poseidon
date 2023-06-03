package main

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func RewriteLinks(reqBody io.Reader, urlPrefix string) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reqBody)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(urlPrefix)
	if err != nil {
		return "", err

	}
	htmlDoc := buf.String()
	// Parse the HTML document
	doc, err := html.Parse(strings.NewReader(htmlDoc))
	if err != nil {
		return "", err
	}

	// Walk through the parsed HTML tree
	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// Get the value of the href attribute
			for i, attr := range n.Attr {
				if attr.Key == "href" {

					// Replace the value of the href attribute with the proxy URL
					newurl, _ := url.JoinPath(u.String(), attr.Val)
					if strings.Contains(attr.Val, "youtube.com") == true {
						newurl = rewriteYT(attr.Val)
					}

					n.Attr[i].Val = newurl
					break
				}
			}
		}

		// Recursively walk through the child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	// Serialize the modified HTML document
	var sb strings.Builder
	err = html.Render(&sb, doc)
	if err != nil {
		return "", err
	}

	// Print the modified HTML document
	return sb.String(), nil
}

func rewriteYT(link string) string {
	u, _ := url.Parse(link)
	u.Host = YTProxy
	return u.String()

}
