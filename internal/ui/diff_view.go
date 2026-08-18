package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type diffViewModel struct {
	viewport       viewport.Model
	boxStyle       lipgloss.Style
	rawDiff        string
	compactSummary bool
}

func initialDiffViewModel(width, height int) *diffViewModel {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	vp.SoftWrap = false

	return &diffViewModel{
		viewport: vp,
		boxStyle: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("6")). // Cyan
			Padding(0, 1),
	}
}

func (m *diffViewModel) Init() tea.Cmd {
	return nil
}

func (m *diffViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return cancelDiffViewMsg{} }
		case "ctrl+d":
			if m.compactSummary {
				return m, func() tea.Msg { return showDiffViewMsg{} }
			}
			return m, func() tea.Msg { return showDiffSummaryViewMsg{} }
		}
	case showDiffViewMsg:
		m.compactSummary = false
		m.viewport.SetContent(m.rawDiff)
	case diffColorMsg:
		m.rawDiff = string(msg)
		m.compactSummary = false
		m.viewport.SetContent(string(msg))
	case diffSummaryMsg:
		m.compactSummary = true
		m.viewport.SetContent(string(msg))
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *diffViewModel) View() tea.View {
	title := " Raw diff "
	if m.compactSummary {
		title = " Compact summary "
	}
	titleBorder := lipgloss.RoundedBorder()

	repeatCount := max((m.viewport.Width()+2)-lipgloss.Width(title), 0)
	titleBorder.Top = title + strings.Repeat("─", repeatCount)

	toggleLabel := "list"
	if m.compactSummary {
		toggleLabel = "diff"
	}
	helpText := fmt.Sprintf("%s:commit %s:clear %s:regen %s:editor  %s:%s %s:back",
		Strikethrough.Render("CTRL+A"),
		Strikethrough.Render("CTRL+K"),
		Strikethrough.Render("CTRL+R"),
		Strikethrough.Render("CTRL+P"),
		BoldYellow.Render("CTRL+D"),
		toggleLabel,
		BoldYellow.Render("ESC"))

	return tea.NewView(m.boxStyle.
		BorderStyle(titleBorder).
		Render(m.viewport.View()) + "\n" + helpText)
}
