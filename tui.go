package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	fingerprint  string
	username     string
	isRegistered bool
	textInput    textinput.Model
	errMessage   string
}

func initialModel(fingerprint, username string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter desired username"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	registered := username != ""

	return model{
		fingerprint:  fingerprint,
		username:     username,
		isRegistered: registered,
		textInput:    ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit

		case "enter":
			if !m.isRegistered && m.fingerprint != "" {
				val := strings.TrimSpace(m.textInput.Value())
				if val == "" {
					m.errMessage = "Username cannot be empty"
					return m, nil
				}

				err := RegisterUser(m.fingerprint, val)
				if err != nil {
					m.errMessage = "Username is taken, or error"
					return m, nil
				}

				m.username = val
				m.isRegistered = true
				m.errMessage = ""
			}
		}
	}

	if !m.isRegistered {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString("\n  Hermes file sharing\n\n")

	if m.fingerprint == "" {
		s.WriteString("  Hello anonymous (No SSH key provided)\n")
	} else if m.isRegistered {
		s.WriteString("  Hello " + m.username + "!\n")
		s.WriteString("  [todo add file sharing stuff]\n")
	} else {
		s.WriteString("  Hello anonymous! Your key is not registered.\n\n")
		s.WriteString("  Choose a username to register:\n")
		s.WriteString("  " + m.textInput.View() + "\n\n")
		if m.errMessage != "" {
			s.WriteString("  [!] " + m.errMessage + "\n\n")
		}
		s.WriteString("  Press Enter to register, or 'esc' to quit.\n")
	}

	s.WriteString("\n  Press 'esc' to exit.\n")
	return s.String()
}
