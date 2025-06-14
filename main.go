package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"rupi/config"
	"rupi/element"
	"rupi/parser"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	statusBarHeight = 1
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

type Browser struct {
	width        int
	height       int
	contentWidth int
	isKitty      bool
	title        string
	url          string
	// url          textinput.Model
	document element.Node
	ready    bool
	rendered string

	viewport viewport.Model

	scrollPos int
}

func main() {
	err := config.LoadConfig("./rupi.conf")

	urlFlag := flag.String("url", "", "The URL to parse and render.")
	kittyFlag := flag.Bool("kitty", true, "Enable Kitty terminal font size extensions.")
	contentWidth := flag.Int("width", 0, "Content word wrap, default 80")
	flag.Parse()

	if *urlFlag == "" {
		fmt.Println("Please provide a URL with the -url flag.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	resp, err := http.Get(*urlFlag)
	if err != nil {
		log.Fatalf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get a valid response: %s", resp.Status)
	}

	rootNode, title, err := parser.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	termProgram := os.Getenv("TERM")

	width, height, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}

	ti := textinput.New()
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	ti.Placeholder = "Search DuckDuckGo or type Url"
	ti.SetValue(*urlFlag)
	ti.Focus()
	// ti.Width
	ti.Prompt = "Url: "

	b := Browser{
		width:        width,
		height:       height,
		title:        title,
		contentWidth: *contentWidth,
		url:          *urlFlag,
		isKitty:      strings.Contains(termProgram, "kitty") || *kittyFlag,
		document: element.Node{
			Element: element.ElementData{
				NodeType: element.ROOT,
			},
			Children: []element.Node{
				rootNode,
			},
		},
		scrollPos: 0,
		ready:     false,
	}
	b.wordWrap()

	p := tea.NewProgram(b, tea.WithAltScreen())
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
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return b, tea.Quit
		}

		// if b.url.Focused() {
		// 	b.url, cmd = b.url.Update(msg)
		// 	cmds = append(cmds, cmd)
		// }
	case tea.WindowSizeMsg:

		b.width = msg.Width
		b.height = msg.Height
		b.rendered = element.WordWrap(b.document.Render(b.isKitty), b.wordWrap())

		if !b.ready {
			b.viewport = viewport.New(b.width-5, msg.Height-3)
			b.viewport.SetContent(b.rendered)
			b.ready = true
		} else {
			b.viewport.Width = msg.Width - 10
			b.viewport.Height = msg.Height - 2
		}
	}

	b.viewport, cmd = b.viewport.Update(message)
	cmds = append(cmds, cmd)

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

	statusBar := statusStyle.Width(b.width - 2).Render(fmt.Sprintf("%s%s%s", logoStyle.Render("Rupi 🐦"), statusColor.Width(b.width-20).PaddingLeft(1).Render(b.url), statusColor.Render(fmt.Sprintf("%3.f%%", b.viewport.ScrollPercent()*100))))

	title := fmt.Sprintf("%s \n", tytleStyle.Render(b.title))
	body := fmt.Sprintf("%s\n%s%s", statusBar, title, b.viewport.View())
	return lipgloss.Place(b.width, b.height, lipgloss.Left, lipgloss.Top, appStyle.Width(b.width).Render(body))
}
