package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// item wrappers for bubbles list
type fileItem struct {
	file FileInfo
}

func (i fileItem) FilterValue() string { return i.file.OriginalName + " " + i.file.Owner }
func (i fileItem) Title() string       { return i.file.OriginalName }
func (i fileItem) Description() string {
	return fmt.Sprintf("Owner: %s | ID: %s", i.file.Owner, i.file.ID)
}

// item wrappers for bubbles list
type userItem struct {
	username string
	selected bool
	isSelf   bool
}

// implement list.Item for userItem
func (i userItem) FilterValue() string { return i.username }
func (i userItem) Title() string {
	box := "[ ]"
	if i.selected {
		box = "[x]"
	}
	if i.isSelf {
		return fmt.Sprintf("%s %s (You - Required)", box, i.username)
	}
	return fmt.Sprintf("%s %s", box, i.username)
}
func (i userItem) Description() string { return "Press Space to toggle access" }

type actionItem struct {
	title       string
	description string
	id          int
}

func (i actionItem) FilterValue() string { return i.title }
func (i actionItem) Title() string       { return i.title }
func (i actionItem) Description() string { return i.description }

// Bubbletea makes you use a model struct for the tui state
type model struct {
	fingerprint     string
	username        string
	isRegistered    bool
	textInput       textinput.Model
	errMessage      string
	state           string // register, dashboard, file_actions, share_picker
	fileList        list.Model
	userList        list.Model
	currentFileId   string
	currentFilename string
	actionCursor    int
	actionList      list.Model
}

// init the tui model
func initialModel(fingerprint, username string) model {
	// text for username input
	ti := textinput.New()
	ti.Placeholder = "Enter desired username"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	// determine state based on registration and unshare files
	registered := username != ""
	initialState := "register"
	if registered {
		if fileId, _, err := GetUnsharedFileForUser(fingerprint); err == nil && fileId != "" {
			initialState = "share_picker"
		} else {
			initialState = "dashboard"
		}
	}

	// init file list component
	files, _ := GetAccessibleFiles(username)
	var fileItems []list.Item
	for _, f := range files {
		fileItems = append(fileItems, fileItem{file: f})
	}
	// set up list
	fList := list.New(fileItems, list.NewDefaultDelegate(), 60, 14)
	fList.Title = "File Inbox"
	fList.SetShowHelp(true)

	// init user list component
	allUsers, _ := GetAllUsers()
	fileId, filename, _ := GetUnsharedFileForUser(fingerprint)

	var userItems []list.Item
	for _, u := range allUsers {
		isSelf := (u == username)
		userItems = append(userItems, userItem{
			username: u,
			selected: isSelf,
			isSelf:   isSelf,
		})
	}
	uList := list.New(userItems, list.NewDefaultDelegate(), 60, 14)
	uList.Title = fmt.Sprintf("Share: %s", filename)
	uList.SetShowHelp(true)

	actionItems := []list.Item{
		actionItem{title: "Edit access / Share picker", description: "Modify who can access this file", id: 0},
		actionItem{title: "Delete file", description: "Permanently delete this file", id: 1},
		actionItem{title: "Back to file list", description: "Return to the main file list", id: 2},
	}
	aList := list.New(actionItems, list.NewDefaultDelegate(), 60, 8)
	aList.Title = fmt.Sprintf("Action: %s", filename)
	aList.SetShowHelp(true)

	// return the model to be used
	return model{
		fingerprint:     fingerprint,
		username:        username,
		isRegistered:    registered,
		textInput:       ti,
		state:           initialState,
		fileList:        fList,
		userList:        uList,
		currentFileId:   fileId,
		currentFilename: filename,
		actionList:      aList,
	}
}

