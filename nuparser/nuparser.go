package nuparser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Article struct {
	Title   string
	Byline  string
	Content string
	Image   string
	Length  int
}
type ParserResult struct {
	Result     string
	EntryPoint *html.Node
}

var allowdTags = []string{
	"a", "b", "base", "blockquote", "body", "br", "center", "code", "dd",
	"dfn", "div", "dl", "dt", "em", "h1", "h2", "h3", "h4", "h5", "h6",
	"head", "hr", "html", "i", "img", "kbd", "li", "link", "meta", "ol",
	"p", "pre", "samp", "title", "ul", "var", "article",
	"audio", "span", "div", "section", "main", "aside",
}

func FromReader(input io.Reader) (Article, error) {
	rv := Article{}
	parsed, err := Parse(input)
	outstr := parsed.Result
	if err != nil {
		return rv, err
	}
	rv.Content = outstr
	rv.Title = extractTitle(parsed.EntryPoint)
	rv.Byline = "Gregory Brightwing"
	rv.Length = len(outstr)
	return rv, nil

}

func Parse(rdr io.Reader) (ParserResult, error) {
	doc, err := html.Parse(rdr)
	if err != nil {
		return ParserResult{}, err
	}
	rv := ParserResult{}
	rv.EntryPoint = doc
	fullrender := reconstructHTML(doc)
	rv.Result = fullrender
	return rv, nil

}
func reconstructHTML(n *html.Node) string {
	var sb strings.Builder

	// Reconstruct the node
	reconstructNodeHTML(&sb, n)

	// Reconstruct child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		reconstructNodeHTML(&sb, c)
	}

	return sb.String()
}

func reconstructNodeHTML(sb *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		// Reconstruct element node
		reconstructElementNodeHTML(sb, n)
	case html.TextNode:
		// Reconstruct text node
		reconstructTextNodeHTML(sb, n)
	}
}

func reconstructElementNodeHTML(sb *strings.Builder, n *html.Node) {
	// Check if the element tag is available in our reduced feature set
	if isAllowedTag(n.Data) {
		// Reconstruct element tag
		sb.WriteString("<")
		sb.WriteString(n.Data)
		sb.WriteString(">")

		// Reconstruct child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			reconstructNodeHTML(sb, c)
		}

		// Reconstruct closing tag
		sb.WriteString("</")
		sb.WriteString(n.Data)
		sb.WriteString(">")
	}
}

func reconstructTextNodeHTML(sb *strings.Builder, n *html.Node) {
	// Reconstruct text content
	sb.WriteString(n.Data)
}

func isAllowedTag(tagName string) bool {

	// Check if the tag name is in the HTML2 tag list
	for _, tag := range allowdTags {
		if tag == tagName {
			return true
		}
	}

	return false
}
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {

		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title := extractTitle(c)
		if title != "" {
			return title
		}
	}

	return ""
}
