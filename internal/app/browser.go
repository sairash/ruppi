package app

import (
	"fmt"
	"ruppi/internal/logger"
	"ruppi/pkg/httpclient"
	"ruppi/pkg/style"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type active_session int

const (
	ruppiUIBufferSize   = 5
	inspectorBufferSize = 10

	ACTIVE_VIEWPORT active_session = iota
	ACTIVE_INPUT_URL
)

type Browser struct {
	Width        int
	Height       int
	ContentWidth int
	IsKitty      bool
	Ready        bool

	Url      textinput.Model
	Viewport viewport.Model

	IsInspectorOpen   bool
	InspectorViewport viewport.Model

	Tabs       *Tabs
	ActivePane active_session

	Logger *logger.Logger
}

func (b Browser) Init() tea.Cmd {
	return b.Logger.Listen()
}

func (b Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return b, tea.Quit
		}

		if b.Url.Focused() {
			switch msg.String() {
			case "enter":
				cmds = append(cmds, b.submitURL())
			case "esc":
				b.Url.Blur()
				b.ActivePane = ACTIVE_VIEWPORT
			default:
				b.Url, cmd = b.Url.Update(msg)
				cmds = append(cmds, cmd)
			}
			return b, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q":
			return b, tea.Quit
		case "i", "/":
			b.ActivePane = ACTIVE_INPUT_URL
			return b, b.Url.Focus()
		case "?":
			b.IsInspectorOpen = !b.IsInspectorOpen

			return b, toggleInspectorWindow(b.IsInspectorOpen)
		}

	case refreshViewport:
		viewportHeight := b.Height - ruppiUIBufferSize

		if b.IsInspectorOpen {
			viewportHeight -= inspectorBufferSize
		}

		if !b.Ready {
			b.Viewport = viewport.New(b.Width, viewportHeight)
			b.InspectorViewport = viewport.New(b.Width, inspectorBufferSize)
			b.Viewport.SetContent(b.Tabs.Rendered())
			b.InspectorViewport.SetContent(b.Logger.Get())
			b.Ready = true
		} else {
			b.Viewport.Width = b.Width
			b.Viewport.Height = viewportHeight
			b.InspectorViewport.Width = b.Width
			b.InspectorViewport.Height = inspectorBufferSize
			b.InspectorViewport.SetContent(b.Logger.Get())
		}

		b.Url.Width = b.Width - 27

	case updateScrollPosition:
		b.Viewport.ScrollDown(int(msg))
	case updateURL:
		b.Url.SetValue(string(msg))
	case newTabMsg:
		b.Tabs.NewTab(string(msg), b.WordWrap(), b.IsKitty)
		cmds = append(cmds, updateURLCmd(b.Tabs.ActiveTab().url))
		b.Viewport.SetContent(b.Tabs.Rendered())
		b.Viewport.GotoTop()
	case changeTabMsg:
		b.Logger.Add(strconv.Itoa(int(msg)))
		b.Logger.Add(strconv.Itoa(b.Tabs.visibleTabStartIndex))
		b.Tabs.ChangeTab(int(msg))
		cmds = append(cmds, updateURLCmd(b.Tabs.ActiveTab().url))
		b.Viewport.SetContent(b.Tabs.Rendered())
		b.Viewport.GotoTop()
		cmds = append(cmds, updateScrollPositionCmd(b.Tabs.activeTab.scrollPos))

		// b.Tabs.SetScrollPos(b.Tabs.activeTab.scrollPos)
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			if zone.Get("new_tab").InBounds(msg) {
				b.Logger.Add("New tab button clicked")
				cmds = append(cmds, createNewTabCmd(""))
			}

			if zone.Get("go_previous_tab").InBounds(msg) {
				b.Logger.Add("Left tab button clicked")
				b.Tabs.MoveLeft()

			}

			if zone.Get("go_next_tab").InBounds(msg) {
				b.Logger.Add("Right tab button clicked")
				b.Tabs.MoveRight()
			}

			if zone.Get("url_input_bar").InBounds(msg) {
				b.Logger.Add("URL input bar clicked")
				b.ActivePane = ACTIVE_INPUT_URL
				cmds = append(cmds, b.Url.Focus())
			}

			for i := 0; i <= b.Tabs.TotalTabCount; i++ {
				if zone.Get(fmt.Sprintf("%s%d", TAB_ID, i)).InBounds(msg) {
					b.Logger.Add(fmt.Sprintf("%s%d", TAB_ID, i))
					cmds = append(cmds, createChangeTabCmd(i))
					break
				}
			}
		}

	case logger.LogMsg:
		b.InspectorViewport.SetContent(string(msg))
		b.InspectorViewport.ScrollDown(1)

		cmds = append(cmds, b.Logger.Listen())
	case tea.WindowSizeMsg:
		b.Width = msg.Width
		b.Height = msg.Height

		b.Tabs.Render(b.WordWrap(), b.IsKitty)
		cmds = append(cmds, toggleInspectorWindow(b.IsInspectorOpen))
	}

	if b.ActivePane == ACTIVE_VIEWPORT && !b.IsInspectorOpen {
		b.Viewport, cmd = b.Viewport.Update(msg)
		if !b.Viewport.AtTop() {
			b.Tabs.SetScrollPos(int(b.Viewport.ScrollPercent() * float64(b.Viewport.TotalLineCount())))
		}
		cmds = append(cmds, cmd)
	} else if b.ActivePane == ACTIVE_VIEWPORT && b.IsInspectorOpen {
		b.InspectorViewport, cmd = b.InspectorViewport.Update(msg)
		if !b.InspectorViewport.AtTop() {
			b.Tabs.SetScrollPos(int(b.InspectorViewport.ScrollPercent() * float64(b.InspectorViewport.TotalLineCount())))
		}
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}

