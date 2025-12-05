package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/ramayac/multi-cmd/internal/models"
)

// Execute runs the selected commands on the selected repositories
func Execute(repos []models.Repository, commands []models.Command) []models.ExecutionResult {
	var results []models.ExecutionResult

	for _, repo := range repos {
		if !repo.Selected {
			continue
		}

		for _, cmd := range commands {
			result := ExecuteCommand(repo, cmd)
			results = append(results, result)
		}
	}

	return results
}

func ExecuteCommand(repo models.Repository, command models.Command) models.ExecutionResult {
	// Build the full command string for display
	cmdString := command.Cmd
	if len(command.Args) > 0 {
		for _, arg := range command.Args {
			cmdString += " " + arg
		}
	}

	result := models.ExecutionResult{
		RepoName:        repo.Name,
		RepoPath:        repo.Path,
		CommandName:     command.Name,
		CommandExecuted: cmdString,
	}

	cmd := exec.Command(command.Cmd, command.Args...)
	cmd.Dir = repo.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Output = stdout.String()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("%v: %s", err, stderr.String())
	} else {
		result.Success = true
	}

	return result
}

// WriteResults writes the execution results to a file
func WriteResults(results []models.ExecutionResult, outputPath string) error {
	var buf bytes.Buffer

	buf.WriteString("# Repository Check Results\n\n")

	currentRepo := ""
	for _, result := range results {
		if result.RepoName != currentRepo {
			currentRepo = result.RepoName
			buf.WriteString(fmt.Sprintf("## %s\n", result.RepoName))
			buf.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", result.RepoPath))
		}

		buf.WriteString(fmt.Sprintf("### %s\n", result.CommandName))
		buf.WriteString(fmt.Sprintf("**Command:** `%s`\n\n", result.CommandExecuted))

		if result.Success {
			buf.WriteString("```\n")
			buf.WriteString(result.Output)
			if len(result.Output) > 0 && result.Output[len(result.Output)-1] != '\n' {
				buf.WriteString("\n")
			}
			buf.WriteString("```\n\n")
		} else {
			buf.WriteString(fmt.Sprintf("**Error:** %s\n\n", result.Error))
		}
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}
