package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/cockroachdb/errors"
	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
)

type commitViewModel struct {
	textarea     textarea.Model
	viewport     viewport.Model
	history      *History
	boxStyle     lipgloss.Style
	preview      bool
	helpText     bool
	duration     time.Duration
	renderer     *glamour.TermRenderer
	commitParser conventionalcommits.Machine
}

func initialCommitViewModel(message string, duration time.Duration) (*commitViewModel, error) {
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
	for ta.Line() > 0 {
		ta.CursorUp()
	}
	ta.CursorStart()
	if message == "" {
		ta.Placeholder = "Unable to provide a commit summary: staged files may be too large to\nbe summarized or were excluded from the visible diff."
	} else {
		ta.Placeholder = "Please supply a commit message."
	}

	// Disable the default cursor line background highlighting to match v1 behavior.
	textStyles := textarea.DefaultStyles(false)
	textStyles.Focused.CursorLine = lipgloss.NewStyle()
	textStyles.Blurred.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(textStyles)

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
		viewport: viewport.New(viewport.WithWidth(ta.Width()), viewport.WithHeight(ta.Height())),
		history:  NewHistory(message),
		boxStyle: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("6")). // Cyan
			Padding(0, 1),
		preview:      false,
		helpText:     true,
		duration:     duration,
		renderer:     renderer,
		commitParser: parser.NewMachine(parser.WithTypes(conventionalcommits.TypesConventional)),
	}, nil
}

func (m *commitViewModel) Init() tea.Cmd {
	m.textarea.Focus()
	m.helpText = true
	return textarea.Blink
}

func (m *commitViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	oldValue := m.textarea.Value()

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+a":
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return commitMsg(m.textarea.Value()) }

		case "ctrl+r":
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return regenerateMsg{} }

		case "esc":
			if m.preview {
				m.preview = false
				m.textarea.Focus()
				return m, nil
			}
			m.helpText = false
			m.textarea.Blur()
			return m, func() tea.Msg { return abortMsg{} }

		case "ctrl+p":
			if m.preview {
				m.textarea.Focus()
			} else {
				m.textarea.Blur()
				out, err := m.renderer.Render(m.textarea.Value())
				if err != nil {
					message := fmt.Sprintf("%s:\n%v", BoldRed.Render("Error rendering preview:"), err)
					m.viewport.SetContent(message)
				} else {
					m.viewport.SetContent(strings.TrimSpace(out))
				}
			}
			m.preview = !m.preview
			return m, nil
		}

		if m.preview {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+z":
			if value, ok := m.history.Undo(); ok {
				m.textarea.SetValue(value)
			}
			return m, nil

		case "ctrl+y":
			if value, ok := m.history.Redo(); ok {
				m.textarea.SetValue(value)
			}
			return m, nil

		case "ctrl+x":
			if m.textarea.Value() == "" {
				return m, nil
			}
			lines := strings.Split(m.textarea.Value(), "\n")
			lineIdx := m.textarea.Line()
			if lineIdx < len(lines) {
				cutLine := lines[lineIdx] + "\n"
				err := clipboard.WriteAll(cutLine)
				if err != nil {
					return m, func() tea.Msg { return errMsg{err} }
				}

				lines = append(lines[:lineIdx], lines[lineIdx+1:]...)
				newVal := strings.Join(lines, "\n")
				m.textarea.SetValue(newVal)
				m.history.Add(newVal)

				if len(lines) == 0 {
					lineIdx = 0
				} else if lineIdx >= len(lines) {
					lineIdx = len(lines) - 1
				}

				// Restore cursor position to the desired line
				m.textarea.CursorStart()
				for m.textarea.Line() > lineIdx {
					m.textarea.CursorUp()
				}
			}

			return m, nil

		case "ctrl+k":
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

	case tea.WindowSizeMsg:
		// Adjust textarea height
		helpTextHeight := lipgloss.Height(m.helpTextView())
		borderHeight := m.boxStyle.GetVerticalBorderSize()
		paddingHeight := m.boxStyle.GetVerticalPadding()
		remainingHeight := msg.Height - helpTextHeight - borderHeight - paddingHeight

		m.textarea.SetHeight(max(1, remainingHeight))

		// Adjust viewport for preview
		m.viewport.SetWidth(msg.Width - m.boxStyle.GetHorizontalBorderSize() - m.boxStyle.GetHorizontalPadding())
		m.viewport.SetHeight(max(1, remainingHeight))

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

func (m *commitViewModel) View() tea.View {
	var view string
	var title string

	if m.preview {
		view = m.viewport.View()
		title = " Commit message [preview] "
	} else {
		view = m.textarea.View()
		title = " Commit message "
	}

	durationStr := ""
	if m.duration > 0 {
		durationStr = fmt.Sprintf("┤%.1fs├─", m.duration.Seconds())
	}

	titleBorder := lipgloss.RoundedBorder()
	titleBorder.Top = title + strings.Repeat(
		"─", m.textarea.Width()-lipgloss.Width(title)-lipgloss.Width(durationStr)+2) + durationStr

	if lintErr := m.commitLint(m.textarea.Value()); lintErr != "" {
		padding := m.textarea.Width() - lipgloss.Width(lintErr)
		if padding < 0 {
			padding = 0
		}
		titleBorder.Bottom = strings.Repeat("─", padding) + lintErr
	}

	return tea.NewView(m.boxStyle.
		BorderStyle(titleBorder).
		Render(view) + "\n" + m.helpTextView())
}

func (m *commitViewModel) helpTextView() string {
	if !m.helpText {
		return ""
	}

	if m.preview {
		return fmt.Sprintf("%s:commit %s:clear %s:regen %s:editor  %s:diff %s:back",
			BoldYellow.Render("CTRL+A"),
			Strikethrough.Render("CTRL+K"),
			BoldYellow.Render("CTRL+R"),
			BoldYellow.Render("CTRL+P"),
			BoldYellow.Render("CTRL+D"),
			BoldYellow.Render("ESC"))
	}

	return fmt.Sprintf("%s:commit %s:clear %s:regen %s:preview %s:diff %s:abort",
		BoldYellow.Render("CTRL+A"),
		BoldYellow.Render("CTRL+K"),
		BoldYellow.Render("CTRL+R"),
		BoldYellow.Render("CTRL+P"),
		BoldYellow.Render("CTRL+D"),
		BoldYellow.Render("ESC"))
}

func (m *commitViewModel) commitLint(msg string) string {
	_, err := m.commitParser.Parse([]byte(msg))
	if err != nil {
		return strings.TrimSpace(strings.Split(err.Error(), ": col=")[0])
	}

	lines := strings.Split(msg, "\n")
	for idx, line := range lines {
		if idx == 0 && lipgloss.Width(line) > 50 {
			return "subject line should be no more than 50 chars"
		}

		if lipgloss.Width(line) > 72 {
			return fmt.Sprintf("line %d should be no more than 72 chars", idx+1)
		}
	}
	return ""
}

func uintPtr(v uint) *uint { return new(v) }
