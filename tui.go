package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Bubbletea makes you use a model struct for the tui state
type model struct {
	fingerprint  string
	username     string
	isRegistered bool
	textInput    textinput.Model
	errMessage   string

	state           string // register dashboard or share picker
	allUsers        []string
	selectedUsers   map[string]bool
	currentFileId   string
	currentFilename string
	cursor          int
}

// creates new model with the fingerprint and user, to be sent across ssh
func initialModel(fingerprint, username string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter desired username"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	registered := username != ""

	initialState := "register"
	if registered {
		if fileId, _, err := GetUnsharedFileForUser(fingerprint); err == nil && fileId != "" {
			initialState = "share_picker"
		} else {
			initialState = "dashboard"
		}
	}

	fileUsers, _ := GetAllUsers()
	fileId, filename, _ := GetUnsharedFileForUser(fingerprint)

	return model{
		fingerprint:     fingerprint,
		username:        username,
		isRegistered:    registered,
		textInput:       ti,
		state:           initialState,
		selectedUsers:   make(map[string]bool),
		allUsers:        fileUsers,
		currentFileId:   fileId,
		currentFilename: filename,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// update when message called (like keypress)
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
				m.state = "dashboard"
				m.errMessage = ""
			} else if m.state == "dashboard" {
				if fileID, filename, err := GetUnsharedFileForUser(m.fingerprint); err == nil && fileID != "" {
					m.currentFileId = fileID
					m.currentFilename = filename
					m.allUsers, _ = GetAllUsers()
					m.selectedUsers = make(map[string]bool)
					m.cursor = 0
					m.state = "share_picker"
				}
			} else if m.state == "share_picker" {
				storagePath := fmt.Sprintf("./storage/%s", m.currentFileId)
				_ = SaveFileRecord(m.currentFileId, m.currentFilename, m.fingerprint, storagePath)

				for u, allowed := range m.selectedUsers {
					if allowed {
						_ = GrantAccess(m.currentFileId, u)
					}
				}
				// Search for next file to assign sharing to
				if nextId, nextName, err := GetUnsharedFileForUser(m.fingerprint); err == nil && nextId != "" {
					m.currentFileId = nextId
					m.currentFilename = nextName
					m.selectedUsers = make(map[string]bool)

				} else {
					m.state = "dashboard"
				}
			}

		case "up", "k":
			if m.state == "share_picker" && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.state == "share_picker" && m.cursor < len(m.allUsers)-1 {
				m.cursor++
			}
		case " ":
			if m.state == "share_picker" && len(m.allUsers) > 0 {
				targetUser := m.allUsers[m.cursor]
				m.selectedUsers[targetUser] = !m.selectedUsers[targetUser]
			}
		}
	}

	if !m.isRegistered && m.state == "register" {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

// view the tui
func (m model) View() string {
	var s strings.Builder

	s.WriteString("\n  Hermes file sharing\n\n")

	if !m.isRegistered {
		s.WriteString("  Hello, anonymous! Your key is not registered.\n\n")
		s.WriteString("  Choose a username:\n  " + m.textInput.View() + "\n\n")
		if m.errMessage != "" {
			s.WriteString("  [!] " + m.errMessage + "\n\n")
		}
		s.WriteString("  Press Enter to register.\n")
	} else if m.state == "dashboard" {
		s.WriteString(fmt.Sprintf("  Hello, %s!\n\n", m.username))
		s.WriteString("  No pending uploads to share.\n")
		s.WriteString("  (Pipe a file in from your terminal using: cat file | ssh ... send file)\n\n")
	} else if m.state == "share_picker" {
		s.WriteString(fmt.Sprintf("  Select users who can access '%s':\n\n", m.currentFilename))
		for i, u := range m.allUsers {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := "[ ]"
			if m.selectedUsers[u] {
				checked = "[x]"
			}
			s.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checked, u))
		}
		s.WriteString("\n  [Space to toggle, Enter to confirm and save]\n")
	}

	s.WriteString("\n  Press 'esc' to exit.\n")
	return s.String()
}
