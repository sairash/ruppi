package app

import (
	"fmt"
	"log"
	"ruppi/internal/dom"
	"ruppi/pkg/helper"
	"ruppi/pkg/httpclient"
	"ruppi/pkg/style"
	"strings"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	TRUNCATE_MIN_WIDTH       = 10
	REMOVE_EXTRA_TAB_BUTTONS = 14
	MAX_TABS_IN_PAGE         = 9
)

var (
	tabPrefixNumber = []rune("🯱🯲🯳🯴🯵🯶🯷🯸🯹")
)

type Tab struct {
	id            int
	document      dom.Node
	rendered      string
	title         string
	scrollPos     int
	renderedWidth int
	url           string
}

func (t *Tab) Render(wordwrap int, isKitty bool) {
	t.rendered = dom.WordWrap(t.document.Render(isKitty), wordwrap)
	t.renderedWidth = wordwrap
}

func (t *Tab) setScrollPos(pos int) {
	t.scrollPos = pos
}

func (t *Tab) ChangeURL(url string, wordWrap int, isKitty bool) {
	documentNode, title, err := httpclient.GetUrlAsNode(url)
	if err != nil {
		log.Printf("Failed to get URL %s: %v", url, err)
		documentNode, title, _ = httpclient.ErrorPage(err)
	}
	t.document = documentNode
	t.title = title
	t.url = url

	t.Render(wordWrap, isKitty)
}

type Tabs struct {
	Tabs          []*Tab
	TotalTabCount int
	activeTab     *Tab
	activeTabID   int

	showingTabsIndex struct {
		from int
		to   int
	}
}

func (ts *Tabs) Render(wordWrap int, isKitty bool) {
	if ts.activeTab != nil {
		ts.activeTab.Render(wordWrap, isKitty)
	}
}

func (ts *Tabs) Rendered() string {
	if ts.activeTab == nil {
		return "Initializing..."
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

func (ts *Tabs) ShowTabs(width int) string {
	tab_str := strings.Builder{}
	tabContainerWidth := width - REMOVE_EXTRA_TAB_BUTTONS

	tabsWidth := TRUNCATE_MIN_WIDTH
	tabsThatCanBeContained := (tabContainerWidth / TRUNCATE_MIN_WIDTH) - 1

	if tabsThatCanBeContained >= MAX_TABS_IN_PAGE {
		tabsWidth = tabContainerWidth / (MAX_TABS_IN_PAGE + 1)
		tabsThatCanBeContained = MAX_TABS_IN_PAGE
	}

	k := 0
	tabBackgroundColor := "#202020"
	for id, tab := range ts.Tabs {
		if k >= tabsThatCanBeContained {
			break
		}
		if id == ts.activeTabID {
			tabBackgroundColor = "#3a3a3a"
		} else {
			tabBackgroundColor = "#202020"
		}

		tab_str.WriteString(style.TabContainerColor.PaddingRight(1).Render(lipgloss.NewStyle().Background(lipgloss.Color(tabBackgroundColor)).Render(zone.Mark(fmt.Sprintf("%s%d", TAB_ID, k), style.PaddingX.Render(string(tabPrefixNumber[k]))+helper.TruncateString(tab.title, tabsWidth-6, true)) + style.PaddingX.Render("x"))))
		k += 1
	}

	return zone.Mark("go_previous_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).MarginRight(1).Background(lipgloss.Color("30")).Render("<")) +
		style.TabContainerColor.Width(tabContainerWidth).Render(tab_str.String()) +
		zone.Mark("go_next_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).Margin(0, 1).Background(lipgloss.Color("30")).Render(">")) +
		zone.Mark("new_tab", style.PaddingX.Foreground(lipgloss.NoColor{}).Background(lipgloss.Color("30")).Render("+"))
}

func (ts *Tabs) ChangeTab(id int) {
	ts.activeTab = ts.Tabs[id]
	ts.activeTabID = id
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
		id:        len(ts.Tabs),
		document:  documentNode,
		title:     title,
		scrollPos: 0,
		url:       url,
	}

	tab.Render(wordWrap, isKitty)
	ts.TotalTabCount = len(ts.Tabs)

	ts.Tabs = append(ts.Tabs, tab)
	ts.activeTabID = tab.id
	ts.activeTab = tab
}

func (ts *Tabs) SetScrollPos(pos int) {
	ts.activeTab.setScrollPos(pos)
}
