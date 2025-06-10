package parser

import (
	"io"
	"strings"

	"rupi/element"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) (element.Node, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return element.Node{}, err
	}
	return transform(doc), nil
}

func transform(n *html.Node) element.Node {
	if n.Type == html.TextNode {
		trimmedData := strings.TrimSpace(n.Data)
		if trimmedData == "" {
			return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: ""}
		}
		return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: n.Data}
	}

	if n.Type == html.ElementNode || n.Type == html.DocumentNode {
		nodeType, ok := element.TagToType[n.Data]
		if !ok {
			if isBlockTag(n.Data) {
				nodeType = element.DIV
			} else {
				nodeType = element.SPAN
			}
		}

		// Create our node
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
			childNode := transform(c)
			if childNode.Element.NodeType == element.TEXT && childNode.InnerText == "" {
				continue
			}
			newNode.Children = append(newNode.Children, childNode)
		}
		return newNode
	}

	return element.Node{Element: element.ElementData{NodeType: element.TEXT}, InnerText: ""}
}

func isBlockTag(tag string) bool {
	switch tag {
	case "article", "section", "header", "footer", "aside", "figure", "figcaption", "main":
		return true
	default:
		return false
	}
}
