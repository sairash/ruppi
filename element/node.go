package element

import (
	"fmt"
	"rupi/config"
	"strings"

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
)

var (
	BoldStyle       = lipgloss.NewStyle().Bold(true)
	ItalicStyle     = lipgloss.NewStyle().Italic(true)
	CodeStyle       = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("229"))
	BlockquoteStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).PaddingLeft(1).MarginLeft(2)
	LinkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Underline(true)
	HrStyle         = lipgloss.NewStyle().Faint(true)

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

	case STYLE, SCRIPT, IFRAME:
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

func isBlockElement(nodeType uint) bool {
	switch nodeType {
	case H1, H2, H3, H4, H5, H6, P, DIV, UL, OL, LI, PRE, BLOCKQUOTE, HR, ROOT:
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
