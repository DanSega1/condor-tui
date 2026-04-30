package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme holds the full colour palette for a named theme.
type Theme struct {
	Name string

	Base     lipgloss.Color
	Surface  lipgloss.Color
	Muted    lipgloss.Color
	Text     lipgloss.Color
	Subtext  lipgloss.Color
	Green    lipgloss.Color
	Yellow   lipgloss.Color
	Red      lipgloss.Color
	Blue     lipgloss.Color
	Mauve    lipgloss.Color
	Peach    lipgloss.Color
	Teal     lipgloss.Color
	Lavender lipgloss.Color
}

// Built-in themes.
var (
	ThemeCatppuccinMocha = Theme{
		Name:     "catppuccin-mocha",
		Base:     "#1e1e2e",
		Surface:  "#313244",
		Muted:    "#6c7086",
		Text:     "#cdd6f4",
		Subtext:  "#a6adc8",
		Green:    "#a6e3a1",
		Yellow:   "#f9e2af",
		Red:      "#f38ba8",
		Blue:     "#89b4fa",
		Mauve:    "#cba6f7",
		Peach:    "#fab387",
		Teal:     "#94e2d5",
		Lavender: "#b4befe",
	}

	ThemeCatppuccinLatte = Theme{
		Name:     "catppuccin-latte",
		Base:     "#eff1f5",
		Surface:  "#ccd0da",
		Muted:    "#9ca0b0",
		Text:     "#4c4f69",
		Subtext:  "#5c5f77",
		Green:    "#40a02b",
		Yellow:   "#df8e1d",
		Red:      "#d20f39",
		Blue:     "#1e66f5",
		Mauve:    "#8839ef",
		Peach:    "#fe640b",
		Teal:     "#179299",
		Lavender: "#7287fd",
	}

	ThemeDracula = Theme{
		Name:     "dracula",
		Base:     "#282a36",
		Surface:  "#44475a",
		Muted:    "#6272a4",
		Text:     "#f8f8f2",
		Subtext:  "#bfbfbf",
		Green:    "#50fa7b",
		Yellow:   "#f1fa8c",
		Red:      "#ff5555",
		Blue:     "#6272a4",
		Mauve:    "#bd93f9",
		Peach:    "#ffb86c",
		Teal:     "#8be9fd",
		Lavender: "#ff79c6",
	}

	ThemeNord = Theme{
		Name:     "nord",
		Base:     "#2e3440",
		Surface:  "#3b4252",
		Muted:    "#4c566a",
		Text:     "#eceff4",
		Subtext:  "#d8dee9",
		Green:    "#a3be8c",
		Yellow:   "#ebcb8b",
		Red:      "#bf616a",
		Blue:     "#5e81ac",
		Mauve:    "#b48ead",
		Peach:    "#d08770",
		Teal:     "#8fbcbb",
		Lavender: "#88c0d0",
	}

	ThemeGruvbox = Theme{
		Name:     "gruvbox",
		Base:     "#282828",
		Surface:  "#3c3836",
		Muted:    "#928374",
		Text:     "#ebdbb2",
		Subtext:  "#d5c4a1",
		Green:    "#b8bb26",
		Yellow:   "#fabd2f",
		Red:      "#fb4934",
		Blue:     "#83a598",
		Mauve:    "#d3869b",
		Peach:    "#fe8019",
		Teal:     "#8ec07c",
		Lavender: "#458588",
	}
)

var allThemes = []Theme{
	ThemeCatppuccinMocha,
	ThemeCatppuccinLatte,
	ThemeDracula,
	ThemeNord,
	ThemeGruvbox,
}

// ThemeByName returns the theme matching name (case-insensitive). If name is
// unknown, it logs a warning to stderr and returns catppuccin-mocha.
func ThemeByName(name string) Theme {
	for _, t := range allThemes {
		if t.Name == name {
			return t
		}
	}
	if name != "" {
		fmt.Printf("condor-tui: unknown theme %q — falling back to catppuccin-mocha\n", name)
	}
	return ThemeCatppuccinMocha
}

// ThemeNames returns the list of all built-in theme names.
func ThemeNames() []string {
	names := make([]string, len(allThemes))
	for i, t := range allThemes {
		names[i] = t.Name
	}
	return names
}
