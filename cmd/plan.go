package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/printutils"
	"github.com/zackiles/task-graph-fs/internal/state"
)

// NewPlanCmd creates and returns the "plan" command.
func NewPlanCmd(parser *fsparse.Parser) *cobra.Command {
	var workflowDir string

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Show planned changes to workflows and tasks",
		Long: `The "plan" command analyzes the workflows in the specified directory,
compares them with the current state, and displays a summary of the changes
that would be made by running "apply".`,
		Args: cobra.NoArgs, // Ensure no unexpected positional arguments
		RunE: func(planCmd *cobra.Command, args []string) error {
			return runPlan(parser, workflowDir)
		},
	}

	planCmd.PersistentFlags().StringVarP(&workflowDir, "dir", "d", ".", "Directory containing workflows")
	return planCmd
}

// runPlan contains the logic for the "plan" command.
func runPlan(parser *fsparse.Parser, workflowDir string) error {
	// Parse current workflows using the injected parser
	workflows, err := parser.ParseWorkflows(workflowDir)
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

	// Print plan details
	printPlanSummary(workflows, currentState, added, updated, removed)
	return nil
}

func printPlanSummary(
	workflows []fsparse.Workflow,
	currentState *state.StateFile,
	added, updated, removed []string,
) {
	fmt.Println("TaskGraphFS Plan:\n")
	fmt.Println("Workflow actions are indicated with the following symbols:")
	fmt.Println("  + add (new workflow/task)")
	fmt.Println("  ~ update (modified workflow/task)")
	fmt.Println("  - remove (deleted workflow/task)\n")
	fmt.Println("The following statefile changes will be made:\n")

	for _, name := range added {
		workflow := findWorkflow(workflows, name)
		printutils.PrintWorkflowAddition(workflow)
	}

	for _, name := range updated {
		newWorkflow := findWorkflow(workflows, name)
		currentWorkflow := findStateWorkflow(currentState, name)
		printutils.PrintWorkflowUpdate(newWorkflow, currentWorkflow)
	}

	for _, name := range removed {
		removedWorkflow := findStateWorkflow(currentState, name)
		printutils.PrintWorkflowRemoval(removedWorkflow)
	}

	taskChanges := computeTaskChanges(workflows, currentState)
	fmt.Println("\nPlan Summary:")
	fmt.Printf("- Workflows: %d to add, %d to update, %d to remove.\n", len(added), len(updated), len(removed))
	fmt.Printf("- Tasks: %d to add, %d to update, %d to remove.\n", taskChanges.added, taskChanges.updated, taskChanges.removed)
	fmt.Println("\nRun `tgfs apply` to execute these changes.")
}

type taskChangeCount struct {
	added   int
	updated int
	removed int
}

func computeTaskChanges(workflows []fsparse.Workflow, currentState *state.StateFile) taskChangeCount {
	changes := taskChangeCount{}

	currentTasks := make(map[string]map[string]state.TaskState)
	for _, w := range currentState.Workflows {
		currentTasks[w.WorkflowID] = make(map[string]state.TaskState)
		for _, t := range w.Tasks {
			currentTasks[w.WorkflowID][t.ID] = t
		}
	}

	for _, w := range workflows {
		for _, t := range w.Tasks {
			if wTasks, exists := currentTasks[w.Name]; !exists {
				changes.added++
			} else if _, exists := wTasks[t.ID]; !exists {
				changes.added++
			} else if printutils.TaskHasChanges(t, wTasks[t.ID]) {
				changes.updated++
			}
		}
	}

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

func findStateWorkflow(state *state.StateFile, name string) *state.WorkflowState {
	for i := range state.Workflows {
		if state.Workflows[i].WorkflowID == name {
			return &state.Workflows[i]
		}
	}
	return nil
}

func taskExistsInWorkflow(w fsparse.Workflow, taskID string) bool {
	for _, t := range w.Tasks {
		if t.ID == taskID {
			return true
		}
	}
	return false
}
