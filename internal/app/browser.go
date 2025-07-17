package app

import (
	"fmt"
	"ruppi/pkg/httpclient"
	"ruppi/pkg/style"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type active_session int

const (
	ruppiUIBufferSize = 5

	ACTIVE_VIEWPORT active_session = iota
	ACTIVE_INPUT_URL
)

type Browser struct {
	Width        int
	Height       int
	ContentWidth int
	IsKitty      bool
	Ready        bool

	Url        textinput.Model
	Viewport   viewport.Model
	Tabs       *Tabs
	ActivePane active_session
}

func (b Browser) NewTab(url string) {
	b.Tabs.NewTab("", b.WordWrap(), b.IsKitty)
	b.Viewport.SetContent(b.Tabs.Rendered())
}

func (b Browser) Init() tea.Cmd {
	return nil
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
				b.submitURL()
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
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			if zone.Get("new_tab").InBounds(msg) {
				b.NewTab("")
			}
			if zone.Get("url_input_bar").InBounds(msg) {
				b.ActivePane = ACTIVE_INPUT_URL
				cmds = append(cmds, b.Url.Focus())
			}
		}

	case tea.WindowSizeMsg:
		b.Width = msg.Width
		b.Height = msg.Height

		b.Tabs.Render(b.WordWrap(), b.IsKitty)

		if !b.Ready {
			b.Viewport = viewport.New(b.Width, msg.Height-ruppiUIBufferSize)
			b.Viewport.SetContent(b.Tabs.Rendered())
			b.Ready = true
		} else {
			b.Viewport.Width = msg.Width
			b.Viewport.Height = msg.Height - ruppiUIBufferSize
		}

		b.Url.Width = b.Width - 27
	}

	if b.ActivePane == ACTIVE_VIEWPORT {
		b.Viewport, cmd = b.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}

func (b Browser) View() string {
	if !b.Ready {
		return "\n  Initializing..."
	}

	statusBar := style.StatusStyle.Width(b.Width - 2).Render(fmt.Sprintf("%s%s%s%s", style.LogoStyle.Render("Ruppi 🐦"), zone.Mark("url_input_bar", style.StatusColor.PaddingLeft(1).Render(b.Url.View())), style.StatusColor.PaddingRight(1).Render(fmt.Sprintf("%3.f%%", b.Viewport.ScrollPercent()*100)), style.LogoStyle.Render("?")))
	tabs := lipgloss.NewStyle().MarginBottom(1).Render(b.Tabs.ShowTabs(b.Width))
	body := fmt.Sprintf("%s%s%s", tabs, statusBar, b.Viewport.View())
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

func (b *Browser) submitURL() {
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
}
