package main

import (
	"fmt"
	"log"
	"os"
	"rupi/element"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	appStyle       = lipgloss.NewStyle().Padding(0, 1)
	BorderTopStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false)
	bodyStyle      = lipgloss.NewStyle()
)

type Browser struct {
	width    int
	height   int
	isKitty  bool
	url      textinput.Model
	elements []element.Node
}

func main() {
	termProgram := os.Getenv("TERM")
	width, height, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}

	ti := textinput.New()
	ti.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	ti.Placeholder = "Search DuckDuckGo or type Url"
	ti.Width = width - 10
	ti.Focus()
	// ti.Width
	ti.Prompt = "Url: "

	elements := []element.Node{
		{
			Element: element.ElementData{
				NodeType: element.H1,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H1",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
			},
		},
		{
			Element: element.ElementData{
				NodeType: element.H2,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H2",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
			},
		},
		{
			Element: element.ElementData{
				NodeType: element.H3,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H3",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
			},
		},
		{
			Element: element.ElementData{
				NodeType: element.H4,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H4",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
			},
		},
		{
			Element: element.ElementData{
				NodeType: element.H5,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H5",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
			},
		},
		{
			InnerText: "H6",
			Element: element.ElementData{
				NodeType: element.H6,
				Attrs:    map[string]string{},
			},
			Children: []element.Node{
				{
					InnerText: "H6",
					Element: element.ElementData{
						NodeType: element.TEXT,
					},
				},
				{
					InnerText: "Italic",
					Element: element.ElementData{
						NodeType: element.ITALIC,
					},
					Children: []element.Node{
						{
							InnerText: "Itallic",
							Element: element.ElementData{
								NodeType: element.TEXT,
							},
						},
						{
							Element: element.ElementData{
								NodeType: element.BOLD,
							},
							Children: []element.Node{
								{
									Element: element.ElementData{
										NodeType: element.TEXT,
									},
									InnerText: "Bold and Italic",
								},
							},
						},
					},
				},
			},
		},
	}

	p := tea.NewProgram(Browser{
		width:    width,
		height:   height,
		url:      ti,
		isKitty:  strings.Contains(termProgram, "kitty"),
		elements: elements,
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
	x := ""

	for _, element := range b.elements {
		x += element.Render(b.isKitty)
	}

	value := fmt.Sprintf("%s\n%s", bodyStyle.Width(b.width-4).Height(b.height-9).Render(x), BorderTopStyle.Width(b.width-4).Render(b.url.View()))
	return lipgloss.Place(b.width, b.height, lipgloss.Left, lipgloss.Bottom, appStyle.Width(b.width-2).Render(value))
}
