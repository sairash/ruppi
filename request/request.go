package request

import (
	"fmt"
	"net/http"
	"rupi/element"
	"rupi/parser"
)

func GetUrlAsNode(url string) (element.Node, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return element.Node{}, "", fmt.Errorf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return element.Node{}, "", fmt.Errorf("Failed to get a valid response: %s", resp.Status)
	}

	rootNode, title, err := parser.Parse(resp.Body)
	if err != nil {
		return element.Node{}, "", fmt.Errorf("Failed to parse HTML: %v", err)
	}

	// document node
	return element.Node{
		Element: element.ElementData{
			NodeType: element.ROOT,
		},
		Children: []element.Node{
			rootNode,
		},
	}, title, nil
}
