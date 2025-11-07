package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gookit/color"
)

func prompt(message, placeholder string) (string, Action, error) {
	p := tea.NewProgram(initialPromptModel(message, placeholder))
	finalModel, err := p.Run()
	if err != nil {
		return "", Abort, err
	}
	m := finalModel.(promptModel)
	return m.textInput.Value(), m.action, nil
}

type promptModel struct {
	message   string
	textInput textinput.Model
	action    Action
	err       error
}

func initialPromptModel(prompt, placeholder string) promptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "❯ "
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	return promptModel{
		message:   prompt,
		textInput: ti,
		action:    Abort,
		err:       nil,
	}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.action = Ok
			m.textInput.Blur()
			return m, tea.Quit

		case tea.KeyCtrlC, tea.KeyEsc:
			m.action = Abort
			m.textInput.Blur()
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	return color.Sprintf(
		"%s\n%s",
		m.message,
		m.textInput.View(),
	) + "\n"
}
