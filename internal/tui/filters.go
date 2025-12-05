package tui

import (
	"strings"

	"github.com/ramayac/multi-cmd/internal/models"
)

func (m Model) getFilteredFolders() []models.Folder {
	if m.folderFilterText == "" {
		return m.folders
	}

	var filtered []models.Folder
	filterLower := strings.ToLower(m.folderFilterText)
	for _, folder := range m.folders {
		if strings.Contains(strings.ToLower(folder.Name), filterLower) {
			filtered = append(filtered, folder)
		}
	}

	return filtered
}

func (m Model) getFilteredCommands() []models.Command {
	if m.commandFilterText == "" {
		return m.commands
	}

	var filtered []models.Command
	filterLower := strings.ToLower(m.commandFilterText)
	for _, cmd := range m.commands {
		if strings.Contains(strings.ToLower(cmd.Name), filterLower) {
			filtered = append(filtered, cmd)
		}
	}

	return filtered
}
