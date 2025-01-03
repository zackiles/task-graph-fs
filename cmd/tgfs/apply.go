package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/company/task-graph-fs/internal/apply"
	"github.com/company/task-graph-fs/internal/fsparse"
	"github.com/company/task-graph-fs/internal/state"
	"github.com/spf13/cobra"
)

func newApplyCmd(parser *fsparse.Parser) *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the planned changes to workflows and tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse current workflows
			workflows, err := parser.ParseWorkflows(".")
			if err != nil {
				return fmt.Errorf("failed to parse workflows: %w", err)
			}

			// Load existing state
			currentState, err := state.LoadState()
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}

			// Show plan
			added, updated, removed := currentState.ComputeDiff(workflows)
			if len(added)+len(updated)+len(removed) == 0 {
				fmt.Println("No changes to apply")
				return nil
			}

			// Confirm unless auto-approve is set
			if !autoApprove {
				fmt.Printf("\nDo you want to apply these changes? [y/N] ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Apply cancelled")
					return nil
				}
			}

			// Create context with cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle interrupts
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\nReceived interrupt signal, gracefully shutting down...")
				cancel()
			}()

			// Apply changes
			newState := &state.StateFile{}

			// Process additions and updates
			for _, workflow := range workflows {
				workflowState := state.WorkflowState{
					WorkflowID: workflow.Name,
					Status:     "running",
					Tasks:      make([]state.TaskState, len(workflow.Tasks)),
				}

				// Initialize task states
				for i, task := range workflow.Tasks {
					workflowState.Tasks[i] = state.TaskState{
						ID:           task.ID,
						Command:      task.Command,
						Dependencies: task.Dependencies,
						Priority:     task.Priority,
						Retries:      task.Retries,
						Status:       "pending",
					}
				}

				// Execute workflow
				orchestrator := apply.NewOrchestrator(workflow, &workflowState)
				if err := orchestrator.Execute(ctx); err != nil {
					workflowState.Status = "failed"
					fmt.Printf("Workflow %s failed: %v\n", workflow.Name, err)
				} else {
					workflowState.Status = "completed"
				}

				newState.Workflows = append(newState.Workflows, workflowState)
			}

			// Save final state
			if err := newState.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			fmt.Println("\nApply complete!")
			return nil
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	return cmd
}
