package models

// Command represents a command that can be executed on repositories
type Command struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

// Config represents the application configuration
type Config struct {
	Commands []Command `yaml:"commands"`
}

// Repository represents a selected repository folder
type Repository struct {
	Path     string
	Name     string
	Selected bool
}

// ExecutionResult represents the result of executing a command on a repository
type ExecutionResult struct {
	RepoName        string
	RepoPath        string
	CommandName     string
	CommandExecuted string
	Output          string
	Error           string
	Success         bool
}
