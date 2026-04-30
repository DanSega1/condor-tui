package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// openInEditor suspends the TUI, opens path in the configured editor, then resumes.
func openInEditor(editor, path string) tea.Cmd {
	if editor == "" {
		editor = "vi"
	}
	parts := strings.Fields(editor)
	args := append(parts[1:], path)
	c := exec.Command(parts[0], args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return toolErrMsg{err}
		}
		return toolDoneMsg{fmt.Sprintf("Opened %s in %s", path, parts[0])}
	})
}

// diffTaskJSON suspends the TUI and opens a diff of input vs result JSON.
func diffTaskJSON(diffTool string, inputData, resultData map[string]any) tea.Cmd {
	if diffTool == "" {
		diffTool = "vimdiff"
	}

	inputJSON, _ := json.MarshalIndent(inputData, "", "  ")
	resultJSON, _ := json.MarshalIndent(resultData, "", "  ")

	// Write to temp files.
	inputFile, err := os.CreateTemp("", "condor-input-*.json")
	if err != nil {
		return func() tea.Msg { return toolErrMsg{err} }
	}
	resultFile, err := os.CreateTemp("", "condor-result-*.json")
	if err != nil {
		return func() tea.Msg { return toolErrMsg{err} }
	}

	if _, err := inputFile.Write(inputJSON); err != nil {
		return func() tea.Msg { return toolErrMsg{err} }
	}
	if _, err := resultFile.Write(resultJSON); err != nil {
		return func() tea.Msg { return toolErrMsg{err} }
	}
	inputFile.Close()
	resultFile.Close()

	parts := strings.Fields(diffTool)
	args := append(parts[1:], inputFile.Name(), resultFile.Name())
	c := exec.Command(parts[0], args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		os.Remove(inputFile.Name())
		os.Remove(resultFile.Name())
		if err != nil {
			return toolErrMsg{err}
		}
		return toolDoneMsg{"diff closed"}
	})
}

// openWithOpener opens a path or URL with the system opener (open / xdg-open).
func openWithOpener(opener, target string) tea.Cmd {
	if opener == "" {
		opener = "open"
	}
	c := exec.Command(opener, target)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return toolErrMsg{err}
		}
		return toolDoneMsg{"opened " + target}
	})
}

type toolDoneMsg struct{ msg string }
type toolErrMsg struct{ err error }
