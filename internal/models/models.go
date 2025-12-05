package models

// Command represents a command that can be executed on folders
type Command struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

// Config represents the application configuration
type Config struct {
	Commands []Command `yaml:"commands"`
}

// Folder represents a selectable folder discovered in the scan path
type Folder struct {
	Path     string
	Name     string
	Selected bool
}

// ExecutionResult represents the result of executing a command on a folder
type ExecutionResult struct {
	FolderName      string
	FolderPath      string
	CommandName     string
	CommandExecuted string
	Output          string
	Error           string
	Success         bool
}
