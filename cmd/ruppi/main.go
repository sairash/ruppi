package main

import (
	"flag"
	"log"
	"os"
	"ruppi/internal/app"
	"ruppi/internal/config"
	"ruppi/internal/logger"
	"ruppi/pkg/style"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"golang.org/x/term"
)

func main() {
	if err := config.LoadConfig("./config/ruppi.conf.example"); err != nil {
		log.Printf("Could not load config file: %v. Using default values.", err)
	}

	urlFlag := flag.String("url", "", "The URL to parse and render.")
	kittyFlag := flag.Bool("kitty", true, "Enable Kitty terminal graphics protocol extensions.")
	contentWidth := flag.Int("width", 80, "Content word wrap width. Default is 80.")
	flag.Parse()

	zone.NewGlobal()
	defer zone.Close()

	termProgram := os.Getenv("TERM")
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("could not get terminal size: %v", err)
	}

	browserModel := NewBrowser(width, height, *contentWidth, *kittyFlag || strings.Contains(termProgram, "kitty"))

	browserModel.Url.SetValue(*urlFlag)

	browserModel.Tabs.NewTab(*urlFlag, browserModel.WordWrap(), browserModel.IsKitty)

	p := tea.NewProgram(
		browserModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}

func NewBrowser(width, height, contentWidth int, isKitty bool) app.Browser {
	logger := logger.NewLogger()

	ti := textinput.New()
	ti.PlaceholderStyle = style.StatusColor
	ti.TextStyle = style.StatusColor
	ti.Cursor.Style = style.StatusColor
	ti.PromptStyle = style.StatusColor
	ti.Cursor.TextStyle = style.StatusColor
	ti.Placeholder = "Search or type a URL"
	ti.Blur()
	ti.Prompt = "🔗 > "
	ti.CharLimit = 256
	ti.Width = width - 28 // Initial width, will be updated on window resize

	return app.Browser{
		Width:        width,
		Height:       height,
		ContentWidth: contentWidth,
		Url:          ti,
		Tabs: &app.Tabs{
			Tabs: []*app.Tab{},
		},
		IsInspectorOpen: false,
		IsKitty:         isKitty,
		Ready:           false,
		ActivePane:      app.ACTIVE_VIEWPORT,
		Logger:          logger,
	}
}
