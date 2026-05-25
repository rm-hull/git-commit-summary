package ui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type promptViewModel struct {
	message   string
	textinput textinput.Model
}

func initialPromptViewModel(message, placeholder, hint string) *promptViewModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(hint)
	ti.Prompt = "❯ "
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(80)

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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.textinput.Blur()
			return m, func() tea.Msg { return userResponseMsg(m.textinput.Value()) }

		case "ctrl+c", "esc":
			m.textinput.Blur()
			return m, func() tea.Msg { return cancelRegenPromptMsg{} }
		}

	case errMsg:
		return m, tea.Quit
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m promptViewModel) View() tea.View {
	return tea.NewView(fmt.Sprintf(
		"%s\n%s\n",
		m.message,
		m.textinput.View(),
	))
}
