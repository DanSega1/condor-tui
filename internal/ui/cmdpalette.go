package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// cmdResult carries the outcome of executing a palette command.
type cmdResult struct {
	msg string
	err bool
}

// cmdPalette is the vim/k9s-style command input activated by ":".
type cmdPalette struct {
	input   textinput.Model
	active  bool
	result  *cmdResult // last execution result (shown briefly)
	width   int
	lastCmd string // most recently confirmed command, consumed by App.Update
}

func newCmdPalette(width int) cmdPalette {
	ti := textinput.New()
	ti.Placeholder = "command  (tab to complete, esc to cancel)"
	ti.CharLimit = 200
	ti.Width = width - 4
	return cmdPalette{input: ti, width: width}
}

// knownCommands lists all built-in commands for tab-completion.
var knownCommands = []string{
	"theme", "store", "refresh", "stats", "open", "quit", "help",
}

func (p cmdPalette) Init() tea.Cmd { return nil }

func (p cmdPalette) Update(msg tea.Msg) (cmdPalette, tea.Cmd) {
	if !p.active {
		return p, nil
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.active = false
			p.input.SetValue("")
			p.result = nil
			return p, nil

		case "tab":
			val := p.input.Value()
			for _, c := range knownCommands {
				if strings.HasPrefix(c, val) && c != val {
					p.input.SetValue(c + " ")
					p.input.CursorEnd()
					return p, nil
				}
			}
			return p, nil

		case "enter":
			result, lastCmd := execCmd(p.input.Value())
			p.result = &result
			p.lastCmd = lastCmd
			p.input.SetValue("")
			if result.err {
				return p, nil
			}
			p.active = false
			return p, nil
		}

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.input.Width = msg.Width - 4
	}

	p.input, cmd = p.input.Update(msg)
	return p, cmd
}

// execCmd parses a raw command string and returns a result + the validated command.
func execCmd(raw string) (cmdResult, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return cmdResult{}, ""
	}

	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return cmdResult{}, ""
	}
	cmdWord := parts[0]
	for _, known := range knownCommands {
		if known == cmdWord {
			return cmdResult{msg: "✓ " + raw}, raw
		}
	}
	return cmdResult{
		msg: fmt.Sprintf("unknown command %q — try: %s", cmdWord, strings.Join(knownCommands, ", ")),
		err: true,
	}, ""
}

// View renders the palette bar (one line, pinned to bottom).
func (p cmdPalette) View() string {
	if !p.active && p.result == nil {
		return ""
	}

	bg := lipgloss.NewStyle().Background(colorSurface).Width(p.width)

	if p.result != nil && !p.active {
		style := lipgloss.NewStyle().Foreground(colorGreen)
		if p.result.err {
			style = lipgloss.NewStyle().Foreground(colorRed)
		}
		return bg.Render("  " + style.Render(p.result.msg))
	}

	prompt := lipgloss.NewStyle().Foreground(colorMauve).Bold(true).Render(":")
	return bg.Render(prompt + p.input.View())
}

// ClearResult dismisses the result flash.
func (p *cmdPalette) ClearResult() { p.result = nil }

// CmdResult is the exported form of cmdResult used in tests.
type CmdResult struct {
	Msg string
	Err bool
}

// ExecCmd is the exported form of execCmd for use in tests.
func ExecCmd(raw string) (CmdResult, string) {
	r, cmd := execCmd(raw)
	return CmdResult{Msg: r.msg, Err: r.err}, cmd
}

