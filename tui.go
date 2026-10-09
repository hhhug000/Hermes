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
	searchQuery     string

	accessibleFiles  []FileInfo
	fileSearchQuery  string
	fileActionCursor int
}

func (m model) filteredUsers() []string {
	q := strings.ToLower(strings.TrimSpace(m.searchQuery))
	if q == "" {
		return m.allUsers
	}
	var res []string
	for _, u := range m.allUsers {
		if strings.Contains(strings.ToLower(u), q) {
			res = append(res, u)
		}
	}
	return res
}

func (m model) filteredFiles() []FileInfo {
	q := strings.ToLower(strings.TrimSpace(m.fileSearchQuery))
	if q == "" {
		return m.accessibleFiles
	}
	var res []FileInfo
	for _, f := range m.accessibleFiles {
		if strings.Contains(strings.ToLower(f.OriginalName), q) || strings.Contains(strings.ToLower(f.Owner), q) {
			res = append(res, f)
		}
	}
	return res
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

	m := model{
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
	if initialState == "share_picker" {
		m.selectedUsers[username] = true
	}
	if initialState == "dashboard" {
		m.accessibleFiles, _ = GetAccessibleFiles(username)
	}

	return m
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
				m.accessibleFiles, _ = GetAccessibleFiles(m.username)
				m.errMessage = ""
			} else if m.state == "dashboard" {
				filtered := m.filteredFiles()
				if len(filtered) > 0 && m.cursor < len(filtered) {
					selected := filtered[m.cursor]
					m.currentFileId = selected.ID
					m.currentFilename = selected.OriginalName
					m.state = "file_actions"
					m.cursor = 0
				}
			} else if m.state == "file_actions" {
				if m.cursor == 0 {
					m.allUsers, _ = GetAllUsers()
					m.selectedUsers = make(map[string]bool)
					m.selectedUsers[m.username] = true
					m.state = "share_picker"
					m.cursor = 0
				} else if m.cursor == 1 {
					_ = DeleteFile(m.currentFileId)
					m.accessibleFiles, _ = GetAccessibleFiles(m.username)
					m.state = "dashboard"
					m.cursor = 0
				} else if m.cursor == 2 {
					m.state = "dashboard"
					m.cursor = 0
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
					m.searchQuery = ""

				} else {
					m.state = "dashboard"
				}
			}

		case "up", "k":
			if m.state == "share_picker" && m.cursor > 0 {
				m.cursor--
			} else if m.state == "dashboard" && m.cursor > 0 {
				m.cursor--
			} else if m.state == "file_actions" && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.state == "share_picker" {
				filtered := m.filteredUsers()
				if m.cursor < len(filtered)-1 {
					m.cursor++
				}
			} else if m.state == "dashboard" {
				filtered := m.filteredFiles()
				if m.cursor < len(filtered)-1 {
					m.cursor++
				}
			} else if m.state == "file_actions" {
				if m.cursor < 2 {
					m.cursor++
				}
			}
		case " ":
			if m.state == "share_picker" && len(m.allUsers) > 0 {
				filtered := m.filteredUsers()
				if len(filtered) > 0 && m.cursor < len(filtered) {
					targetUser := filtered[m.cursor]
					if targetUser != m.username {
						m.selectedUsers[targetUser] = !m.selectedUsers[targetUser]
					}
				} else if m.state == "dashboard" && len(m.fileSearchQuery) > 0 {
					m.fileSearchQuery = m.fileSearchQuery[:len(m.fileSearchQuery)-1]
					m.cursor = 0
				}
			}
		case "backspace":
			if m.state == "share_picker" && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.cursor = 0
			}
		default:
			if m.state == "share_picker" && msg.Type == tea.KeyRunes {
				m.searchQuery += string(msg.Runes)
				m.cursor = 0
			} else if m.state == "dashboard" && msg.Type == tea.KeyRunes {
				m.fileSearchQuery += string(msg.Runes)
				m.cursor = 0
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
		s.WriteString(fmt.Sprintf("  Search files: %s_\n\n", m.fileSearchQuery))

		filtered := m.filteredFiles()
		if len(filtered) == 0 {
			s.WriteString("  No files found. (Pipe a file in using: cat file | ssh ... send name)\n")
		}

		for i, f := range filtered {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("  %s %-20s (Owner: %s)\n", cursor, f.OriginalName, f.Owner))
		}
		s.WriteString("\n  [Type to search | Up/Down to navigate | Enter to select file]\n")
	} else if m.state == "file_actions" {
		s.WriteString(fmt.Sprintf("  File: %s\n\n", m.currentFilename))
		s.WriteString("  Select an action:\n\n")

		options := []string{
			"Edit access / Share picker",
			"Delete file",
			"Back to file list",
		}
		for i, opt := range options {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s.WriteString(fmt.Sprintf("  %s %s\n", cursor, opt))
		}
		s.WriteString("\n  [Up/Down to select | Enter to confirm]\n")
	} else if m.state == "share_picker" {
		s.WriteString(fmt.Sprintf("  Sharing file: %s\n\n", m.currentFilename))
		s.WriteString(fmt.Sprintf("  Search users: %s_\n\n", m.searchQuery))

		filtered := m.filteredUsers()
		if len(filtered) == 0 {
			s.WriteString("  No matching users found.\n")
		}

		for i, u := range filtered {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := "[ ]"
			if m.selectedUsers[u] {
				checked = "[x]"
			}
			displayLabel := u
			if u == m.username {
				displayLabel += " (You - Required)"
			}

			s.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, checked, displayLabel))
		}
		s.WriteString("\n  [Type to search | Space to toggle | Enter to confirm]\n")
	}

	s.WriteString("\n  Press 'esc' to exit.\n")
	return s.String()
}
