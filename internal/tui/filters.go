package tui

import (
	"strings"

	"github.com/ramayac/multi-cmd/internal/models"
)

func (m Model) getFilteredRepos() []models.Repository {
	if m.repoFilterText == "" {
		return m.repos
	}

	var filtered []models.Repository
	filterLower := strings.ToLower(m.repoFilterText)
	for _, repo := range m.repos {
		if strings.Contains(strings.ToLower(repo.Name), filterLower) {
			filtered = append(filtered, repo)
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
