package help

import (
	"os"
	"strconv"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"github.com/charmbracelet/x/term"
)

func TerminalWidth() int {
	if w, _, err := term.GetSize(uintptr(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if w, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && w > 0 {
		return w
	}
	return 72
}

func RenderMarkdownHelp(markdown string, terminalWidth int) string {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}

	customStyle := styles.DarkStyleConfig
	customStyle.Document.Margin = uintPtr(0)
	customStyle.Paragraph.Margin = uintPtr(2)
	customStyle.List.Margin = uintPtr(2)
	customStyle.CodeBlock.Margin = uintPtr(6)
	customStyle.H2.BlockSuffix = ""

	renderer, err := glamour.NewTermRenderer(
		glamour.WithPreservedNewLines(),
		glamour.WithStyles(customStyle),
		glamour.WithWordWrap(terminalWidth),
	)
	if err != nil {
		return markdown
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}

	return strings.TrimSpace(rendered)
}

func uintPtr(v uint) *uint { return new(v) }
