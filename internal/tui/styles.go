package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title     lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Warning   lipgloss.Style
	Muted     lipgloss.Style
	Border    lipgloss.Style
	Status    lipgloss.Style
	Help      lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Title:   lipgloss.NewStyle().Bold(true),
		Label:   lipgloss.NewStyle().Foreground(lipgloss.Color("62")),
		Value:   lipgloss.NewStyle().Foreground(lipgloss.Color("230")),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Border:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		Status:  lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		Help:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}
