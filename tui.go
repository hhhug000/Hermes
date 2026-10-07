package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	message string
}

func InitialModel() model {
	return model{
		message: "Welcome to Hermes",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "\n  " + m.message + "\n\n  Press q to exit.\n\n"
}
