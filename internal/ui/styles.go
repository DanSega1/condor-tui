// Package ui contains all Bubble Tea model and view logic for condor-tui.
package ui

import "github.com/charmbracelet/lipgloss"

// Colour palette used throughout the TUI.
var (
	colorBase     = lipgloss.Color("#1e1e2e") // base background
	colorSurface  = lipgloss.Color("#313244") // surface / card backgrounds
	colorMuted    = lipgloss.Color("#6c7086") // muted text
	colorText     = lipgloss.Color("#cdd6f4") // primary text
	colorSubtext  = lipgloss.Color("#a6adc8") // secondary text
	colorGreen    = lipgloss.Color("#a6e3a1")
	colorYellow   = lipgloss.Color("#f9e2af")
	colorRed      = lipgloss.Color("#f38ba8")
	colorBlue     = lipgloss.Color("#89b4fa")
	colorMauve    = lipgloss.Color("#cba6f7")
	colorPeach    = lipgloss.Color("#fab387")
	colorTeal     = lipgloss.Color("#94e2d5")
	colorLavender = lipgloss.Color("#b4befe")
)

// Tab bar styles.
var (
	styleActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBase).
			Background(colorBlue).
			Padding(0, 2)

	styleInactiveTab = lipgloss.NewStyle().
				Foreground(colorSubtext).
				Background(colorSurface).
				Padding(0, 2)

	styleTabBar = lipgloss.NewStyle().
			Background(colorSurface)
)

// Status badge styles.
var (
	styleBadgePending = lipgloss.NewStyle().
				Foreground(colorYellow).Bold(true)
	styleBadgeRunning = lipgloss.NewStyle().
				Foreground(colorBlue).Bold(true)
	styleBadgeCompleted = lipgloss.NewStyle().
				Foreground(colorGreen).Bold(true)
	styleBadgeFailed = lipgloss.NewStyle().
				Foreground(colorRed).Bold(true)
	styleBadgeOther = lipgloss.NewStyle().
				Foreground(colorMuted).Bold(true)
)

// General content styles.
var (
	stylePanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLavender).
			MarginBottom(1)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorBase).
			Background(colorLavender).
			Bold(true)

	styleDimmed = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleDetail = lipgloss.NewStyle().
			Foreground(colorSubtext)

	styleKey = lipgloss.NewStyle().
			Foreground(colorMauve).
			Bold(true)

	styleValue = lipgloss.NewStyle().
			Foreground(colorText)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSurface).
			Padding(0, 1)
)

// statusBadge renders a coloured status string.
func statusBadge(status string) string {
	switch status {
	case "pending":
		return styleBadgePending.Render("⏳ pending")
	case "running":
		return styleBadgeRunning.Render("▶ running")
	case "completed":
		return styleBadgeCompleted.Render("✔ completed")
	case "failed":
		return styleBadgeFailed.Render("✖ failed")
	case "awaiting_approval":
		return styleBadgeOther.Render("⏸ awaiting_approval")
	case "approved":
		return styleBadgeOther.Render("✅ approved")
	case "policy_denied":
		return styleBadgeFailed.Render("🚫 policy_denied")
	case "cancelled":
		return styleBadgeOther.Render("⊘ cancelled")
	default:
		return styleBadgeOther.Render(status)
	}
}
