package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/DanSega1/condor-tui/internal/client"
	"github.com/DanSega1/condor-tui/internal/config"
	"github.com/DanSega1/condor-tui/internal/sysinfo"
)

// logoLines is the condor-tui ASCII art (figlet standard font, cleaner).
var logoLines = []string{
	`_________  _  _____  ____  ___ `,
	` / ___/ __ \/ |/ / _ \/ __ \/ _ \`,
	`/ /__/ /_/ /    / // / /_/ / , _/`,
	`\___/\____/_/|_/____/\____/_/|_|`,
}

const appVersion = "v0.2.0"

// renderHeader produces the full-width info+logo banner that sits above the tab bar.
func renderHeader(width int, cfg AppConfig, stats sysinfo.Stats, tasks []client.TaskRecord) string {
	// Left info panel.
	engineKind := resolveEngineKind(cfg)
	engineLabel := engineKind
	if cfg.Engine.URL != "" {
		engineLabel = cfg.Engine.URL
	}

	storeLabel := cfg.StorePath
	if storeLabel == "" {
		storeLabel = ".conductor/tasks.json"
	}
	if len(storeLabel) > 40 {
		storeLabel = "…" + storeLabel[len(storeLabel)-39:]
	}

	infoStyle := lipgloss.NewStyle().Foreground(colorSubtext)
	labelStyle := lipgloss.NewStyle().Foreground(colorMuted)

	row := func(label, value string) string {
		return labelStyle.Render(fmt.Sprintf(" %-9s", label+":")) + infoStyle.Render(value)
	}

	leftLines := []string{
		row("Engine", engineLabel),
		row("Store", storeLabel),
		row("Version", styleKey.Render(appVersion)),
	}

	// Stats line (only shown when at least one metric is enabled).
	if cfg.Stats.CPU || cfg.Stats.Memory || cfg.Stats.Network {
		var parts []string
		if cfg.Stats.CPU {
			parts = append(parts, "CPU "+styleStat(stats.CPU))
		}
		if cfg.Stats.Memory {
			parts = append(parts, "MEM "+styleStat(stats.Memory))
		}
		if cfg.Stats.Network {
			parts = append(parts, stats.NetUp+"  "+stats.NetDn)
		}
		leftLines = append(leftLines, " "+strings.Join(parts, "  "))
	}

	// Task status summary (compact one-liner).
	if taskSummary := renderStatusBar(tasks); taskSummary != "" {
		leftLines = append(leftLines, " "+taskSummary)
	}

	// Pad left panel to match logo height.
	for len(leftLines) < len(logoLines) {
		leftLines = append(leftLines, "")
	}

	// Right logo panel.
	logoWidth := 0
	for _, l := range logoLines {
		if w := lipgloss.Width(l); w > logoWidth {
			logoWidth = w
		}
	}

	logoStyle := lipgloss.NewStyle().Foreground(colorTeal).Bold(true)

	// Compose side by side.
	leftWidth := width - logoWidth - 2
	if leftWidth < 20 {
		// Terminal too narrow — just render left info stacked.
		var sb strings.Builder
		bg := lipgloss.NewStyle().Background(colorBase)
		for _, l := range leftLines {
			sb.WriteString(bg.Width(width).Render(l) + "\n")
		}
		return sb.String()
	}

	var sb strings.Builder
	rows := len(logoLines)
	if len(leftLines) > rows {
		rows = len(leftLines)
	}

	bg := lipgloss.NewStyle().Background(colorBase)

	for i := 0; i < rows; i++ {
		var left, right string
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(logoLines) {
			right = logoStyle.Render(logoLines[i])
		}

		leftPadded := lipgloss.NewStyle().
			Background(colorBase).
			Width(leftWidth).
			Render(left)

		rightPadded := lipgloss.NewStyle().
			Background(colorBase).
			Width(logoWidth + 2).
			Render(right)

		sb.WriteString(bg.Render(leftPadded+rightPadded) + "\n")
	}

	// Divider.
	divider := lipgloss.NewStyle().
		Foreground(colorSurface).
		Render(strings.Repeat("─", width))
	sb.WriteString(divider + "\n")

	return sb.String()
}

