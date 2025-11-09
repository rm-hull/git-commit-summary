package ui

import "github.com/charmbracelet/lipgloss"

var (
	Magenta       = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	Blue          = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	Cyan          = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	BoldBlue      = Blue.Bold(true).Underline(true)
	BoldRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	BoldYellow    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#FFD700", Dark: "#FFFF00"}).Bold(true)
	Background    = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#222222"}).Bold(true)
	Strikethrough = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Strikethrough(true)
)