func (b Browser) View() string {
	if !b.Ready {
		return "\n  Initializing..."
	}

	inspectorWindow := ""
	if b.IsInspectorOpen {
		inspectorWindow = lipgloss.NewStyle().Width(b.Width-2).Border(lipgloss.NormalBorder(), true, false, false).Render(b.InspectorViewport.View())
	}

	statusBar := style.StatusStyle.Width(b.Width - 2).Render(fmt.Sprintf("%s%s%s%s", style.LogoStyle.Render("Ruppi 🐦"), zone.Mark("url_input_bar", style.StatusColor.PaddingLeft(1).Render(b.Url.View())), style.StatusColor.PaddingRight(1).Render(fmt.Sprintf("%3.f%%", b.Viewport.ScrollPercent()*100)), style.LogoStyle.Render("?")))
	tabs := lipgloss.NewStyle().MarginBottom(1).Render(b.Tabs.ShowTabs(b.Width))
	body := fmt.Sprintf("%s%s%s%s", tabs, statusBar, b.Viewport.View(), inspectorWindow)
	return zone.Scan(lipgloss.Place(b.Width, b.Height, lipgloss.Left, lipgloss.Top, style.AppStyle.Width(b.Width).Render(body)))
}

func (b *Browser) WordWrap() int {
	contentWidth := b.ContentWidth
	if contentWidth > 120 {
		contentWidth = 120
	}
	if contentWidth > b.Width {
		contentWidth = b.Width - 2
	}
	return contentWidth
}

func (b *Browser) submitURL() tea.Cmd {
	b.Url.Blur()
	b.ActivePane = ACTIVE_VIEWPORT
	url := b.Url.Value()

	var finalURL string
	if !httpclient.IsURL(url) {
		finalURL = httpclient.SearchURL(url)
	} else {
		finalURL = url
	}

	b.Tabs.ChangeActiveTabURL(finalURL, b.WordWrap(), b.IsKitty)
	b.Viewport.SetContent(b.Tabs.Rendered())
	b.Viewport.GotoTop()
	return updateURLCmd(finalURL)
}
