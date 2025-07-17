package app

import (
	"log"
	"math/rand"
	"ruppi/internal/dom"
	"ruppi/pkg/helper"
	"ruppi/pkg/httpclient"
	"ruppi/pkg/style"
	"strings"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	TRUNCATE_MIN = 13
)

var (
	tabPrefixNumber = []rune("🯰🯱🯲🯳🯴🯵🯶🯷🯸🯹")
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

// func merge()

func (ts *Tabs) ShowTabs(width int) string {
	tab_str := strings.Builder{}
	tabContainerWidth := width - 14
	k := 0

	TabBackgroundColor := "#202020"
	for id, tab := range ts.Tabs {
		if id == ts.activeTabID {
			TabBackgroundColor = "#3a3a3a"
		} else {
			TabBackgroundColor = "#202020"
		}

		tab_str.WriteString(style.TabContainerColor.PaddingRight(1).Render(lipgloss.NewStyle().Background(lipgloss.Color(TabBackgroundColor)).Render(style.PaddingX.Render(string(tabPrefixNumber[k])) + helper.TruncateString(tab.title, 10, true) + style.PaddingX.Render("𜸲"))))
		k += 1
	}

	return zone.Mark("go_previous_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).MarginRight(1).Background(lipgloss.Color("30")).Render("<")) +
		style.TabContainerColor.Width(tabContainerWidth).Render(tab_str.String()) +
		zone.Mark("go_next_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).Margin(0, 1).Background(lipgloss.Color("30")).Render(">")) +
		zone.Mark("new_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).Background(lipgloss.Color("30")).Render("+"))
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
