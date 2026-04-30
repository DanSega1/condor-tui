// Package ui contains all Bubble Tea model and view logic for condor-tui.
package ui

import "github.com/charmbracelet/lipgloss"

// activeTheme is the theme applied at startup via applyTheme.
var activeTheme = ThemeCatppuccinMocha

// Colour aliases — populated by applyTheme from the active Theme.
var (
	colorBase     lipgloss.Color
	colorSurface  lipgloss.Color
	colorMuted    lipgloss.Color
	colorText     lipgloss.Color
	colorSubtext  lipgloss.Color
	colorGreen    lipgloss.Color
	colorYellow   lipgloss.Color
	colorRed      lipgloss.Color
	colorBlue     lipgloss.Color
	colorMauve    lipgloss.Color
	colorPeach    lipgloss.Color
	colorTeal     lipgloss.Color
	colorLavender lipgloss.Color
)

// Tab bar styles — rebuilt by applyTheme.
var (
	styleActiveTab   lipgloss.Style
	styleInactiveTab lipgloss.Style
	styleTabBar      lipgloss.Style
)

// Status badge styles — rebuilt by applyTheme.
var (
	styleBadgePending   lipgloss.Style
	styleBadgeRunning   lipgloss.Style
	styleBadgeCompleted lipgloss.Style
	styleBadgeFailed    lipgloss.Style
	styleBadgeOther     lipgloss.Style
)

// General content styles — rebuilt by applyTheme.
var (
	stylePanelTitle lipgloss.Style
	styleHelp       lipgloss.Style
	styleSelected   lipgloss.Style
	styleDimmed     lipgloss.Style
	styleDetail     lipgloss.Style
	styleKey        lipgloss.Style
	styleMauve      lipgloss.Style
	styleValue      lipgloss.Style
	styleError      lipgloss.Style
	styleBorder     lipgloss.Style
)

func init() {
	applyTheme(ThemeCatppuccinMocha)
}

// applyTheme sets the active theme and rebuilds all style vars.
func applyTheme(t Theme) {
	activeTheme = t

	colorBase = t.Base
	colorSurface = t.Surface
	colorMuted = t.Muted
	colorText = t.Text
	colorSubtext = t.Subtext
	colorGreen = t.Green
	colorYellow = t.Yellow
	colorRed = t.Red
	colorBlue = t.Blue
	colorMauve = t.Mauve
	colorPeach = t.Peach
	colorTeal = t.Teal
	colorLavender = t.Lavender

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

	styleBadgePending = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	styleBadgeRunning = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	styleBadgeCompleted = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleBadgeFailed = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	styleBadgeOther = lipgloss.NewStyle().Foreground(colorMuted).Bold(true)

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

	styleDimmed = lipgloss.NewStyle().Foreground(colorMuted)
	styleDetail = lipgloss.NewStyle().Foreground(colorSubtext)
	styleKey = lipgloss.NewStyle().Foreground(colorMauve).Bold(true)
	styleMauve = lipgloss.NewStyle().Foreground(colorMauve)
	styleValue = lipgloss.NewStyle().Foreground(colorText)
	styleError = lipgloss.NewStyle().Foreground(colorRed)

	styleBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSurface).
		Padding(0, 1)
}

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
