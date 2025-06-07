package element

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The cast of characters in our DOM-to-terminal play.
const (
	H1 = iota
	H2
	H3
	H4
	H5
	H6
	DIV
	SPAN
	P
	TEXT
	BOLD
	ITALIC
	DOCUMENT
)

var (
	BoldStyle   = lipgloss.NewStyle().Bold(true)
	ItalicStyle = lipgloss.NewStyle().Italic(true)
)

type Node struct {
	Children  []Node
	Element   ElementData
	InnerText string
}

type ElementData struct {
	NodeType uint
	Attrs    map[string]string
}

type renderState struct {
	builder *strings.Builder
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

	var content string
	if len(n.Children) > 0 {
		var childrenBuilder strings.Builder

		childrenState := &renderState{builder: &childrenBuilder}

		for i, child := range n.Children {
			child.renderRecursive(childrenState, isKitty)

			if i < len(n.Children)-1 && !isBlockElement(child.Element.NodeType) && !isBlockElement(n.Children[i+1].Element.NodeType) {
				childrenBuilder.WriteString(" ")
			}
		}
		content = childrenBuilder.String()
	} else {
		content = n.InnerText
	}

	var finalOutput string
	switch n.Element.NodeType {
	case H1:
		finalOutput = BoldStyle.Render(setKittyFontSize(content, 4, isKitty))
	case H2:
		finalOutput = BoldStyle.Render(setKittyFontSize(content, 3, isKitty))
	case H3:
		finalOutput = BoldStyle.Render(setKittyFontSize(content, 2, isKitty))
	case H4, H5, H6, DIV, P:
		finalOutput = content
	case BOLD:
		finalOutput = BoldStyle.Render(content)
	case ITALIC:
		finalOutput = ItalicStyle.Render(content)
	case TEXT:
		finalOutput = content
	default:
		finalOutput = content
	}

	state.builder.WriteString(finalOutput)

	if !isBlock {
		return
	}

	if (n.Element.NodeType == H1 || n.Element.NodeType == H2) && isKitty {
		state.builder.WriteString("\n\n")
	} else {
		state.builder.WriteRune('\n')
	}
}

func isBlockElement(nodeType uint) bool {
	switch nodeType {
	case H1, H2, H3, H4, H5, H6, DIV, P:
		return true
	default:
		return false
	}
}

func setKittyFontSize(content string, size int, isKitty bool) string {
	if !isKitty || size <= 1 {
		return content
	}
	if size > 4 {
		size = 4
	}

	return fmt.Sprintf("\x1b]66;s=%d;%s\x07", size, content)
}