// init func for the tui
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// update, changes tui model based on messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// switch for messages
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

				// file list refresh
				files, _ := GetAccessibleFiles(m.username)
				var items []list.Item
				for _, f := range files {
					items = append(items, fileItem{file: f})
				}
				m.fileList.SetItems(items)
				m.errMessage = ""
				return m, nil
			}

			if m.state == "dashboard" {
				selectedItem := m.fileList.SelectedItem()
				if selectedItem != nil {
					fi := selectedItem.(fileItem).file
					m.currentFileId = fi.ID
					m.currentFilename = fi.OriginalName

					// update action list title for selected file
					actionItems := []list.Item{
						actionItem{title: "Edit access / Share picker", description: "Modify who can access this file", id: 0},
						actionItem{title: "Delete file", description: "Remove file from disk and database", id: 1},
						actionItem{title: "Back to file list", description: "Return to the inbox", id: 2},
					}
					m.actionList.SetItems(actionItems)
					m.actionList.Title = fmt.Sprintf("Actions: %s", fi.OriginalName)

					m.state = "file_actions"
				}
				return m, nil
			}

			if m.state == "file_actions" {
				selectedItem := m.actionList.SelectedItem()
				if selectedItem != nil {
					ai := selectedItem.(actionItem)
					if ai.id == 0 {
						// reenter share picker
						allUsers, _ := GetAllUsers()
						var userItems []list.Item
						for _, u := range allUsers {
							userItems = append(userItems, userItem{
								username: u,
								selected: (u == m.username),
								isSelf:   (u == m.username),
							})
						}
						m.userList.SetItems(userItems)
						m.userList.Title = fmt.Sprintf("Edit Access: %s", m.currentFilename)
						m.state = "share_picker"
					} else if ai.id == 1 {
						// Delete file
						_ = DeleteFile(m.currentFileId)
						files, _ := GetAccessibleFiles(m.username)
						var items []list.Item
						for _, f := range files {
							items = append(items, fileItem{file: f})
						}
						m.fileList.SetItems(items)
						m.state = "dashboard"
					} else if ai.id == 2 {
						m.state = "dashboard"
					}
				}
				return m, nil
			}

			if m.state == "share_picker" {
				storagePath := fmt.Sprintf("./storage/%s", m.currentFileId)
				_ = SaveFileRecord(m.currentFileId, m.currentFilename, m.fingerprint, storagePath)

				// save selected users from list
				for _, item := range m.userList.Items() {
					ui := item.(userItem)
					if ui.selected {
						_ = GrantAccess(m.currentFileId, ui.username)
					}
				}

				// Search for next file to assign sharing to
				if nextId, nextName, err := GetUnsharedFileForUser(m.fingerprint); err == nil && nextId != "" {
					m.currentFileId = nextId
					m.currentFilename = nextName
					allUsers, _ := GetAllUsers()
					var userItems []list.Item
					for _, u := range allUsers {
						userItems = append(userItems, userItem{
							username: u,
							selected: (u == m.username),
							isSelf:   (u == m.username),
						})
					}
					m.userList.SetItems(userItems)
					m.userList.Title = fmt.Sprintf("Share: %s", nextName)
				} else {
					m.state = "dashboard"
					files, _ := GetAccessibleFiles(m.username)
					var items []list.Item
					for _, f := range files {
						items = append(items, fileItem{file: f})
					}
					m.fileList.SetItems(items)
				}
				return m, nil
			}

		case " ":
			if m.state == "share_picker" {
				idx := m.userList.Index()
				items := m.userList.Items()
				if idx >= 0 && idx < len(items) {
					ui := items[idx].(userItem)
					if !ui.isSelf {
						ui.selected = !ui.selected
						items[idx] = ui
						m.userList.SetItems(items)
					}
				}
				return m, nil
			}

		case "up", "k":
			if m.state == "file_actions" && m.actionCursor > 0 {
				m.actionCursor--
				return m, nil
			}
		case "down", "j":
			if m.state == "file_actions" && m.actionCursor < 2 {
				m.actionCursor++
				return m, nil
			}
		}
	}

	// send messages to right component
	if !m.isRegistered && m.state == "register" {
		m.textInput, cmd = m.textInput.Update(msg)
	} else if m.state == "dashboard" {
		m.fileList, cmd = m.fileList.Update(msg)
	} else if m.state == "file_actions" {
		m.actionList, cmd = m.actionList.Update(msg)
	} else if m.state == "share_picker" {
		m.userList, cmd = m.userList.Update(msg)
	}

	return m, cmd
}

// view the tui
func (m model) View() string {
	var s strings.Builder

	// all this is the tui view, done with a string builder
	// string builder is also better for performance
	s.WriteString("\n  Hermes file sharing\n\n")

	if !m.isRegistered {
		s.WriteString("  Hello, anonymous! Your key is not registered.\n\n")
		s.WriteString("  Choose a username:\n  " + m.textInput.View() + "\n\n")
		if m.errMessage != "" {
			s.WriteString("  [!] " + m.errMessage + "\n\n")
		}
		s.WriteString("  Press Enter to register.\n")
	} else if m.state == "dashboard" {
		cfg := LoadConfig()
		s.WriteString(fmt.Sprintf("  Hello, %s!\n", m.username))
		s.WriteString(fmt.Sprintf("\n  (Save a file using: cat file | ssh %s -p %s send name)\n\n", cfg.Host, cfg.Port))
		s.WriteString(m.fileList.View())
	} else if m.state == "file_actions" {
		s.WriteString(fmt.Sprintf("  File: %s\n", m.currentFilename))
		s.WriteString(fmt.Sprintf("  ID:   %s\n\n", m.currentFileId))
		cfg := LoadConfig()
		s.WriteString("  Download via CLI:\n")
		s.WriteString(fmt.Sprintf("  ssh %s -p %s download %s > %s/%s\n\n", cfg.Host, cfg.Port, m.currentFileId, cfg.DownloadDir, m.currentFilename))

		s.WriteString(m.actionList.View())

	} else if m.state == "share_picker" {
		s.WriteString(m.userList.View())
		s.WriteString("\n  [Type to filter | Space to toggle access | Enter to confirm]\n")
	}

	return s.String()
}
