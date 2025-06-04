package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	appStyle     = lipgloss.NewStyle().Padding(0, 1)
	testingStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false)
)

type model struct {
	width   int
	height  int
	isKitty bool
}

type tickMsg time.Time

func (m model) block(value string, scale int) string {
	if !m.isKitty || scale <= 1 {
		return value + "\n"
	}
	amount := scale - 1
	if scale == 4 {
		amount = 2
	}

	return fmt.Sprintf("\x1b]66;s=%d;%s\x07%s", scale, value, strings.Repeat("\n", amount))
}

func main() {
	termProgram := os.Getenv("TERM")
	width, height, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	p := tea.NewProgram(model{
		width:   width,
		height:  height,
		isKitty: strings.Contains(termProgram, "kitty"),
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		return m, tick()
	}

	return m, nil
}

func (m model) View() string {
	value := fmt.Sprintf("%sA simple terminal Web browser written in golang! \n%s\n%s%s%s%s",
		m.block("- Rupi 🐦 -", 3),
		testingStyle.Width(m.width-4).Render("Testing:"),
		m.block("H1", 4),
		m.block("H2", 3),
		m.block("H3", 2),
		m.block("H4, H5, H6", 1),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, appStyle.Width(m.width-2).Height(m.height-15).Render(value))
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
