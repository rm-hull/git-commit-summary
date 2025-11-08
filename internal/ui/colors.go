package ui

import "github.com/charmbracelet/lipgloss"

var (
	Magenta  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	Blue     = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	BoldBlue = Blue.Bold(true).Underline(true)
	BoldRed  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)
