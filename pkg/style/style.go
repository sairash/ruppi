package style

import (
	"ruppi/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// Static styles that don't depend on theme
var (
	TitleStyle = lipgloss.NewStyle().Bold(true)
	PaddingX   = lipgloss.NewStyle().Padding(0, 1)
)

// Theme-aware style getters
func AppStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color(theme.BrowserBackground))
}

func StatusColor() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.StatusBarColor)).
		Foreground(lipgloss.Color(theme.BrowserForeground))
}

func StatusStyle() lipgloss.Style {
	return StatusColor().MarginBottom(1)
}

func TabContainerColor() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.BrowserBackground))
}

func LogoStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.TabActiveColor)).
		PaddingRight(1).
		PaddingLeft(1).
		Bold(true).
		Foreground(lipgloss.Color(theme.TabActiveTextColor))
}

func InspectorStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.InspectorBackground)).
		Foreground(lipgloss.Color(theme.InspectorForeground))
}

// DefaultStyle provides the base theme background for elements without specific styling
func DefaultStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.BrowserBackground))
}

// BodyStyle provides themed background for content areas
func BodyStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.BrowserBackground)).
		Foreground(lipgloss.Color(theme.BrowserForeground))
}

// ViewportStyle provides themed background for viewport content
func ViewportStyle() lipgloss.Style {
	theme := config.GetTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.BrowserBackground)).
		Foreground(lipgloss.Color(theme.BrowserForeground))
}
