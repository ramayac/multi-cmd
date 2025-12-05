package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ramayac/multi-cmd/internal/models"
)

func (m Model) View() string {
	switch m.currentView {
	case executingView:
		return m.renderExecutingView()
	case doneView:
		return m.renderDoneView()
	default:
		return m.renderMainView()
	}
}

func (m Model) renderExecutingView() string {
	if m.totalCommands == 0 {
		return boxStyle.Render(titleStyle.Render("⚙️  Executing...") + "\n\nInitializing execution...\n")
	}

	percentage := float64(m.completedCommands) / float64(m.totalCommands) * 100
	progress := fmt.Sprintf("\n\nProgress: %d/%d (%.0f%%)\n", m.completedCommands, m.totalCommands, percentage)

	if m.currentExecRepo != "" {
		progress += fmt.Sprintf("\nRepository: %s\n", m.currentExecRepo)
	}
	if m.currentExecCommand != "" {
		progress += fmt.Sprintf("Command: %s\n", m.currentExecCommand)
	}

	barWidth := 40
	filledWidth := int(float64(barWidth) * percentage / 100)
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"
	progress += "\n" + bar + "\n"

	return boxStyle.Render(titleStyle.Render("⚙️  Executing...") + progress)
}

func (m Model) renderDoneView() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🛠️  Multi Commands"))
	s.WriteString("\n\n")

	repoWidth := int(float64(m.windowWidth-4) * 0.7)
	cmdWidth := m.windowWidth - repoWidth - 4
	if repoWidth < 40 {
		repoWidth = 40
	}
	if cmdWidth < 25 {
		cmdWidth = 25
	}

	filterSection := m.renderFilterSection(repoWidth, cmdWidth)
	s.WriteString(filterSection)
	s.WriteString("\n")

	reposPanel := m.renderReposPanel(repoWidth)
	commandsPanel := m.renderCommandsPanel(cmdWidth)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, reposPanel, commandsPanel)
	s.WriteString(panels)

	s.WriteString("\n")
	s.WriteString(m.renderOutputPanel())

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: scroll output • enter: return to main • q: quit"))

	return s.String()
}

func (m Model) renderMainView() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🛠️  Multi Commands"))
	s.WriteString("\n\n")

	repoWidth := int(float64(m.windowWidth-4) * 0.7)
	cmdWidth := m.windowWidth - repoWidth - 4
	if repoWidth < 40 {
		repoWidth = 40
	}
	if cmdWidth < 25 {
		cmdWidth = 25
	}

	filterSection := m.renderFilterSection(repoWidth, cmdWidth)
	s.WriteString(filterSection)
	s.WriteString("\n")

	reposPanel := m.renderReposPanel(repoWidth)
	commandsPanel := m.renderCommandsPanel(cmdWidth)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, reposPanel, commandsPanel)
	s.WriteString(panels)

	s.WriteString("\n")
	s.WriteString(m.renderOutputLog())

	s.WriteString("\n")
	if m.filterActive {
		if m.focus == reposFocus {
			s.WriteString(helpStyle.Render("Filter Repos: " + m.repoFilterText + "█ • esc: cancel • enter: done"))
		} else {
			s.WriteString(helpStyle.Render("Filter Commands: " + m.commandFilterText + "█ • esc: cancel • enter: done"))
		}
	} else {
		s.WriteString(helpStyle.Render("↑/↓: navigate • space: toggle • a: all • /: filter • r: reset • tab: switch • enter: execute • q: quit"))
	}

	return s.String()
}

func (m Model) renderFilterSection(repoWidth, cmdWidth int) string {
	renderFilter := func(filterText string, isFocused, isActive bool, width int) string {
		displayText := "(press / to filter)"
		if filterText != "" {
			displayText = filterText
		}
		if isFocused && isActive {
			displayText += "█"
		}

		style := inactivePanelStyle
		if isFocused && isActive {
			style = activePanelStyle
		}

		return style.Width(width).Render("🔍 Filter: " + displayText)
	}

	repoFilterPanel := renderFilter(m.repoFilterText, m.focus == reposFocus, m.filterActive, repoWidth)
	cmdFilterPanel := renderFilter(m.commandFilterText, m.focus == commandsFocus, m.filterActive, cmdWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, repoFilterPanel, cmdFilterPanel)
}

func (m Model) renderReposPanel(width int) string {
	filtered := m.getFilteredRepos()

	selectedRepos := make(map[string]bool)
	for _, repo := range m.repos {
		if repo.Selected {
			selectedRepos[repo.Name] = true
		}
	}

	items := make([]string, len(filtered))
	for i, repo := range filtered {
		items[i] = repo.Name
	}

	selectedCount := 0
	for _, repo := range m.repos {
		if repo.Selected {
			selectedCount++
		}
	}

	return m.renderListPanel(
		"📁 Repositories",
		items,
		selectedRepos,
		m.repoCursorPos,
		m.repoScrollOffset,
		m.focus == reposFocus,
		width,
		selectedCount,
	)
}

