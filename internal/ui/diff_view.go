package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type diffColorMsg string

type diffViewModel struct {
	viewport viewport.Model
	boxStyle lipgloss.Style
}

func initialDiffViewModel(width, height int) *diffViewModel {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	vp.SoftWrap = false
	vp.MouseWheelEnabled = true

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
	case diffColorMsg:
		m.viewport.SetContent(string(msg))
		return m, nil
	case tea.WindowSizeMsg:
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height)
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *diffViewModel) View() tea.View {
	title := " Raw diff "
	titleBorder := lipgloss.RoundedBorder()

	repeatCount := max((m.viewport.Width()+2)-lipgloss.Width(title), 0)
	titleBorder.Top = title + strings.Repeat("─", repeatCount)

	helpText := fmt.Sprintf("%s:commit %s:clear %s:regen %s:editor  %s:diff %s:back",
		Strikethrough.Render("CTRL+A"),
		Strikethrough.Render("CTRL+K"),
		Strikethrough.Render("CTRL+R"),
		Strikethrough.Render("CTRL+P"),
		BoldYellow.Render("CTRL+D"),
		BoldYellow.Render("ESC"))

	return tea.NewView(m.boxStyle.
		BorderStyle(titleBorder).
		Render(m.viewport.View()) + "\n" + helpText)
}
