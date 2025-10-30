package internal

import (
	"fmt"
	"strings"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TextArea(value string) (string, bool, error) {
	p := tea.NewProgram(initialModel(value))

	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}
	m := finalModel.(model)

	return m.Value(), m.Accepted(), nil
}

type model struct {
	textarea textarea.Model
	accepted bool
	err      error
}

func initialModel(value string) model {
	ti := textarea.New()
	ti.Focus()
	ti.SetHeight(strings.Count(value, "\n") + 1)
	ti.SetWidth(72)
	ti.SetValue(value)
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.Prompt = ""

	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle()

	return model{
		textarea: ti,
		accepted: false,
		err:      nil,
	}
}

func (m model) Value() string {
	return m.textarea.Value()
}

func (m model) Accepted() bool {
	return m.accepted
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.accepted = false
			m.textarea.Blur()
			return m, tea.Quit
		case tea.KeyCtrlX:
			m.accepted = true
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
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	box := box.New(box.Config{Px: 1, Py: 0, Type: "Round", Color: "Cyan", TitlePos: "Top"})
	view := box.String("Commit message", m.textarea.View())

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#FFD700", Dark: "#FFFF00"}).
		Bold(true)
	helpText := fmt.Sprintf("(%s to commit, %s to abort)",
		keyStyle.Render("Ctrl-X"),
		keyStyle.Render("ESC"))

	return view + helpText
}
