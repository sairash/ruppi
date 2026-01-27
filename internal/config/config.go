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

var (
	ruppiConfig = make(StyleMap)
	maxGaps     = 3 // Default maximum consecutive gaps
)

func parseKeyValue(line string, info StyleInfo) (StyleInfo, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return info, fmt.Errorf("invalid key-value pair: %s", line)
	}
	key := strings.ToLower(parts[0])
	value := strings.Join(parts[1:], " ")

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
	value := parts[1]

	switch key {
	case "max_gaps":
		if gaps, err := strconv.Atoi(value); err == nil {
			if gaps < 1 {
				gaps = 1
			}
			if gaps > 10 {
				gaps = 10
			}
			maxGaps = gaps
		} else {
			return fmt.Errorf("invalid max_gaps value: %s", value)
		}
	default:
		return fmt.Errorf("unknown rupi setting: %s", key)
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
	if len(s) < 2 || s[0] != '#' {
		log.Fatalf("expected hex value like #ff0000, but got %s insted.", s)
	}
	for _, c := range s[1:] {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')) {
			log.Fatalf("expected hex value like #ff0000, but got %s insted.", s)
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

func GetMaxGaps() int {
	return maxGaps
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
			if currentTag != "" && currentTag != "rupi" {
				ruppiConfig[currentTag] = currentInfo
			}

			currentTag = strings.Trim(line, "[]")
			currentInfo = StyleInfo{Style: lipgloss.NewStyle()}
		} else if currentTag != "" {
			if currentTag == "rupi" {
				if err := parseRuppiSetting(line); err != nil {
					fmt.Printf("Warning: skipping rupi setting: %v\n", err)
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

	if currentTag != "" && currentTag != "rupi" {
		ruppiConfig[currentTag] = currentInfo
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}
