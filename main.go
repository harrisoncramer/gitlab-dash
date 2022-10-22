package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	err := tea.NewProgram(&Model{}).Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Model struct {
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	return ""
}

func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
