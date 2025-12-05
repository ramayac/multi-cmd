package tui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ramayac/multi-cmd/internal/executor"
	"github.com/ramayac/multi-cmd/internal/models"
)

type view int

const (
	mainView view = iota
	executingView
	doneView
)

type focusArea int

const (
	reposFocus focusArea = iota
	commandsFocus
)

type executionProgressMsg struct {
	repoName    string
	commandName string
	result      models.ExecutionResult
}

type executionCompleteMsg struct {
	results []models.ExecutionResult
	err     error
}

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Select    key.Binding
	Execute   key.Binding
	Quit      key.Binding
	SelectAll key.Binding
	Filter    key.Binding
	Tab       key.Binding
	Reset     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "repos panel"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "commands panel"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Execute: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "execute"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle all"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reset"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type Model struct {
	currentView         view
	focus               focusArea
	repos               []models.Repository
	commands            []models.Command
	selectedCommands    map[int]bool
	repoCursorPos       int
	commandCursorPos    int
	repoScrollOffset    int
	commandScrollOffset int
	outputScrollOffset  int
	repoFilterText      string
	commandFilterText   string
	filterActive        bool
	scanPath            string
	outputPath          string
	results             []models.ExecutionResult
	outputLog           []string
	err                 error
	windowHeight        int
	windowWidth         int
	maxVisibleItems     int
	totalCommands       int
	completedCommands   int
	currentExecRepo     string
	currentExecCommand  string
}

func NewModel(scanPath, configPath, outputPath string, config *models.Config) Model {
	repos := scanRepositories(scanPath)

	return Model{
		currentView:         mainView,
		focus:               reposFocus,
		repos:               repos,
		commands:            config.Commands,
		selectedCommands:    make(map[int]bool),
		repoCursorPos:       0,
		commandCursorPos:    0,
		repoScrollOffset:    0,
		commandScrollOffset: 0,
		repoFilterText:      "",
		commandFilterText:   "",
		filterActive:        false,
		scanPath:            scanPath,
		outputPath:          outputPath,
		outputLog:           []string{"Ready to execute commands..."},
		windowHeight:        0,
		windowWidth:         0,
		maxVisibleItems:     15,
		totalCommands:       0,
		completedCommands:   0,
		currentExecRepo:     "",
		currentExecCommand:  "",
	}
}

func scanRepositories(basePath string) []models.Repository {
	var repos []models.Repository

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return repos
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(basePath, entry.Name())
		gitPath := filepath.Join(fullPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, models.Repository{
				Path:     fullPath,
				Name:     entry.Name(),
				Selected: false,
			})
		}
	}

	return repos
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) addLog(msg string) {
	m.outputLog = append(m.outputLog, msg)
	if len(m.outputLog) > 100 {
		m.outputLog = m.outputLog[len(m.outputLog)-100:]
	}
}

func (m *Model) reset() {
	for i := range m.repos {
		m.repos[i].Selected = false
	}
	m.selectedCommands = make(map[int]bool)
	m.repoFilterText = ""
	m.commandFilterText = ""
	m.filterActive = false
	m.repoCursorPos = 0
	m.commandCursorPos = 0
	m.repoScrollOffset = 0
	m.commandScrollOffset = 0
	m.outputLog = []string{"Ready to execute commands..."}
	m.addLog("Reset: cleared all selections, filters, and output")
}

func (m Model) executeCommands() (tea.Model, tea.Cmd) {
	m.currentView = executingView

	var selectedCmds []models.Command
	for i, selected := range m.selectedCommands {
		if selected && i < len(m.commands) {
			selectedCmds = append(selectedCmds, m.commands[i])
		}
	}

	selectedRepoCount := 0
	for _, repo := range m.repos {
		if repo.Selected {
			selectedRepoCount++
		}
	}

	m.totalCommands = selectedRepoCount * len(selectedCmds)
	m.completedCommands = 0
	m.currentExecRepo = ""
	m.currentExecCommand = ""

	return m, m.executeCommandsAsync(selectedCmds)
}

func (m Model) executeCommandsAsync(selectedCmds []models.Command) tea.Cmd {
	return func() tea.Msg {
		var results []models.ExecutionResult

		for _, repo := range m.repos {
			if !repo.Selected {
				continue
			}

			for _, cmd := range selectedCmds {
				result := executor.ExecuteCommand(repo, cmd)
				results = append(results, result)
			}
		}

		err := executor.WriteResults(results, m.outputPath)

		return executionCompleteMsg{
			results: results,
			err:     err,
		}
	}
}
