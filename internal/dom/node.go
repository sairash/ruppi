package dom

import (
	"fmt"
	"rupi/internal/config"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

const (
	TEXT = iota
	ROOT

	H1
	H2
	H3
	H4
	H5
	H6
	P
	DIV
	UL
	OL
	LI
	PRE
	BLOCKQUOTE
	BR
	HR

	SPAN
	A
	BOLD
	STRONG
	ITALIC
	EM
	CODE
	IMG

	STYLE
	SCRIPT
	IFRAME

	INPUT
)

var (
	BoldStyle       = lipgloss.NewStyle().Bold(true)
	ItalicStyle     = lipgloss.NewStyle().Italic(true)
	BlockquoteStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).PaddingLeft(1).MarginLeft(2)
	LinkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Underline(true)
	HrStyle         = lipgloss.NewStyle().Faint(true)

	InputStyle           = lipgloss.NewStyle().Background(lipgloss.Color("#242424")).Faint(true)
	InputBackgroundStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#242424"))

	NoStyle = lipgloss.NewStyle()

	AllStyles = map[string]lipgloss.Style{}
)

var TagToType = map[string]uint{
	"h1":         H1,
	"h2":         H2,
	"h3":         H3,
	"h4":         H4,
	"h5":         H5,
	"h6":         H6,
	"p":          P,
	"div":        DIV,
	"ul":         UL,
	"ol":         OL,
	"li":         LI,
	"pre":        PRE,
	"blockquote": BLOCKQUOTE,
	"br":         BR,
	"hr":         HR,
	"span":       SPAN,
	"a":          A,
	"b":          BOLD,
	"strong":     STRONG,
	"i":          ITALIC,
	"em":         EM,
	"code":       CODE,
	"img":        IMG,
	"style":      STYLE,
	"script":     SCRIPT,
	"iframe":     IFRAME,
	"input":      INPUT,
}

type Node struct {
	Children  []Node
	Element   ElementData
	InnerText string
}
type ElementData struct {
	NodeType uint
	Name     string
	Attrs    map[string]string
}

type renderState struct {
	builder   *strings.Builder
	listIndex int
}

func (s *renderState) ensureNewline() {
	if s.builder.Len() > 0 && s.builder.String()[s.builder.Len()-1] != '\n' {
		s.builder.WriteRune('\n')
	}
}

func (n *Node) Render(isKitty bool) string {
	var sb strings.Builder
	state := &renderState{builder: &sb}
	n.renderRecursive(state, isKitty)
	return sb.String()
}

func (n *Node) renderRecursive(state *renderState, isKitty bool) {
	isBlock := isBlockElement(n.Element.NodeType)

	if isBlock {
		state.ensureNewline()
	}

	isList := (n.Element.NodeType == UL || n.Element.NodeType == OL)
	if isList {
		state.listIndex = 0
	}

	var content string
	if len(n.Children) > 0 {
		var childrenBuilder strings.Builder
		childrenState := &renderState{builder: &childrenBuilder, listIndex: state.listIndex}
		for i, child := range n.Children {
			child.renderRecursive(childrenState, isKitty)

			if i < len(n.Children)-1 && !isBlockElement(child.Element.NodeType) && !isBlockElement(n.Children[i+1].Element.NodeType) {
				childrenBuilder.WriteString(" ")
			}
		}
		state.listIndex = childrenState.listIndex
		content = childrenBuilder.String()
	} else {
		content = n.InnerText
	}

	var finalOutput string
	switch n.Element.NodeType {
	case LI:
		if state.listIndex > 0 {
			finalOutput = fmt.Sprintf("  %d. %s", state.listIndex, content)
		} else {
			finalOutput = "  • " + content
		}
	case OL:
		state.listIndex = 1
		finalOutput = content
	case UL:
		state.listIndex = 0
		finalOutput = content

	case A:
		href, ok := n.Element.Attrs["href"]
		if ok {
			finalOutput = fmt.Sprintf("%s %s", content, config.AddStyle("a", href))
		} else {
			finalOutput = content
		}
	case IMG:
		alt := n.Element.Attrs["alt"]
		finalOutput = ItalicStyle.Render(fmt.Sprintf("[Image: %s]", alt))

	case BLOCKQUOTE:
		finalOutput = BlockquoteStyle.Render(content)
	case PRE:
		finalOutput = lipgloss.NewStyle().Margin(1, 2).Render(content)
	case HR:
		finalOutput = HrStyle.Render(strings.Repeat("─", 50))
	case BR:
		finalOutput = HrStyle.Render(" ")
	case STYLE, SCRIPT, IFRAME:

	case INPUT:
		// TODO: Make this input system better
		finalOutput = InputBackgroundStyle.Render("█") + InputStyle.Render(n.Element.Attrs["placeholder"]) + InputBackgroundStyle.Render(strings.Repeat("█", 10))
	default:
		finalOutput = config.AddStyle(n.Element.Name, content)
	}

	state.builder.WriteString(finalOutput)

	if isBlock {

		if n.Element.NodeType != LI {
			state.builder.WriteRune('\n')
		}

		if n.Element.NodeType == LI && state.listIndex > 0 {
			state.listIndex++
		}
	}
}

func stripANSICodes(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func normalizeNewlines(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	newlineCount := 0
	for _, r := range s {
		if r == '\n' {
			newlineCount++
			if newlineCount <= 2 {
				b.WriteRune(r)
			}
		} else {
			newlineCount = 0
			b.WriteRune(r)
		}
	}
	return b.String()
}

func WordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	text = normalizeNewlines(text)

	var wrappedText strings.Builder
	wrappedText.Grow(len(text) + len(text)/maxWidth*2)

	currentLineLength := 0
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})

	remainingText := text
	for _, word := range words {
		idx := strings.Index(remainingText, word)
		if idx == -1 {
			continue
		}

		segment := remainingText[:idx+len(word)]

		remainingText = remainingText[idx+len(word):]

		actualWord := word
		leadingWhitespace := ""
		if idx > 0 {
			leadingWhitespace = segment[:idx]
		}

		displayWord := stripANSICodes(actualWord)
		wordDisplayLength := utf8.RuneCountInString(displayWord)

		newlineInSegment := strings.ContainsRune(leadingWhitespace, '\n')
		if newlineInSegment {
			if wrappedText.Len() > 0 && wrappedText.String()[wrappedText.Len()-1] != '\n' {
				wrappedText.WriteRune('\n')
			}
			currentLineLength = 0
		} else if currentLineLength > 0 && currentLineLength+1+wordDisplayLength > maxWidth {
			wrappedText.WriteRune('\n')
			currentLineLength = 0
		} else if currentLineLength > 0 {
			wrappedText.WriteRune(' ')
			currentLineLength++
		}

		wrappedText.WriteString(actualWord)
		currentLineLength += wordDisplayLength
	}

	return wrappedText.String()
}

func isBlockElement(nodeType uint) bool {
	switch nodeType {
	case H1, H2, H3, H4, H5, H6, P, DIV, UL, OL, LI, PRE, BLOCKQUOTE, HR, ROOT, BR:
		return true
	default:
		return false
	}
}

// This is used to set different text size in kitty terminal. Really hard to work with.
// Have to fix it later.

// func setKittyFontSize(content string, size int, isKitty bool) string {

// 	// make the text rendering correct later.
// 	return content

// 	if !isKitty || size <= 1 {
// 		return content
// 	}
// 	if size > 4 {
// 		size = 4
// 	}

// 	return fmt.Sprintf("\x1b]66;s=%d;%s\x07", size, content)
// }
