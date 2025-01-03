package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const taskTemplate = `# %s

## Command
python example_script.py

## Dependencies
None

## Priority
medium

## Retries
1

## Timeout
30m
`

func newInitCmd() *cobra.Command {
	var workflowName, taskName string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new workflow and task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if workflowName == "" || taskName == "" {
				return fmt.Errorf("workflow name and task name are required")
			}

			// Create workflow directory
			if err := os.MkdirAll(workflowName, 0o755); err != nil {
				return fmt.Errorf("failed to create workflow directory: %w", err)
			}

			// Create task file
			taskPath := filepath.Join(workflowName, taskName+".md")
			content := fmt.Sprintf(taskTemplate, taskName)
			if err := os.WriteFile(taskPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("failed to create task file: %w", err)
			}

			fmt.Printf("Successfully initialized workflow '%s' with task '%s'\n", workflowName, taskName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "Name of the workflow")
	cmd.Flags().StringVarP(&taskName, "task", "t", "", "Name of the task")
	cmd.MarkFlagRequired("workflow")
	cmd.MarkFlagRequired("task")

	return cmd
}
