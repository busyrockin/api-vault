package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (unexported — for building styles)
	primaryColor   = lipgloss.Color("#00D9FF")
	secondaryColor = lipgloss.Color("#7C3AED")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")
	errorColor     = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")

	// Shorthand styles (exported — for direct .Render() calls)
	Primary = lipgloss.NewStyle().Foreground(primaryColor)
	Success = lipgloss.NewStyle().Foreground(successColor)
	Muted   = lipgloss.NewStyle().Foreground(mutedColor)

	// Base styles
	TitleStyle         = lipgloss.NewStyle().Bold(true).Foreground(primaryColor).MarginBottom(1)
	SubtitleStyle      = lipgloss.NewStyle().Foreground(mutedColor)
	SelectedStyle      = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).PaddingLeft(2)
	NormalStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).PaddingLeft(4)
	StatusRecentStyle  = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	StatusWarningStyle = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	StatusErrorStyle   = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	HelpStyle          = lipgloss.NewStyle().Foreground(mutedColor).MarginTop(1)
	BoxStyle           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Padding(1, 2)
)
