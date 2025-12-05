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
	panel := m.renderExecutionPanel()
	help := helpStyle.Render("↑/↓: scroll output • enter: return to main • q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, panel, help)
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
	content := m.renderListContent(
		header,
		items,
		selectedItems,
		cursorPos,
		scrollOffset,
		isFocused,
		true,
		fmt.Sprintf("%d/%d selected", totalSelected, len(items)),
		m.maxVisibleItems,
	)

	style := inactivePanelStyle
	if isFocused {
		style = activePanelStyle
	}

	return style.Width(width).Height(m.maxVisibleItems + 4).Render(content)
}

func (m Model) renderListContent(
	header string,
	items []string,
	selectedItems map[string]bool,
	cursorPos int,
	scrollOffset int,
	isFocused bool,
	showSelection bool,
	footer string,
	maxVisible int,
) string {
	var content strings.Builder

	content.WriteString(header + "\n")

	if len(items) == 0 {
		content.WriteString("\nNo items match filter\n")
	} else {
		start := scrollOffset
		end := scrollOffset + maxVisible
		if end > len(items) {
			end = len(items)
		}

		if scrollOffset > 0 {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▲ %d more above...", scrollOffset)))
			content.WriteString("\n")
		}

		for i := start; i < end; i++ {
			itemName := items[i]
			if showSelection {
				cursor := " "
				if i == cursorPos && isFocused {
					cursor = ">"
				}

				checkbox := "[ ]"
				if selectedItems != nil && selectedItems[itemName] {
					checkbox = "[✓]"
				}

				line := fmt.Sprintf("%s %s %s", cursor, checkbox, itemName)
				if i == cursorPos && isFocused {
					content.WriteString(selectedStyle.Render(line))
				} else {
					content.WriteString(line)
				}
			} else {
				content.WriteString(itemName)
			}
			content.WriteString("\n")
		}

		if end < len(items) {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("▼ %d more below...", len(items)-end)))
			content.WriteString("\n")
		}
	}

	if footer != "" {
		content.WriteString("\n")
		content.WriteString(footer)
	}

	return content.String()
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

	return panelStyle.Width(m.windowWidth - 2).Render(content.String())
}

func (m Model) getTotalOutputLines() int {
	return len(m.executionPanelLines())
}

func (m Model) renderExecutionPanel() string {
	lines := m.executionPanelLines()
	if len(lines) == 0 {
		lines = []string{"No execution results yet"}
	}

	maxVisible := m.visibleLinesForHeight(m.windowHeight)
	if maxVisible > len(lines) {
		maxVisible = len(lines)
	}
	if maxVisible < 1 {
		maxVisible = 1
	}

	scrollOffset := m.outputScrollOffset
	maxScroll := len(lines) - maxVisible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	} else if scrollOffset < 0 {
		scrollOffset = 0
	}

	content := m.renderListContent(
		"📊 Execution Output",
		lines,
		nil,
		-1,
		scrollOffset,
		false,
		false,
		"",
		maxVisible,
	)

	return executionPanelStyle.
		Width(m.contentWidthForPanel(m.windowWidth - 2)).
		Height(m.contentHeightForPanel(m.windowHeight)).
		Render(content)
}

func (m Model) executionPanelLines() []string {
	var lines []string

	if m.err != nil {
		lines = append(lines, "❌ Error", "")
		lines = append(lines, fmt.Sprintf("Error writing results: %v", m.err))
		return lines
	}

	successCount := 0
	failCount := 0
	for _, result := range m.results {
		if result.Success {
			successCount++
		} else {
			failCount++
		}
	}

	folderCount := countSelectedFolders(m.folders)
	selectedCmdCount := 0
	for _, selected := range m.selectedCommands {
		if selected {
			selectedCmdCount++
		}
	}

	lines = append(lines, "✅ Execution Complete", "")
	if m.outputPath != "" {
		lines = append(lines, fmt.Sprintf("Results written to: %s", m.outputPath))
	} else {
		lines = append(lines, "Results file path unavailable")
	}
	lines = append(lines, fmt.Sprintf("Executed: %d commands on %d folders | Success: %d | Failed: %d",
		selectedCmdCount, folderCount, successCount, failCount))

	if len(m.results) == 0 {
		return lines
	}

	currentFolder := ""
	for _, result := range m.results {
		if result.FolderName != currentFolder {
			currentFolder = result.FolderName
			lines = append(lines, "")
			lines = append(lines, successStyle.Render("Folder: ")+result.FolderName)
			lines = append(lines, dimmedStyle.Render("Path: ")+result.FolderPath)
		}

		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Command: %s", result.CommandName))
		lines = append(lines, dimmedStyle.Render(fmt.Sprintf("Executed: %s", result.CommandExecuted)))

		if result.Success {
			trimmed := strings.TrimRight(result.Output, "\n")
			if trimmed == "" {
				lines = append(lines, dimmedStyle.Render("(no output)"))
			} else {
				lines = append(lines, strings.Split(trimmed, "\n")...)
			}
		} else {
			lines = append(lines, errorStyle.Render(fmt.Sprintf("Error: %s", result.Error)))
		}
	}

	return lines
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

func (m Model) visibleLinesForHeight(height int) int {
	lines := height - 6
	if lines < 5 {
		lines = 5
	}
	return lines
}

func (m Model) contentWidthForPanel(fullWidth int) int {
	width := fullWidth - 4 // account for 1 padding each side + border
	if width < 10 {
		width = 10
	}
	return width
}

func (m Model) contentHeightForPanel(fullHeight int) int {
	height := fullHeight - 2 // border top/bottom
	if height < 5 {
		height = 5
	}
	return height
}
