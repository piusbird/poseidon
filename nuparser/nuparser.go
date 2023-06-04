package nuparser

import (
	"fmt"
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
	"head", "hr", "html", "i", "img", "kbd", "li", "meta", "ol",
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
	fullrender := buildRestrictedHTML(doc)
	rv.Result = fullrender
	return rv, nil

}
func buildRestrictedHTML(n *html.Node) string {
	var sb strings.Builder

	// Reconstruct the node
	emitNode(&sb, n)

	// Reconstruct child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		emitNode(&sb, c)
	}
	return sb.String()
}

func emitNode(sb *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		// Reconstruct element node
		emitElementNode(sb, n)
	case html.TextNode:
		// Reconstruct text node
		emitTextNode(sb, n)
	}
}

func emitElementNode(sb *strings.Builder, n *html.Node) {
	// Allowed tags get emited traversed and then closed
	if isAllowedTag(n.Data) {
		sb.WriteString(TagToString(n))

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			emitNode(sb, c)
		}

		// Reconstruct closing tag
		sb.WriteString(closingTagString(n.Data))

	} else {
		// If the tag is not allowed emit nothing but traverse it's children for
		// allowed members, in case we've hit some dumb html5ish container element
		// Which seem to be proliferating thanks to the "living standard"
		//for c := n.FirstChild; c != nil; c = c.NextSibling {
		//	emitNode(sb, c)
		//}
		sb.WriteString("")

		// Reconstruct closing tag
		//sb.WriteString(closingTagString(n.Data))

	}

}

func emitTextNode(sb *strings.Builder, n *html.Node) {
	// Just emit the text between the element tags as is
	sb.WriteString(n.Data)
}

func isAllowedTag(tagName string) bool {

	for _, tag := range allowdTags {
		if tag == tagName {
			return true
		}
	}

	return false
}

func TagToString(n *html.Node) string {
	var sb strings.Builder
	sb.WriteString("<")
	sb.WriteString(n.Data)
	for _, attr := range n.Attr {
		if attr.Key == "style" {
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Key, attr.Val))
	}
	sb.WriteString(">")
	return sb.String()
}
func closingTagString(tagName string) string {
	return fmt.Sprintf("</%s>", tagName)
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