// resolveEngineKind returns a display string for the connection kind.
func resolveEngineKind(cfg AppConfig) string {
	if cfg.Engine.Kind != "" {
		return cfg.Engine.Kind
	}
	if cfg.Engine.URL != "" {
		url := cfg.Engine.URL
		switch {
		case strings.Contains(url, "k8s") || strings.Contains(url, "kubernetes"):
			return "kubernetes"
		case strings.Contains(url, "docker"):
			return "docker"
		default:
			return "remote"
		}
	}
	return "local"
}

// styleStat colours a metric value: red >80%, yellow >60%, green otherwise.
func styleStat(val string) string {
	// Try to parse percentage.
	var pct float64
	if _, err := fmt.Sscanf(val, "%f%%", &pct); err == nil {
		switch {
		case pct > 80:
			return lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render(val)
		case pct > 60:
			return lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render(val)
		default:
			return lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(val)
		}
	}
	return lipgloss.NewStyle().Foreground(colorSubtext).Render(val)
}

// renderStatusBar builds a compact task status summary line.
func renderStatusBar(tasks []client.TaskRecord) string {
	if len(tasks) == 0 {
		return ""
	}
	counts := make(map[client.TaskStatus]int)
	for _, t := range tasks {
		counts[t.Status]++
	}
	var parts []string
	order := []client.TaskStatus{
		client.StatusRunning,
		client.StatusPending,
		client.StatusCompleted,
		client.StatusFailed,
		client.StatusAwaitingApproval,
		client.StatusApproved,
		client.StatusPolicyDenied,
		client.StatusCancelled,
	}
	for _, s := range order {
		if n := counts[s]; n > 0 {
			badge := miniStatusBadge(s)
			parts = append(parts, fmt.Sprintf("%s %d", badge, n))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	label := lipgloss.NewStyle().Foreground(colorMuted).Render("Tasks:")
	return label + " " + strings.Join(parts, " ")
}

// miniStatusBadge returns a concise coloured badge for the header.
func miniStatusBadge(s client.TaskStatus) string {
	switch s {
	case client.StatusRunning:
		return lipgloss.NewStyle().Foreground(colorBlue).Render("RUN")
	case client.StatusPending:
		return lipgloss.NewStyle().Foreground(colorYellow).Render("PND")
	case client.StatusCompleted:
		return lipgloss.NewStyle().Foreground(colorGreen).Render("OK")
	case client.StatusFailed:
		return lipgloss.NewStyle().Foreground(colorRed).Render("FAIL")
	case client.StatusAwaitingApproval:
		return lipgloss.NewStyle().Foreground(colorMauve).Render("WAIT")
	case client.StatusApproved:
		return lipgloss.NewStyle().Foreground(colorGreen).Render("APV")
	case client.StatusPolicyDenied:
		return lipgloss.NewStyle().Foreground(colorRed).Render("DENY")
	case client.StatusCancelled:
		return lipgloss.NewStyle().Foreground(colorSubtext).Render("CNCL")
	default:
		return string(s)
	}
}

// RenderHeader is the exported form of renderHeader for use in tests.
func RenderHeader(width int, cfg AppConfig, stats sysinfo.Stats) string {
	return renderHeader(width, cfg, stats, nil)
}

type sysInfoMsg struct {
	stats sysinfo.Stats
}

// collectStats returns a tea.Cmd that gathers system metrics.
func collectStats(cfg config.StatsConfig) func() sysInfoMsg {
	return func() sysInfoMsg {
		return sysInfoMsg{stats: sysinfo.Collect(cfg.CPU, cfg.Memory, cfg.Network)}
	}
}
