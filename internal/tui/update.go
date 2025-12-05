package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case executionProgressMsg:
		return m.handleExecutionProgress(msg)
	case executionCompleteMsg:
		return m.handleExecutionComplete(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	default:
		return m, nil
	}
}

func (m Model) handleExecutionProgress(msg executionProgressMsg) (tea.Model, tea.Cmd) {
	m.completedCommands++
	m.currentExecFolder = msg.folderName
	m.currentExecCommand = msg.commandName
	m.results = append(m.results, msg.result)
	return m, nil
}

func (m Model) handleExecutionComplete(msg executionCompleteMsg) (tea.Model, tea.Cmd) {
	m.results = msg.results
	m.err = msg.err
	m.currentView = doneView
	m.outputScrollOffset = 0
	return m, nil
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.windowHeight = msg.Height
	m.windowWidth = msg.Width
	m.maxVisibleItems = msg.Height - 12
	if m.maxVisibleItems < 5 {
		m.maxVisibleItems = 5
	}
	if m.maxVisibleItems > 20 {
		m.maxVisibleItems = 20
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filterActive {
		return m.handleFilterKey(msg)
	}

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Filter):
		return m.enableFilterMode()
	case key.Matches(msg, keys.Reset):
		return m.handleReset()
	case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Left), key.Matches(msg, keys.Right):
		return m.toggleFocus()
	case key.Matches(msg, keys.Up):
		return m.navigateUp()
	case key.Matches(msg, keys.Down):
		return m.navigateDown()
	case key.Matches(msg, keys.Select):
		return m.toggleSelection()
	case key.Matches(msg, keys.SelectAll):
		return m.toggleSelectAll()
	case key.Matches(msg, keys.Execute):
		return m.handleExecute()
	default:
		return m, nil
	}
}

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filterActive = false
		if m.focus == foldersFocus {
			m.folderFilterText = ""
			m.folderCursorPos = 0
			m.folderScrollOffset = 0
		} else {
			m.commandFilterText = ""
			m.commandCursorPos = 0
			m.commandScrollOffset = 0
		}
	case "enter":
		m.filterActive = false
	case "backspace":
		if m.focus == foldersFocus {
			if len(m.folderFilterText) > 0 {
				m.folderFilterText = m.folderFilterText[:len(m.folderFilterText)-1]
				m.folderCursorPos = 0
				m.folderScrollOffset = 0
			}
		} else {
			if len(m.commandFilterText) > 0 {
				m.commandFilterText = m.commandFilterText[:len(m.commandFilterText)-1]
				m.commandCursorPos = 0
				m.commandScrollOffset = 0
			}
		}
	default:
		if len(msg.String()) == 1 {
			if m.focus == foldersFocus {
				m.folderFilterText += msg.String()
				m.folderCursorPos = 0
				m.folderScrollOffset = 0
			} else {
				m.commandFilterText += msg.String()
				m.commandCursorPos = 0
				m.commandScrollOffset = 0
			}
		}
	}

	return m, nil
}

func (m Model) enableFilterMode() (tea.Model, tea.Cmd) {
	if m.currentView == mainView {
		m.filterActive = true
	}
	return m, nil
}

func (m Model) handleReset() (tea.Model, tea.Cmd) {
	if m.currentView != mainView {
		return m, nil
	}

	m.reset()
	m.addLog("Reset selections, filters, and output")
	return m, nil
}

func (m Model) toggleFocus() (tea.Model, tea.Cmd) {
	if m.currentView == mainView {
		if m.focus == foldersFocus {
			m.focus = commandsFocus
		} else {
			m.focus = foldersFocus
		}
	}
	return m, nil
}

func (m Model) navigateUp() (tea.Model, tea.Cmd) {
	if m.currentView == doneView {
		if m.outputScrollOffset > 0 {
			m.outputScrollOffset--
		}
		return m, nil
	}

	if m.currentView != mainView {
		return m, nil
	}

	if m.focus == foldersFocus {
		if m.folderCursorPos > 0 {
			m.folderCursorPos--
			if m.folderCursorPos < m.folderScrollOffset {
				m.folderScrollOffset = m.folderCursorPos
			}
		}
	} else {
		if m.commandCursorPos > 0 {
			m.commandCursorPos--
			if m.commandCursorPos < m.commandScrollOffset {
				m.commandScrollOffset = m.commandCursorPos
			}
		}
	}

	return m, nil
}

