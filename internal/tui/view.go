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

	if m.currentExecFolder != "" {
		progress += fmt.Sprintf("\nFolder: %s\n", m.currentExecFolder)
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

	folderWidth := int(float64(m.windowWidth-4) * 0.7)
	cmdWidth := m.windowWidth - folderWidth - 4
	if folderWidth < 40 {
		folderWidth = 40
	}
	if cmdWidth < 25 {
		cmdWidth = 25
	}

	filterSection := m.renderFilterSection(folderWidth, cmdWidth)
	s.WriteString(filterSection)
	s.WriteString("\n")

	foldersPanel := m.renderFoldersPanel(folderWidth)
	commandsPanel := m.renderCommandsPanel(cmdWidth)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, foldersPanel, commandsPanel)
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

	folderWidth := int(float64(m.windowWidth-4) * 0.7)
	cmdWidth := m.windowWidth - folderWidth - 4
	if folderWidth < 40 {
		folderWidth = 40
	}
	if cmdWidth < 25 {
		cmdWidth = 25
	}

	filterSection := m.renderFilterSection(folderWidth, cmdWidth)
	s.WriteString(filterSection)
	s.WriteString("\n")

	foldersPanel := m.renderFoldersPanel(folderWidth)
	commandsPanel := m.renderCommandsPanel(cmdWidth)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, foldersPanel, commandsPanel)
	s.WriteString(panels)

	s.WriteString("\n")
	s.WriteString(m.renderOutputLog())

	s.WriteString("\n")
	if m.filterActive {
		if m.focus == foldersFocus {
			s.WriteString(helpStyle.Render("Filter Folders: " + m.folderFilterText + "█ • esc: cancel • enter: done"))
		} else {
			s.WriteString(helpStyle.Render("Filter Commands: " + m.commandFilterText + "█ • esc: cancel • enter: done"))
		}
	} else {
		s.WriteString(helpStyle.Render("↑/↓: navigate • space: toggle • a: all • /: filter • r: reset • tab: switch • enter: execute • q: quit"))
	}

	return s.String()
}

func (m Model) renderFilterSection(folderWidth, cmdWidth int) string {
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

	folderFilterPanel := renderFilter(m.folderFilterText, m.focus == foldersFocus, m.filterActive, folderWidth)
	cmdFilterPanel := renderFilter(m.commandFilterText, m.focus == commandsFocus, m.filterActive, cmdWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, folderFilterPanel, cmdFilterPanel)
}

func (m Model) renderFoldersPanel(width int) string {
	filtered := m.getFilteredFolders()

	selectedFolders := make(map[string]bool)
	for _, folder := range m.folders {
		if folder.Selected {
			selectedFolders[folder.Name] = true
		}
	}

	items := make([]string, len(filtered))
	for i, folder := range filtered {
		items[i] = folder.Name
	}

	selectedCount := 0
	for _, folder := range m.folders {
		if folder.Selected {
			selectedCount++
		}
	}

	return m.renderListPanel(
		"📁 Folders",
		items,
		selectedFolders,
		m.folderCursorPos,
		m.folderScrollOffset,
		m.focus == foldersFocus,
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
	currentFolder := ""

	for _, result := range m.results {
		if result.FolderName != currentFolder {
			currentFolder = result.FolderName
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

		content.WriteString(fmt.Sprintf("Executed: %d commands on %d folders | Success: %d | Failed: %d\n",
			len(m.selectedCommands), countSelectedFolders(m.folders), successCount, failCount))
	}

	if m.err == nil && len(m.results) > 0 {
		content.WriteString("\n")

		var allLines []string
		currentFolder := ""
		for _, result := range m.results {
			if result.FolderName != currentFolder {
				currentFolder = result.FolderName
				allLines = append(allLines, "")
				allLines = append(allLines, successStyle.Render("Folder: ")+result.FolderName)
				allLines = append(allLines, dimmedStyle.Render("Path: ")+result.FolderPath)
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

func countSelectedFolders(folders []models.Folder) int {
	count := 0
	for _, folder := range folders {
		if folder.Selected {
			count++
		}
	}
	return count
}
