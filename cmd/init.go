package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Constants for task template.
const taskTemplate = `# Example Task

## Command
python example_script.py

## Dependencies
None

## Priority
medium

## Retries
1

## Timeout
30m`

// NewInitCmd creates and returns the "init" command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new workflow",
		Long: `The "init" command allows you to create a new workflow directory
with a sanitized name and an example task file.`,
		Args: cobra.NoArgs, // No positional arguments are expected
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

// runInit contains the logic for the "init" command.
func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter workflow name: ")
	workflowName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read workflow name: %w", err)
	}

	workflowName = sanitizeWorkflowName(strings.TrimSpace(workflowName))
	if workflowName == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}

	return createWorkflow(workflowName)
}

// sanitizeWorkflowName ensures the workflow name is valid and URL-safe.
func sanitizeWorkflowName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]+")
	return reg.ReplaceAllString(name, "")
}

// createWorkflow sets up a workflow directory with a template task file.
func createWorkflow(name string) error {
	if err := os.MkdirAll(name, 0o755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	taskPath := filepath.Join(name, "task.example.md")
	if err := os.WriteFile(taskPath, []byte(taskTemplate), 0o644); err != nil {
		return fmt.Errorf("failed to create task file: %w", err)
	}

	fmt.Printf("Successfully initialized workflow '%s' with example task\n", name)
	return nil
}
