package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

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

func sanitizeWorkflowName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]+")
	return reg.ReplaceAllString(name, "")
}

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new workflow",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
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
