package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type commitViewModel struct {
	textarea textarea.Model
	history  *History
	boxStyle lipgloss.Style
}

func initialCommitViewModel(message string) *commitViewModel {
	ti := textarea.New()
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Focus()

	minHeight := 2
	messageLines := strings.Count(message, "\n") + 1
	if messageLines > minHeight {
		minHeight = messageLines
	}
	ti.SetHeight(minHeight)

	ti.SetWidth(74)
	ti.SetValue(message)
	if message == "" {
		ti.Placeholder = "Unable to provide a commit summary: staged files may be too large to\nbe summarized or were excluded from the visible diff."
	} else {
		ti.Placeholder = "Please supply a commit message."
	}

	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()

	title := " Commit message "
	titleBorder := lipgloss.RoundedBorder()
	titleBorder.Top = title + strings.Repeat(
		"─", ti.Width()-lipgloss.Width(title)+2) // +2 is to accommodate for horizontal padding

	return &commitViewModel{
		textarea: ti,
		history:  NewHistory(message),
		boxStyle: lipgloss.NewStyle().
			BorderStyle(titleBorder).
			BorderForeground(lipgloss.Color("6")). // Cyan
			Padding(0, 1),
	}
}

func (m *commitViewModel) Init() tea.Cmd {
	m.textarea.Focus()
	return textarea.Blink
}

func (m *commitViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	oldValue := m.textarea.Value()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlZ:
			if value, ok := m.history.Undo(); ok {
				m.textarea.SetValue(value)
			}
			return m, nil

		case tea.KeyCtrlY:
			if value, ok := m.history.Redo(); ok {
				m.textarea.SetValue(value)
			}
			return m, nil

		case tea.KeyCtrlK:
			if m.textarea.Value() == "" {
				return m, nil
			}
			m.history.Add("")
			m.textarea.SetValue("")
			return m, nil

		case tea.KeyEsc, tea.KeyCtrlC:
			m.textarea.Blur()
			return m, func() tea.Msg { return abortMsg{} }

		case tea.KeyCtrlX:
			m.textarea.Blur()
			return m, func() tea.Msg { return commitMsg(m.textarea.Value()) }

		case tea.KeyCtrlR:
			m.textarea.Blur()
			return m, func() tea.Msg { return regenerateMsg{} }

		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	case errMsg: // Use errMsg from model.go
		return m, tea.Quit
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	newValue := m.textarea.Value()
	if oldValue != newValue {
		m.history.Add(newValue)
	}

	return m, tea.Batch(cmds...)
}

func (m *commitViewModel) View() string {
	return m.boxStyle.Render(m.textarea.View()) + "\n"
}
