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
	"github.com/cockroachdb/errors"
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

func initialCommitViewModel(message string) (*commitViewModel, error) {
	ta := textarea.New()
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	height := 2
	messageLines := strings.Count(message, "\n") + 1
	if height < messageLines {
		height = messageLines
	}
	if height > 15 {
		height = 15
	}
	ta.SetHeight(height)
	ta.SetWidth(72 + 2) // +2 is to accommodate for horizontal padding
	ta.SetValue(message)
	if message == "" {
		ta.Placeholder = "Unable to provide a commit summary: staged files may be too large to\nbe summarized or were excluded from the visible diff."
	} else {
		ta.Placeholder = "Please supply a commit message."
	}

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(ta.Width(), ta.Height())

	customStyle := styles.DarkStyleConfig
	customStyle.Document.Margin = uintPtr(0)
	customStyle.H2.BlockSuffix = ""
	renderer, err := glamour.NewTermRenderer(
		glamour.WithPreservedNewLines(),
		glamour.WithStyles(customStyle),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create glamour renderer")
	}

	return &commitViewModel{
		textarea: ta,
		viewport: vp,
		history:  NewHistory(message),
		boxStyle: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("6")). // Cyan
			Padding(0, 1),
		preview:  false,
		helpText: true,
		renderer: renderer,
	}, nil
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
		case tea.KeyCtrlX:
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return commitMsg(m.textarea.Value()) }

		case tea.KeyCtrlR:
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return regenerateMsg{} }

		case tea.KeyCtrlC, tea.KeyEsc:
			if m.preview && msg.Type == tea.KeyEsc {
				m.preview = false
				m.textarea.Focus()
				return m, nil
			}
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return abortMsg{} }

		case tea.KeyCtrlP:
			if m.preview {
				m.preview = false
				m.textarea.Focus()
				return m, nil
			}
			m.preview = true
			m.textarea.Blur()
			out, err := m.renderer.Render(m.textarea.Value())
			if err != nil {
				message := fmt.Sprintf("%s:\n%v", BoldRed.Render("Error rendering preview:"), err)
				m.viewport.SetContent(message)
			} else {
				m.viewport.SetContent(strings.TrimSpace(out))
			}
			return m, nil

		}

		if m.preview {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

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
		Render(view) + "\n" + m.helpTextView()
}

func (m *commitViewModel) helpTextView() string {
	if !m.helpText {
		return ""
	}

	if m.preview {
		return fmt.Sprintf("%s:commit %s:clear %s:undo %s:regen %s:editor  %s:back",
			BoldYellow.Render("CTRL+X"),
			Strikethrough.Render("CTRL+K"),
			Strikethrough.Render("CTRL+Z"),
			BoldYellow.Render("CTRL+R"),
			BoldYellow.Render("CTRL+P"),
			BoldYellow.Render("ESC"))
	}

	return fmt.Sprintf("%s:commit %s:clear %s:undo %s:regen %s:preview %s:abort",
		BoldYellow.Render("CTRL+X"),
		BoldYellow.Render("CTRL+K"),
		BoldYellow.Render("CTRL+Z"),
		BoldYellow.Render("CTRL+R"),
		BoldYellow.Render("CTRL+P"),
		BoldYellow.Render("ESC"))
}

func uintPtr(v uint) *uint { return &v }
