package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"rupi/config"
	"rupi/element"
	"rupi/request"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	zone "github.com/lrstanley/bubblezone"
)

const (
	STAUTS_BAR_HEIGHT = 1

	ACTIVE_VIEWPORT = iota
	ACTIVE_INPUT_URL
)

var (
	tytleStyle     = lipgloss.NewStyle().Bold(true)
	appStyle       = lipgloss.NewStyle().Padding(0, 1)
	BorderTopStyle = lipgloss.NewStyle().Background(lipgloss.Color("29"))
	bodyStyle      = lipgloss.NewStyle()

	logoStyle   = lipgloss.NewStyle().Background(lipgloss.Color("200")).PaddingRight(1).PaddingLeft(1).Bold(true)
	statusColor = lipgloss.NewStyle().Background(lipgloss.Color("#242424")).Foreground(lipgloss.Color("#7D7D7D"))
	statusStyle = statusColor.MarginBottom(1)
)

type BrowerTab struct {
	id       string
	document element.Node
	rendered string
}

type Browser struct {
	width        int
	height       int
	contentWidth int
	isKitty      bool
	title        string
	url          textinput.Model
	// url          textinput.Model
	document element.Node
	rendered string
	ready    bool

	tabs        []BrowerTab
	curTabIndex int

	viewport viewport.Model

	scrollPos  int
	activePane int
}

func main() {
	err := config.LoadConfig("./rupi.conf")

	urlFlag := flag.String("url", "", "The URL to parse and render.")
	kittyFlag := flag.Bool("kitty", true, "Enable Kitty terminal font size extensions.")
	contentWidth := flag.Int("width", 0, "Content word wrap, default 80")
	flag.Parse()

	var documentNode element.Node
	var title string

	if *urlFlag == "" {
		documentNode, title, err = request.DefaultPage()
	} else {
		documentNode, title, err = request.GetUrlAsNode(*urlFlag)
	}

	if err != nil {
		log.Fatal(err)
	}

	zone.NewGlobal()
	defer zone.Close()

	termProgram := os.Getenv("TERM")

	width, height, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}

	ti := textinput.New()
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	ti.Placeholder = "Search DuckDuckGo or type Url"
	ti.SetValue(*urlFlag)
	ti.Blur()
	ti.Prompt = "Url: "

	b := Browser{
		width:        width,
		height:       height,
		title:        title,
		contentWidth: *contentWidth,
		url:          ti,
		isKitty:      strings.Contains(termProgram, "kitty") || *kittyFlag,
		document:     documentNode,
		scrollPos:    0,
		ready:        false,
		activePane:   ACTIVE_VIEWPORT,
	}
	b.wordWrap()

	p := tea.NewProgram(b, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (b Browser) Init() tea.Cmd {
	return nil
}

func (b Browser) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		if b.url.Focused() {
			b.url, cmd = b.url.Update(msg)
			cmds = append(cmds, cmd)
		}

		switch msg.String() {
		case "ctrl+c":
			return b, tea.Quit
		case "q":
			if b.activePane == ACTIVE_VIEWPORT {
				return b, tea.Quit
			}
		case "i":
			if b.activePane == ACTIVE_VIEWPORT {
				b.url.Focus()
				b.activePane = ACTIVE_INPUT_URL
			} else {
				b.url.Blur()
				b.activePane = ACTIVE_VIEWPORT
			}
		case "esc":
			if b.activePane == ACTIVE_INPUT_URL {
				b.url.Blur()
				b.activePane = ACTIVE_VIEWPORT
			}
		case "enter":
			if b.url.Focused() {
				b.url.Blur()
				b.activePane = ACTIVE_VIEWPORT
				documentNode, title, err := request.GetUrlAsNode(b.url.Value())

				if err != nil {
					log.Fatal(err)
				}
				b.document = documentNode
				b.title = title

				b.rendered = element.WordWrap(b.document.Render(b.isKitty), b.wordWrap())
				b.viewport.SetContent(b.rendered)
			}
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			if zone.Get("url_input_bar").InBounds(msg) {
				b.url.Focus()
				b.activePane = ACTIVE_INPUT_URL
			}
		}

	case tea.WindowSizeMsg:

		b.width = msg.Width
		b.height = msg.Height
		b.rendered = element.WordWrap(b.document.Render(b.isKitty), b.wordWrap())

		if !b.ready {
			b.viewport = viewport.New(b.width-5, msg.Height-3)
			b.viewport.SetContent(b.rendered)
			b.ready = true
		}
		//  else {
		// b.viewport.Width = msg.Width - 10
		// b.viewport.Height = msg.Height - 2
		// }

		b.url.Width = b.width - 25
	}

	if b.activePane == ACTIVE_VIEWPORT {
		b.viewport, cmd = b.viewport.Update(message)
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}

func (b Browser) wordWrap() int {
	contentWidth := 80
	if b.contentWidth > 120 {
		contentWidth = 120
	}

	if contentWidth > b.width {
		contentWidth = b.width - 5
	}

	return contentWidth
}

func (b Browser) View() string {
	if !b.ready {
		return "\n  Initializing..."
	}

	statusBar := statusStyle.Width(b.width - 2).Render(fmt.Sprintf("%s%s%s", logoStyle.Render("Rupi 🐦"), zone.Mark("url_input_bar", statusColor.PaddingLeft(1).Render(b.url.View())), statusColor.Render(fmt.Sprintf("%3.f%%", b.viewport.ScrollPercent()*100))))

	title := fmt.Sprintf("%s \n", tytleStyle.Render(b.title))
	body := fmt.Sprintf("%s\n%s%s", statusBar, title, b.viewport.View())
	return zone.Scan(lipgloss.Place(b.width, b.height, lipgloss.Left, lipgloss.Top, appStyle.Width(b.width).Render(body)))
}
