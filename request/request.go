package request

import (
	"fmt"
	"net/http"
	"rupi/element"
	"rupi/parser"
	"strings"
)

const (
	defaultPageString = `<title>Page Title</title>
<br>
<h1>This is the default Page</h1>
<hr>
<p>What can I do here?</p>
<ul>
	<li>use <code>i</code> or click on the url bar.</li>
    <li>use <code>h</code> to get help.</li>
    <li>use <code>q</code> to quit Rupi.</li>
</ul>
<hr>
`
)

func DefaultPage() (element.Node, string, error) {
	rootNode, title, err := parser.Parse(strings.NewReader(defaultPageString))
	if err != nil {
		return element.Node{}, "", fmt.Errorf("Failed to parse HTML: %v", err)
	}
	return element.Node{
		Element: element.ElementData{
			NodeType: element.ROOT,
		},
		Children: []element.Node{
			rootNode,
		},
	}, title, nil
}

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
