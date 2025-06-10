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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	appStyle       = lipgloss.NewStyle().Padding(0, 1)
	BorderTopStyle = lipgloss.NewStyle().Background(lipgloss.Color("29"))
	bodyStyle      = lipgloss.NewStyle()
)

type Browser struct {
	width    int
	height   int
	isKitty  bool
	url      textinput.Model
	document element.Node

	scrollPos int
}

func main() {
	err := config.LoadConfig("./rupi.conf")

	urlFlag := flag.String("url", "", "The URL to parse and render.")
	kittyFlag := flag.Bool("kitty", true, "Enable Kitty terminal font size extensions.")
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

	rootNode, err := parser.Parse(resp.Body)
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
	ti.Width = width - 10
	ti.SetValue(*urlFlag)
	ti.Focus()
	// ti.Width
	ti.Prompt = "Url: "

	p := tea.NewProgram(Browser{
		width:   width,
		height:  height,
		url:     ti,
		isKitty: strings.Contains(termProgram, "kitty") || *kittyFlag,
		document: element.Node{
			Element: element.ElementData{
				NodeType: element.ROOT,
			},
			Children: []element.Node{
				rootNode,
			},
		},
		scrollPos: 0,
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (b Browser) Init() tea.Cmd {
	return nil
}

func (b Browser) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return b, tea.Quit
		}

		if b.url.Focused() {
			b.url, cmd = b.url.Update(msg)
		}
	}

	return b, cmd
}

func (b Browser) View() string {
	x := b.document.Render(b.isKitty)

	value := fmt.Sprintf("%s\n%s", bodyStyle.Width(b.width-4).Height(b.height-2).Render(x), BorderTopStyle.Width(b.width-4).Render(b.url.View()))
	return lipgloss.Place(b.width, b.height, lipgloss.Left, lipgloss.Bottom, appStyle.Width(b.width-2).Render(value))
}