func (m Model) renderCommandsPanel(width int) string {
	filtered := m.getFilteredCommands()

	selectedCommands := make(map[string]bool)
	for i, selected := range m.selectedCommands {
		if selected && i < len(m.commands) {
			selectedCommands[m.commands[i].Name] = true
		}
	}

	items := make([]string, len(filtered))
	for i, cmd := range filtered {
		items[i] = cmd.Name
	}

	return m.renderListPanel(
		"⚡ Commands",
		items,
		selectedCommands,
		m.commandCursorPos,
		m.commandScrollOffset,
		m.focus == commandsFocus,
		width,
		len(m.selectedCommands),
	)
}

func (m Model) renderListPanel(
	header string,
	items []string,
	selectedItems map[string]bool,
	cursorPos int,
	scrollOffset int,
	isFocused bool,
	width int,
	totalSelected int,
) string {
	var content strings.Builder

	content.WriteString(header + "\n")

	if len(items) == 0 {
		content.WriteString("\nNo items match filter\n")
	} else {
		start := scrollOffset
		end := scrollOffset + m.maxVisibleItems
		if end > len(items) {
			end = len(items)
		}

		if scrollOffset > 0 {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▲ %d more above...", scrollOffset)))
			content.WriteString("\n")
		}

		for i := start; i < end; i++ {
			itemName := items[i]
			cursor := " "
			if i == cursorPos && isFocused {
				cursor = ">"
			}

			checkbox := "[ ]"
			if selectedItems[itemName] {
				checkbox = "[✓]"
			}

			line := fmt.Sprintf("%s %s %s", cursor, checkbox, itemName)
			if i == cursorPos && isFocused {
				content.WriteString(selectedStyle.Render(line))
			} else {
				content.WriteString(line)
			}
			content.WriteString("\n")
		}

		if end < len(items) {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▼ %d more below...", len(items)-end)))
			content.WriteString("\n")
		}
	}

	content.WriteString(fmt.Sprintf("\n%d/%d selected", totalSelected, len(items)))

	style := inactivePanelStyle
	if isFocused {
		style = activePanelStyle
	}

	return style.Width(width).Height(m.maxVisibleItems + 4).Render(content.String())
}

func (m Model) renderOutputLog() string {
	var content strings.Builder

	content.WriteString("📋 Output\n")

	start := 0
	if len(m.outputLog) > 3 {
		start = len(m.outputLog) - 3
	}

	for i := start; i < len(m.outputLog); i++ {
		content.WriteString(dimmedStyle.Render(m.outputLog[i]))
		content.WriteString("\n")
	}

	return panelStyle.Width(m.windowWidth).Render(content.String())
}

func (m Model) getTotalOutputLines() int {
	if m.err != nil {
		return 5
	}

	totalLines := 3
	currentRepo := ""

	for _, result := range m.results {
		if result.RepoName != currentRepo {
			currentRepo = result.RepoName
			totalLines += 2
		}

		totalLines++

		if result.Success {
			lines := strings.Split(result.Output, "\n")
			totalLines += len(lines) + 1
		} else {
			totalLines++
		}
		totalLines++
	}

	return totalLines
}

func (m Model) renderOutputPanel() string {
	var content strings.Builder

	if m.err != nil {
		content.WriteString("❌ Error\n\n")
		content.WriteString(fmt.Sprintf("Error writing results: %v\n", m.err))
	} else {
		content.WriteString("✅ Execution Complete\n\n")
		content.WriteString(fmt.Sprintf("Results written to: %s\n", m.outputPath))

		successCount := 0
		failCount := 0
		for _, result := range m.results {
			if result.Success {
				successCount++
			} else {
				failCount++
			}
		}

		content.WriteString(fmt.Sprintf("Executed: %d commands on %d repos | Success: %d | Failed: %d\n",
			len(m.selectedCommands), countSelected(m.repos), successCount, failCount))
	}

	if m.err == nil && len(m.results) > 0 {
		content.WriteString("\n")

		var allLines []string
		currentRepo := ""
		for _, result := range m.results {
			if result.RepoName != currentRepo {
				currentRepo = result.RepoName
				allLines = append(allLines, "")
				allLines = append(allLines, successStyle.Render("Repository: ")+result.RepoName)
				allLines = append(allLines, dimmedStyle.Render("Path: ")+result.RepoPath)
			}

			allLines = append(allLines, "")
			allLines = append(allLines, fmt.Sprintf("Command: %s", result.CommandName))
			allLines = append(allLines, dimmedStyle.Render(fmt.Sprintf("Executed: %s", result.CommandExecuted)))

			if result.Success {
				lines := strings.Split(strings.TrimRight(result.Output, "\n"), "\n")
				for _, line := range lines {
					allLines = append(allLines, line)
				}
			} else {
				allLines = append(allLines, errorStyle.Render(fmt.Sprintf("Error: %s", result.Error)))
			}
		}

		start := m.outputScrollOffset
		end := start + m.maxVisibleItems
		if end > len(allLines) {
			end = len(allLines)
		}

		if m.outputScrollOffset > 0 {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▲ %d more above...\n", m.outputScrollOffset)))
		}

		for i := start; i < end; i++ {
			content.WriteString(allLines[i])
			content.WriteString("\n")
		}

		if end < len(allLines) {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▼ %d more below...", len(allLines)-end)))
		}
	}

	return panelStyle.Width(m.windowWidth).Height(m.maxVisibleItems + 6).Render(content.String())
}

func countSelected(repos []models.Repository) int {
	count := 0
	for _, repo := range repos {
		if repo.Selected {
			count++
		}
	}
	return count
}
