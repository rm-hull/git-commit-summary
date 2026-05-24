package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type diffColorMsg string

type diffViewModel struct {
	viewport viewport.Model
}

func initialDiffViewModel(width, height int) *diffViewModel {
	return &diffViewModel{
		viewport: viewport.New(width, height),
	}
}

func (m *diffViewModel) Init() tea.Cmd {
	return nil
}

func (m *diffViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diffColorMsg:
		m.viewport.SetContent(string(msg))
		return m, nil
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *diffViewModel) View() string {
	boxStyle := lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("6")). // Cyan
		Padding(0, 1)

	title := " Raw diff "
	titleBorder := lipgloss.RoundedBorder()

	// Ensure we don't have a negative repeat count
	repeatCount := (m.viewport.Width + 2) - lipgloss.Width(title)
	if repeatCount < 0 {
		repeatCount = 0
	}
	titleBorder.Top = title + strings.Repeat("─", repeatCount)

	helpText := fmt.Sprintf("%s:commit %s:regen %s:preview %s:diff  %s:back",
		Strikethrough.Render("CTRL+A"),
		Strikethrough.Render("CTRL+R"),
		Strikethrough.Render("CTRL+P"),
		BoldYellow.Render("CTRL+D"),
		BoldYellow.Render("ESC"))

	return boxStyle.
		BorderStyle(titleBorder).
		Render(m.viewport.View()) + "\n" + helpText
}
