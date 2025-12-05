package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/ramayac/multi-cmd/internal/models"
)

// Execute runs the selected commands on the selected folders
func Execute(folders []models.Folder, commands []models.Command) []models.ExecutionResult {
	var results []models.ExecutionResult

	for _, folder := range folders {
		if !folder.Selected {
			continue
		}

		for _, cmd := range commands {
			result := ExecuteCommand(folder, cmd)
			results = append(results, result)
		}
	}

	return results
}

func ExecuteCommand(folder models.Folder, command models.Command) models.ExecutionResult {
	// Build the full command string for display
	cmdString := command.Cmd
	if len(command.Args) > 0 {
		for _, arg := range command.Args {
			cmdString += " " + arg
		}
	}

	result := models.ExecutionResult{
		FolderName:      folder.Name,
		FolderPath:      folder.Path,
		CommandName:     command.Name,
		CommandExecuted: cmdString,
	}

	cmd := exec.Command(command.Cmd, command.Args...)
	cmd.Dir = folder.Path

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

	buf.WriteString("# Folder Command Results\n\n")

	currentFolder := ""
	for _, result := range results {
		if result.FolderName != currentFolder {
			currentFolder = result.FolderName
			buf.WriteString(fmt.Sprintf("## %s\n", result.FolderName))
			buf.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", result.FolderPath))
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
