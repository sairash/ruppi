package style

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle     = lipgloss.NewStyle().Bold(true)
	AppStyle       = lipgloss.NewStyle().Padding(0, 1)
	BorderTopStyle = lipgloss.NewStyle().Background(lipgloss.Color("29"))
	BodyStyle      = lipgloss.NewStyle()
	PaddingX       = lipgloss.NewStyle().Padding(0, 1)

	LogoStyle   = lipgloss.NewStyle().Background(lipgloss.Color("200")).PaddingRight(1).PaddingLeft(1).Bold(true)
	StatusColor = lipgloss.NewStyle().Background(lipgloss.Color("#242424")).Foreground(lipgloss.Color("#7D7D7D"))

	TabContainerColor        = lipgloss.NewStyle().Background(lipgloss.Color("#0e0e0e"))
	TabBackgroundColor       = lipgloss.NewStyle().Background(lipgloss.Color("#202020"))
	TabActiveBackgroundColor = lipgloss.NewStyle().Background(lipgloss.Color("#3a3a3a"))

	StatusStyle = StatusColor.MarginBottom(1)
)
