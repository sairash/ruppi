package parser

import (
	"io"
	"ruppi/internal/dom"
	"strings"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) (dom.Node, string, error) {
	doc, err := html.Parse(r)
	if err != nil {

		return dom.Node{}, "", err
	}

	transformedNode, title := transform(doc)

	return transformedNode, title, nil
}

func transform(n *html.Node) (dom.Node, string) {
	var foundTitle string

	if n.Type == html.TextNode {
		trimmedData := strings.TrimSpace(n.Data)
		if trimmedData == "" {

			return dom.Node{Element: dom.ElementData{NodeType: dom.TEXT}, InnerText: ""}, ""
		}

		return dom.Node{Element: dom.ElementData{NodeType: dom.TEXT}, InnerText: n.Data}, ""
	}

	if n.Type == html.ElementNode || n.Type == html.DocumentNode {

		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			foundTitle = n.FirstChild.Data
			return dom.Node{}, foundTitle
		}

		nodeType, ok := dom.TagToType[n.Data]
		if !ok {

			if isBlockTag(n.Data) {
				nodeType = dom.DIV
			} else {
				nodeType = dom.SPAN
			}
		}

		newNode := dom.Node{
			Element: dom.ElementData{
				NodeType: nodeType,
				Name:     n.Data,
				Attrs:    make(map[string]string),
			},
		}

		for _, attr := range n.Attr {
			newNode.Element.Attrs[attr.Key] = attr.Val
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			childNode, childTitle := transform(c)

			if childTitle != "" && foundTitle == "" {
				foundTitle = childTitle
			}

			if childNode.Element.NodeType == dom.TEXT && childNode.InnerText == "" {
				continue
			}
			newNode.Children = append(newNode.Children, childNode)
		}

		return newNode, foundTitle
	}

	return dom.Node{Element: dom.ElementData{NodeType: dom.TEXT}, InnerText: ""}, ""
}

func isBlockTag(tag string) bool {
	switch tag {
	case "article", "section", "header", "footer", "aside", "figure", "figcaption", "main":
		return true
	default:
		return false
	}
}
