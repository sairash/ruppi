package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StyleInfo struct {
	Style  lipgloss.Style
	Prefix string
	Infix  string
}

type StyleMap map[string]StyleInfo

type Theme struct {
	// Colors
	TabColor            string
	TabActiveColor      string
	TabTextColor        string
	TabActiveTextColor  string
	BackgroundColor     string
	StatusBarColor      string
	BrowserForeground   string
	InspectorForeground string

	// Icons
	TabCloseIcon string
	TabNewIcon   string
	TabPrevIcon  string
	TabNextIcon  string
	SearchIcon   string

	// Text Labels
	SearchPlaceholder  string
	InspectorToggleKey string
	QuitKey            string
	SearchKey          string
	NewTabTooltip      string

	// Browser
	BrowserBackground   string
	InspectorBackground string
}

// SixelConfig holds sixel image settings
type SixelConfig struct {
	Enabled   bool
	MaxWidth  int
	MaxHeight int
}

var (
	ruppiConfig  = make(StyleMap)
	maxGaps      = 3
	currentTheme = getDefaultTheme()
	sixelConfig  = SixelConfig{Enabled: true, MaxWidth: 400, MaxHeight: 300}
)

// GetSixelConfig returns the sixel configuration
func GetSixelConfig() SixelConfig {
	return sixelConfig
}

func parseKeyValue(line string, info StyleInfo) (StyleInfo, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return info, fmt.Errorf("invalid key-value pair: %s", line)
	}
	key := strings.ToLower(parts[0])
	value := strings.Join(parts[1:], " ")

	// Remove quotes if present
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	switch key {
	case "prefix":
		if unquoted, err := strconv.Unquote(value); err == nil {
			info.Prefix = unquoted
		} else {
			info.Prefix = value
		}

	case "infix":
		if unquoted, err := strconv.Unquote(value); err == nil {
			info.Infix = unquoted
		} else {
			info.Infix = value
		}

	case "bold":
		info.Style = info.Style.Bold(parseBool(value))
	case "italic":
		info.Style = info.Style.Italic(parseBool(value))
	case "faint":
		info.Style = info.Style.Faint(parseBool(value))
	case "underline":
		info.Style = info.Style.Underline(parseBool(value))
	case "strikethrough":
		info.Style = info.Style.Strikethrough(parseBool(value))
	case "border-left":
		info.Style = info.Style.BorderLeft(parseBool(value))
	case "border-right":
		info.Style = info.Style.BorderRight(parseBool(value))
	case "border-top":
		info.Style = info.Style.BorderTop(parseBool(value))
	case "border-bottom":
		info.Style = info.Style.BorderBottom(parseBool(value))

	case "foreground":
		info.Style = info.Style.Foreground(lipgloss.Color(parseHex(value)))
	case "background":
		info.Style = info.Style.Background(lipgloss.Color(parseHex(value)))
	case "border-color":
		info.Style = info.Style.BorderForeground(lipgloss.Color(parseHex(value)))

	case "margin-left":
		info.Style = info.Style.MarginLeft(parseInt(value))
	case "margin-right":
		info.Style = info.Style.MarginRight(parseInt(value))
	case "margin-top":
		info.Style = info.Style.MarginTop(parseInt(value))
	case "margin-bottom":
		info.Style = info.Style.MarginBottom(parseInt(value))
	case "padding-left":
		info.Style = info.Style.PaddingLeft(parseInt(value))
	case "padding-right":
		info.Style = info.Style.PaddingRight(parseInt(value))
	case "padding-top":
		info.Style = info.Style.PaddingTop(parseInt(value))
	case "padding-bottom":
		info.Style = info.Style.PaddingBottom(parseInt(value))

	case "align":
		switch strings.ToLower(value) {
		case "left":
			info.Style = info.Style.Align(lipgloss.Left)
		case "center":
			info.Style = info.Style.Align(lipgloss.Center)
		case "right":
			info.Style = info.Style.Align(lipgloss.Right)
		}
	case "border-type":
		switch strings.ToLower(value) {
		case "normal":
			info.Style = info.Style.Border(lipgloss.NormalBorder())
		case "rounded":
			info.Style = info.Style.Border(lipgloss.RoundedBorder())
		case "thick":
			info.Style = info.Style.Border(lipgloss.ThickBorder())
		case "double":
			info.Style = info.Style.Border(lipgloss.DoubleBorder())
		case "hidden":
			info.Style = info.Style.Border(lipgloss.HiddenBorder())
		}
	default:
		return info, fmt.Errorf("unknown style key: %s", key)
	}

	return info, nil
}

