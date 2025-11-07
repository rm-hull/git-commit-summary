package ui

import (
	"fmt"
	"strings"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func commitMessage(value string) (string, Action, error) {
	p := tea.NewProgram(initialCommitMessageModel(value))

	finalModel, err := p.Run()
	if err != nil {
		return "", Abort, err
	}
	m := finalModel.(commitMessageModel)

	return m.textarea.Value(), m.action, nil
}

type commitMessageModel struct {
	textarea textarea.Model
	history  *History
	action   Action
	err      error
}

func initialCommitMessageModel(value string) commitMessageModel {
	ti := textarea.New()
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Focus()

	minHeight := 2
	messageLines := strings.Count(value, "\n") + 1
	if messageLines > minHeight {
		minHeight = messageLines
	}
	ti.SetHeight(minHeight)

	ti.SetWidth(73)
	ti.SetValue(value)
	if value == "" {
		ti.Placeholder = "Unable to provide a commit summary: staged files may be too large to\nbe summarized or were excluded from the visible diff."
	} else {
		ti.Placeholder = "Please supply a commit message."
	}

	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle()

	return commitMessageModel{
		textarea: ti,
		history:  NewHistory(value),
		action:   Abort,
		err:      nil,
	}
}

func (m commitMessageModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m commitMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Store the value before the update to track changes.
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

		case tea.KeyEsc:
			m.action = Abort
			m.textarea.Blur()
			return m, tea.Quit

		case tea.KeyCtrlX:
			m.action = Commit
			m.textarea.Blur()
			return m, tea.Quit

		case tea.KeyCtrlR:
			m.action = Regenerate
			m.textarea.Blur()
			return m, tea.Quit

		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	// After the update, check if the value has changed.
	newValue := m.textarea.Value()
	if oldValue != newValue {
		m.history.Add(newValue)
	}

	return m, tea.Batch(cmds...)
}

func (m commitMessageModel) View() string {
	box := box.New(box.Config{Px: 1, Py: 0, Type: "Round", Color: "Cyan", TitlePos: "Top"})
	view := box.String("Commit message", m.textarea.View())

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFD700", Dark: "#FFFF00"}).
		Bold(true)
	helpText := fmt.Sprintf("%s:commit  %s:clear  %s:undo  %s:redo  %s:regen  %s:abort",
		keyStyle.Render("CTRL-X"),
		keyStyle.Render("CTRL-K"),
		keyStyle.Render("CTRL-Z"),
		keyStyle.Render("CTRL-Y"),
		keyStyle.Render("CTRL-R"),
		keyStyle.Render("ESC"))

	return view + helpText
}
