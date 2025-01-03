package main

import (
	"fmt"

	"github.com/company/task-graph-fs/internal/fsparse"
	"github.com/company/task-graph-fs/internal/state"
	"github.com/spf13/cobra"
)

func newPlanCmd(parser *fsparse.Parser) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show planned changes to workflows and tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse current workflows using the injected parser
			workflows, err := parser.ParseWorkflows(".")
			if err != nil {
				return fmt.Errorf("failed to parse workflows: %w", err)
			}

			// Load existing state
			currentState, err := state.LoadState()
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}

			// Compute differences
			added, updated, removed := currentState.ComputeDiff(workflows)

			// Print plan header
			fmt.Println("TaskGraphFS Plan:\n")
			fmt.Println("Workflow actions are indicated with the following symbols:")
			fmt.Println("  + add (new workflow/task)")
			fmt.Println("  ~ update (modified workflow/task)")
			fmt.Println("  - remove (deleted workflow/task)\n")
			fmt.Println("The following statefile changes will be made:\n")

			// Print additions
			for _, name := range added {
				printWorkflowAddition(findWorkflow(workflows, name))
			}

			// Print updates
			for _, name := range updated {
				printWorkflowUpdate(findWorkflow(workflows, name), findStateWorkflow(currentState, name))
			}

			// Print removals
			for _, name := range removed {
				printWorkflowRemoval(findStateWorkflow(currentState, name))
			}

			// Print summary
			fmt.Println("\nPlan Summary:")
			fmt.Printf("- Workflows: %d to add, %d to update, %d to remove.\n",
				len(added), len(updated), len(removed))

			taskChanges := computeTaskChanges(workflows, currentState)
			fmt.Printf("- Tasks: %d to add, %d to update, %d to remove.\n",
				taskChanges.added, taskChanges.updated, taskChanges.removed)

			fmt.Println("\nRun `tgfs apply` to execute these changes.")
			return nil
		},
	}
}

type taskChangeCount struct {
	added   int
	updated int
	removed int
}

func computeTaskChanges(workflows []fsparse.Workflow, currentState *state.StateFile) taskChangeCount {
	changes := taskChangeCount{}

	// Create maps for quick lookups
	currentTasks := make(map[string]map[string]state.TaskState)
	for _, w := range currentState.Workflows {
		currentTasks[w.WorkflowID] = make(map[string]state.TaskState)
		for _, t := range w.Tasks {
			currentTasks[w.WorkflowID][t.ID] = t
		}
	}

	// Compare tasks in each workflow
	for _, w := range workflows {
		for _, t := range w.Tasks {
			if wTasks, exists := currentTasks[w.Name]; !exists {
				changes.added++
			} else if _, exists := wTasks[t.ID]; !exists {
				changes.added++
			} else if taskHasChanges(t, wTasks[t.ID]) {
				changes.updated++
			}
		}
	}

	// Count removed tasks
	for _, w := range currentState.Workflows {
		newWorkflow := findWorkflow(workflows, w.WorkflowID)
		if newWorkflow.Name == "" {
			changes.removed += len(w.Tasks)
			continue
		}

		for _, t := range w.Tasks {
			if !taskExistsInWorkflow(newWorkflow, t.ID) {
				changes.removed++
			}
		}
	}

	return changes
}

func findWorkflow(workflows []fsparse.Workflow, name string) fsparse.Workflow {
	for _, w := range workflows {
		if w.Name == name {
			return w
		}
	}
	return fsparse.Workflow{}
}

func findStateWorkflow(state *state.StateFile, name string) state.WorkflowState {
	for _, w := range state.Workflows {
		if w.WorkflowID == name {
			return w
		}
	}
	return state.WorkflowState{}
}

func taskExistsInWorkflow(w fsparse.Workflow, taskID string) bool {
	for _, t := range w.Tasks {
		if t.ID == taskID {
			return true
		}
	}
	return false
}

func taskHasChanges(newTask fsparse.Task, currentTask state.TaskState) bool {
	return newTask.Command != currentTask.Command ||
		!stringSlicesEqual(newTask.Dependencies, currentTask.Dependencies) ||
		newTask.Priority != currentTask.Priority ||
		newTask.Retries != currentTask.Retries ||
		newTask.Timeout != currentTask.Timeout
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