func (m Model) navigateDown() (tea.Model, tea.Cmd) {
	if m.currentView == doneView {
		totalLines := m.getTotalOutputLines()
		visible := m.visibleLinesForHeight(m.windowHeight)
		if visible > totalLines {
			visible = totalLines
		}
		maxScroll := totalLines - visible
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.outputScrollOffset < maxScroll {
			m.outputScrollOffset++
		}
		return m, nil
	}

	if m.currentView != mainView {
		return m, nil
	}

	if m.focus == foldersFocus {
		filtered := m.getFilteredFolders()
		if m.folderCursorPos < len(filtered)-1 {
			m.folderCursorPos++
			if m.folderCursorPos >= m.folderScrollOffset+m.maxVisibleItems {
				m.folderScrollOffset = m.folderCursorPos - m.maxVisibleItems + 1
			}
		}
	} else {
		filtered := m.getFilteredCommands()
		if m.commandCursorPos < len(filtered)-1 {
			m.commandCursorPos++
			if m.commandCursorPos >= m.commandScrollOffset+m.maxVisibleItems {
				m.commandScrollOffset = m.commandCursorPos - m.maxVisibleItems + 1
			}
		}
	}

	return m, nil
}

func (m Model) toggleSelection() (tea.Model, tea.Cmd) {
	if m.currentView != mainView {
		return m, nil
	}

	if m.focus == foldersFocus {
		filtered := m.getFilteredFolders()
		if m.folderCursorPos < len(filtered) {
			selectedName := filtered[m.folderCursorPos].Name
			for i := range m.folders {
				if m.folders[i].Name == selectedName {
					m.folders[i].Selected = !m.folders[i].Selected
					m.addLog(fmt.Sprintf("Toggled %s: %v", m.folders[i].Name, m.folders[i].Selected))
					break
				}
			}
		}
	} else {
		filtered := m.getFilteredCommands()
		if m.commandCursorPos < len(filtered) {
			cmdName := filtered[m.commandCursorPos].Name
			for i, cmd := range m.commands {
				if cmd.Name == cmdName {
					m.selectedCommands[i] = !m.selectedCommands[i]
					m.addLog(fmt.Sprintf("Toggled command: %s", cmdName))
					break
				}
			}
		}
	}

	return m, nil
}

func (m Model) toggleSelectAll() (tea.Model, tea.Cmd) {
	if m.currentView != mainView {
		return m, nil
	}

	if m.focus == foldersFocus {
		filtered := m.getFilteredFolders()
		allSelected := true
		for _, folder := range filtered {
			if !folder.Selected {
				allSelected = false
				break
			}
		}

		for i := range m.folders {
			for _, ff := range filtered {
				if m.folders[i].Name == ff.Name {
					m.folders[i].Selected = !allSelected
					break
				}
			}
		}

		if allSelected {
			m.addLog("Deselected all filtered folders")
		} else {
			m.addLog("Selected all filtered folders")
		}
	} else {
		filtered := m.getFilteredCommands()
		allSelected := true

		for _, cmd := range filtered {
			found := false
			for i, c := range m.commands {
				if c.Name == cmd.Name && m.selectedCommands[i] {
					found = true
					break
				}
			}
			if !found {
				allSelected = false
				break
			}
		}

		for _, cmd := range filtered {
			for i, c := range m.commands {
				if c.Name == cmd.Name {
					m.selectedCommands[i] = !allSelected
					break
				}
			}
		}

		if allSelected {
			m.addLog("Deselected all filtered commands")
		} else {
			m.addLog("Selected all filtered commands")
		}
	}

	return m, nil
}

func (m Model) handleExecute() (tea.Model, tea.Cmd) {
	if m.currentView == doneView {
		m.currentView = mainView
		m.results = nil
		m.err = nil
		m.outputPath = ""
		m.outputScrollOffset = 0
		return m, nil
	}

	if m.currentView != mainView {
		return m, nil
	}

	hasFolders := false
	for _, folder := range m.folders {
		if folder.Selected {
			hasFolders = true
			break
		}
	}

	if hasFolders && len(m.selectedCommands) > 0 {
		return m.executeCommands()
	}

	if !hasFolders {
		m.addLog("Error: No folders selected!")
	} else {
		m.addLog("Error: No commands selected!")
	}

	return m, nil
}
