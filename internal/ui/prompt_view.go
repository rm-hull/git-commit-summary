package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type promptViewModel struct {
	message   string
	textinput textinput.Model
}

func initialPromptViewModel(message, placeholder string) *promptViewModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "❯ "
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	return &promptViewModel{
		message:   message,
		textinput: ti,
	}
}

func (m promptViewModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.textinput.Blur()
			return m, func() tea.Msg { return userResponseMsg(m.textinput.Value()) }

		case tea.KeyCtrlC, tea.KeyEsc:
			m.textinput.Blur()
			return m, func() tea.Msg { return cancelRegenPromptMsg{} }
		}

	case errMsg:
		return m, tea.Quit
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m promptViewModel) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		m.message,
		m.textinput.View(),
	) + "\n"
}
