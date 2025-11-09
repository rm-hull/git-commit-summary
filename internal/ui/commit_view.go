package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

type commitViewModel struct {
	textarea textarea.Model
	viewport viewport.Model
	history  *History
	boxStyle lipgloss.Style
	preview  bool
	helpText bool
	renderer *glamour.TermRenderer
}

func initialCommitViewModel(message string) *commitViewModel {
	ti := textarea.New()
	ti.CharLimit = 0
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Focus()

	height := 2
	messageLines := strings.Count(message, "\n") + 1
	if height < messageLines {
		height = messageLines
	}
	if height > 15 {
		height = 15
	}
	ti.SetHeight(minHeight)
	ti.SetWidth(72 + 2) // +2 is to accommodate for horizontal padding
	ti.SetValue(message)
	if message == "" {
		ti.Placeholder = "Unable to provide a commit summary: staged files may be too large to\nbe summarized or were excluded from the visible diff."
	} else {
		ti.Placeholder = "Please supply a commit message."
	}

	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(ti.Width(), ti.Height())

	auto := styles.DarkStyleConfig
	auto.Document.Margin = uintPtr(0)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithPreservedNewLines(),
		glamour.WithStyles(auto),
	)

	return &commitViewModel{
		textarea: ti,
		viewport: vp,
		history:  NewHistory(message),
		boxStyle: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("6")). // Cyan
			Padding(0, 1),
		preview:  false,
		helpText: true,
		renderer: renderer,
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
		if m.preview {
			switch msg.Type {
			case tea.KeyCtrlP, tea.KeyEsc:
				m.preview = false
				m.textarea.Focus()

			case tea.KeyCtrlC:
				m.helpText = false
				m.textarea.Blur()
				return m, func() tea.Msg { return abortMsg{} }

			default:
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.Type {
		case tea.KeyCtrlP:
			m.preview = true
			m.textarea.Blur()
			out, _ := m.renderer.Render(m.textarea.Value())
			m.viewport.SetContent(strings.TrimSpace(out))
			return m, nil

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
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return abortMsg{} }

		case tea.KeyCtrlX:
			m.textarea.Blur()
			return m, func() tea.Msg { return commitMsg(m.textarea.Value()) }

		case tea.KeyCtrlR:
			m.helpText = false
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
	var view string
	var title string

	if m.preview {
		view = m.viewport.View()
		title = " Commit message [preview] "
	} else {
		view = m.textarea.View()
		title = " Commit message "
	}

	titleBorder := lipgloss.RoundedBorder()
	titleBorder.Top = title + strings.Repeat(
		"─", m.textarea.Width()-lipgloss.Width(title)+2) // +2 is to accommodate for horizontal padding

	return m.boxStyle.
		BorderStyle(titleBorder).
		Render(view) + m.helpTextView()
}

func (m *commitViewModel) helpTextView() string {
	if m.helpText {
		return fmt.Sprintf("\n%s:commit %s:clear %s:undo %s:regen %s:preview %s:abort",
			BoldYellow.Render("CTRL+X"),
			BoldYellow.Render("CTRL+K"),
			BoldYellow.Render("CTRL+Z"),
			BoldYellow.Render("CTRL+R"),
			BoldYellow.Render("CTRL+P"),
			BoldYellow.Render("ESC"))
	}
	return ""
}

func uintPtr(v uint) *uint { return &v }
