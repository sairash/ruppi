package app

import (
	"fmt"
	"ruppi/internal/config"
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
		documentNode, title, _ = httpclient.ErrorPage(err)
	}
	t.document = documentNode
	t.title = title
	t.url = url

	t.Render(wordWrap, isKitty)
}

type Tabs struct {
	Tabs                 []*Tab
	TotalTabCount        int
	activeTab            *Tab
	activeTabID          int
	visibleTabStartIndex int
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
	for i := ts.visibleTabStartIndex; i < len(ts.Tabs); i++ {
		tab := ts.Tabs[i]

		if k >= tabsThatCanBeContained {
			break
		}

		theme := config.GetTheme()
		title := tab.title

		// Use theme colors for tab styling
		var tabStyle lipgloss.Style
		if ts.activeTabID == tab.id {
			title = fmt.Sprintf("%s%s", "🐦 ", tab.title)
			tabStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(theme.TabActiveColor)).
				Foreground(lipgloss.Color(theme.TabActiveTextColor)).
				Padding(0, 1).
				MarginRight(1)
			// Border(lipgloss.NormalBorder(), false, true, false, false).
			// BorderForeground(lipgloss.Color(theme.TabActiveColor))
		} else {
			tabStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(theme.TabColor)).
				Foreground(lipgloss.Color(theme.TabTextColor)).
				Padding(0, 1).
				MarginRight(1)
			// Border(lipgloss.NormalBorder(), false, true, false, false).
			// BorderForeground(lipgloss.Color("#2d3748"))
		}

		tabContent := zone.Mark(fmt.Sprintf("%s%d", TAB_ID, k),
			string(tabPrefixNumber[k])+" "+
				helper.TruncateString(title, tabsWidth-6, true)+" "+theme.TabCloseIcon)

		tab_str.WriteString(tabStyle.Render(tabContent))
		k += 1
	}

	theme := config.GetTheme()

	// Determine button colors based on state
	moveLeftButtonColor := theme.TabColor
	if ts.visibleTabStartIndex > 0 {
		moveLeftButtonColor = theme.TabActiveColor
	}

	moveRightButtonColor := theme.TabColor
	if ts.visibleTabStartIndex+MAX_TABS_IN_PAGE < len(ts.Tabs) {
		moveRightButtonColor = theme.TabActiveColor
	}

	// Styled navigation buttons
	leftButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TabTextColor)).
		Background(lipgloss.Color(moveLeftButtonColor)).
		Padding(0, 1).
		MarginRight(1)
		// Border(lipgloss.RoundedBorder())

	rightButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TabTextColor)).
		Background(lipgloss.Color(moveRightButtonColor)).
		Padding(0, 1).
		Margin(0, 1)
		// Border(lipgloss.RoundedBorder())

	newTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TabActiveTextColor)).
		Background(lipgloss.Color(theme.TabActiveColor)).
		Padding(0, 1)
		// Border(lipgloss.RoundedBorder())

	return zone.Mark("go_previous_tab", leftButtonStyle.Render(theme.TabPrevIcon)) +
		style.TabContainerColor().Width(tabContainerWidth).Render(tab_str.String()) +
		zone.Mark("go_next_tab", rightButtonStyle.Render(theme.TabNextIcon)) +
		zone.Mark("new_tab", newTabStyle.Render(theme.TabNewIcon))
}

func (ts *Tabs) MoveLeft() {
	if ts.visibleTabStartIndex > 0 {
		ts.visibleTabStartIndex -= 1
	}
}

func (ts *Tabs) MoveRight() {
	if ts.visibleTabStartIndex+MAX_TABS_IN_PAGE < len(ts.Tabs) {
		ts.visibleTabStartIndex += 1
	}
}

func (ts *Tabs) ChangeTab(id int) {
	ts.activeTab = ts.Tabs[ts.visibleTabStartIndex+id]
	ts.activeTabID = ts.visibleTabStartIndex + id
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

	changeVisibleTabIndex := ts.TotalTabCount + 1 - MAX_TABS_IN_PAGE
	if changeVisibleTabIndex >= 0 {
		ts.visibleTabStartIndex = changeVisibleTabIndex
	}
}

func (ts *Tabs) SetScrollPos(pos int) {
	ts.activeTab.setScrollPos(pos)
}
