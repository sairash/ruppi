package app

import (
	"log"
	"math/rand"
	"rupi/internal/dom"
	"rupi/pkg/httpclient"
)

type Tab struct {
	id            int
	document      dom.Node
	rendered      string
	title         string
	scrollPos     float64
	renderedWidth int
}

func (t *Tab) Render(wordwrap int, isKitty bool) {
	t.rendered = dom.WordWrap(t.document.Render(isKitty), wordwrap)
	t.renderedWidth = wordwrap
}

func (t *Tab) ChangeURL(url string, wordWrap int, isKitty bool) {
	documentNode, title, err := httpclient.GetUrlAsNode(url)
	if err != nil {
		log.Printf("Failed to get URL %s: %v", url, err)
		documentNode, title, _ = httpclient.ErrorPage(err)
	}
	t.document = documentNode
	t.title = title

	t.Render(wordWrap, isKitty)
}

type Tabs struct {
	Tabs        map[int]*Tab
	activeTab   *Tab
	activeTabID int
}

func (ts *Tabs) Render(wordWrap int, isKitty bool) {
	if ts.activeTab != nil {
		ts.activeTab.Render(wordWrap, isKitty)
	}
}

func (ts *Tabs) Rendered() string {
	if ts.activeTab == nil {
		return ""
	}
	return ts.activeTab.rendered
}

func (ts *Tabs) ActiveTab() *Tab {
	if ts.activeTab == nil {
		return &Tab{title: "No active tab"}
	}
	return ts.activeTab
}

func (ts *Tabs) ChangeActiveTabURL(url string, wordWrap int, isKitty bool) {
	if ts.activeTab != nil {
		ts.activeTab.ChangeURL(url, wordWrap, isKitty)
	}
}

func (ts *Tabs) NewTab(url string, wordWrap int, isKitty bool) {
	var documentNode dom.Node
	var title string
	var err error

	if url == "" {
		documentNode, title, err = httpclient.DefaultPage()
	} else {
		documentNode, title, err = httpclient.GetUrlAsNode(url)
	}

	if err != nil {
		log.Printf("Failed to create new tab for URL %s: %v", url, err)
		documentNode, title, _ = httpclient.ErrorPage(err)
	}

	tab := &Tab{
		id:        rand.Int(),
		document:  documentNode,
		title:     title,
		scrollPos: 0,
	}

	tab.Render(wordWrap, isKitty)

	ts.Tabs[tab.id] = tab
	ts.activeTabID = tab.id
	ts.activeTab = tab
}
