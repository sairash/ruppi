package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"ruppi/internal/dom"
	"ruppi/internal/parser"
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
    <li>use <code>q</code> to quit ruppi.</li>
</ul>
<hr>
`

	errorPageString = `<title>Error</title>
<br>
<h1>There was an unexpected error</h1>
<hr>
`
)

func DefaultPage() (dom.Node, string, error) {
	rootNode, title, err := parser.Parse(strings.NewReader(defaultPageString))
	if err != nil {
		return dom.Node{}, "", fmt.Errorf("Failed to parse HTML: %v", err)
	}
	return dom.Node{
		Element: dom.ElementData{
			NodeType: dom.ROOT,
		},
		Children: []dom.Node{
			rootNode,
		},
	}, title, nil
}

func ErrorPage(err error) (dom.Node, string, error) {
	rootNode, title, err := parser.Parse(strings.NewReader(fmt.Sprintf("%s<div>Error: %s</div>", errorPageString, err.Error())))
	if err != nil {
		return dom.Node{}, "", fmt.Errorf("Failed to parse HTML: %v", err)
	}
	return dom.Node{
		Element: dom.ElementData{
			NodeType: dom.ROOT,
		},
		Children: []dom.Node{
			rootNode,
		},
	}, title, nil
}

func GetUrlAsNode(url string) (dom.Node, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return dom.Node{}, "", fmt.Errorf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dom.Node{}, "", fmt.Errorf("Failed to get a valid response: %s", resp.Status)
	}

	rootNode, title, err := parser.Parse(resp.Body)
	if err != nil {
		return dom.Node{}, "", fmt.Errorf("Failed to parse HTML: %v", err)
	}

	// document root node
	return dom.Node{
		Element: dom.ElementData{
			NodeType: dom.ROOT,
		},
		Children: []dom.Node{
			rootNode,
		},
	}, title, nil
}

func SearchURL(text string) string {
	// TODO: Add a real search engine url
	return fmt.Sprintf("https://sairashgautam.com.np?search=%s", text)
}

func IsURL(possibleUrl string) bool {
	_, err := url.ParseRequestURI(possibleUrl)
	if err != nil {
		return false
	}
	return true
}
