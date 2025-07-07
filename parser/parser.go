package parser

import (
	"io"
	"rupi/element"
	"strings"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) (element.Node, string, error) {
	doc, err := html.Parse(r)
	if err != nil {

		return element.Node{}, "", err
	}

	transformedNode, title := transform(doc)

	return transformedNode, title, nil
}

func transform(n *html.Node) (element.Node, string) {
	var foundTitle string

	if n.Type == html.TextNode {
		trimmedData := strings.TrimSpace(n.Data)
		if trimmedData == "" {

			return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: ""}, ""
		}

		return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: n.Data}, ""
	}

	if n.Type == html.ElementNode || n.Type == html.DocumentNode {

		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			foundTitle = n.FirstChild.Data
			return element.Node{}, foundTitle
		}

		nodeType, ok := element.TagToType[n.Data]
		if !ok {

			if isBlockTag(n.Data) {
				nodeType = element.DIV
			} else {
				nodeType = element.SPAN
			}
		}

		newNode := element.Node{
			Element: element.ElementData{
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

			if childNode.Element.NodeType == element.TEXT && childNode.InnerText == "" {
				continue
			}
			newNode.Children = append(newNode.Children, childNode)
		}

		return newNode, foundTitle
	}

	return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: ""}, ""
}

func isBlockTag(tag string) bool {
	switch tag {
	case "article", "section", "header", "footer", "aside", "figure", "figcaption", "main":
		return true
	default:
		return false
	}
}