func parseRuppiSetting(line string) error {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return fmt.Errorf("invalid setting: %s", line)
	}

	key := strings.ToLower(parts[0])
	value := strings.Join(parts[1:], " ")

	// Remove quotes if present
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	switch key {
	case "max-gaps":
		if gaps, err := strconv.Atoi(value); err == nil {
			if gaps < 1 {
				gaps = 1
			}
			if gaps > 10 {
				gaps = 10
			}
			maxGaps = gaps
		} else {
			return fmt.Errorf("invalid max-gaps value: %s", value)
		}

	// Colors
	case "tab-color":
		currentTheme.TabColor = value
	case "tab-active-color":
		currentTheme.TabActiveColor = value
	case "tab-text-color":
		currentTheme.TabTextColor = value
	case "tab-active-text-color":
		currentTheme.TabActiveTextColor = value
	case "background-color":
		currentTheme.BackgroundColor = value
	case "status-bar-color":
		currentTheme.StatusBarColor = value
	case "browser-background":
		currentTheme.BrowserBackground = value
	case "inspector-background":
		currentTheme.InspectorBackground = value
	case "browser-foreground":
		currentTheme.BrowserForeground = value
	case "inspector-foreground":
		currentTheme.InspectorForeground = value

	// Icons
	case "tab-close-icon":
		currentTheme.TabCloseIcon = value
	case "tab-new-icon":
		currentTheme.TabNewIcon = value
	case "tab-prev-icon":
		currentTheme.TabPrevIcon = value
	case "tab-next-icon":
		currentTheme.TabNextIcon = value
	case "search-icon":
		currentTheme.SearchIcon = value

	// Text Labels
	case "search-placeholder":
		currentTheme.SearchPlaceholder = value
	case "inspector-toggle-key":
		currentTheme.InspectorToggleKey = value
	case "quit-key":
		currentTheme.QuitKey = value
	case "search-key":
		currentTheme.SearchKey = value
	case "new-tab-tooltip":
		currentTheme.NewTabTooltip = value

	// Sixel Configuration
	case "enable-sixel":
		sixelConfig.Enabled = parseBool(value)
	case "sixel-max-width":
		if w, err := strconv.Atoi(value); err == nil && w > 0 {
			sixelConfig.MaxWidth = w
		}
	case "sixel-max-height":
		if h, err := strconv.Atoi(value); err == nil && h > 0 {
			sixelConfig.MaxHeight = h
		}

	default:
		return fmt.Errorf("unknown ruppi setting: %s", key)
	}

	return nil
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseHex(s string) string {
	if len(s) < 4 || s[0] != '#' {
		log.Fatalf("expected hex value like #ff0000, but got %s instead.", s)
	}
	// Support both #RGB and #RRGGBB formats
	hexPart := s[1:]
	if len(hexPart) != 3 && len(hexPart) != 6 {
		log.Fatalf("expected hex value like #ff0000 or #f00, but got %s instead.", s)
	}
	for _, c := range hexPart {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')) {
			log.Fatalf("expected hex value like #ff0000, but got %s instead.", s)
		}
	}
	return s
}

func AddStyle(tag string, content string) string {
	if strings.TrimSpace(content) == "" {
		return content
	}

	if val, has := ruppiConfig[tag]; has {
		return val.Style.Render(val.Prefix + content + val.Infix)
	}

	return content
}

func getDefaultTheme() Theme {
	return Theme{
		// Colors
		TabColor:            "#4a5568",
		TabActiveColor:      "#48bb78",
		TabTextColor:        "#ffffff",
		TabActiveTextColor:  "#ffffff",
		BackgroundColor:     "#1a1a1a",
		StatusBarColor:      "#242424",
		BrowserForeground:   "#ffffff",
		InspectorForeground: "#ffffff",

		// Icons
		TabCloseIcon: "×",
		TabNewIcon:   "＋",
		TabPrevIcon:  "◀",
		TabNextIcon:  "▶",
		SearchIcon:   "🔗",

		// Text Labels
		SearchPlaceholder:  "Search or type a URL",
		InspectorToggleKey: "?",
		QuitKey:            "q",
		SearchKey:          "i",
		NewTabTooltip:      "New Tab",

		// Browser
		BrowserBackground:   "#0e0e0e",
		InspectorBackground: "#1e1e1e",
	}
}

func GetMaxGaps() int {
	return maxGaps
}

func GetTheme() Theme {
	return currentTheme
}

func GetTabColor() string {
	return currentTheme.TabColor
}

func GetTabActiveColor() string {
	return currentTheme.TabActiveColor
}

func GetTabTextColor() string {
	return currentTheme.TabTextColor
}

func GetTabActiveTextColor() string {
	return currentTheme.TabActiveTextColor
}

func GetBackgroundColor() string {
	return currentTheme.BackgroundColor
}

func GetStatusBarColor() string {
	return currentTheme.StatusBarColor
}

func GetBrowserBackground() string {
	return currentTheme.BrowserBackground
}

func GetInspectorBackground() string {
	return currentTheme.InspectorBackground
}

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentTag string
	var currentInfo StyleInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentTag != "" && currentTag != "ruppi" {
				ruppiConfig[currentTag] = currentInfo
			}

			currentTag = strings.Trim(line, "[]")
			currentInfo = StyleInfo{Style: lipgloss.NewStyle()}
		} else if currentTag != "" {
			if currentTag == "ruppi" {
				if err := parseRuppiSetting(line); err != nil {
					fmt.Printf("Warning: skipping ruppi setting: %v\n", err)
				}
			} else {
				var err error
				currentInfo, err = parseKeyValue(line, currentInfo)
				if err != nil {
					fmt.Printf("Warning: skipping line in [%s]: %v\n", currentTag, err)
				}
			}
		}
	}

	if currentTag != "" && currentTag != "ruppi" {
		ruppiConfig[currentTag] = currentInfo
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}
