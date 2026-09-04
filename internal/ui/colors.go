package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

var (
	Magenta       = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	Blue          = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	Cyan          = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	Green         = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	BoldOrange    = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	BoldRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	BoldYellow    = lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#FFD700"), Dark: lipgloss.Color("#FFFF00")}).Bold(true)
	BoldPurple    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	Background    = lipgloss.NewStyle().Background(compat.AdaptiveColor{Light: lipgloss.Color("#DDDDDD"), Dark: lipgloss.Color("#222222")}).Bold(true)
	Muted         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Strikethrough = Muted.Strikethrough(true)
	WhiteBold     = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true).Underline(true)

	YellowBackground = lipgloss.NewStyle().Background(compat.AdaptiveColor{Light: lipgloss.Color("#FFD700"), Dark: lipgloss.Color("#FFFF00")}).Bold(true)
)
