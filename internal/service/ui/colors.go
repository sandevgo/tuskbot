package ui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle Используем ANSI 6 (Cyan) для заголовков — он хорошо читается везде
	TitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).MarginBottom(1)

	// UsageStyle ANSI 2 (Green) для аргументов и использования
	UsageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	// DescStyle ANSI 8 (Bright Black / Gray) для описаний, чтобы они были менее яркими
	DescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// FlagStyle ANSI 3 (Yellow) для флагов
	FlagStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)
