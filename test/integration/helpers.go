package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/company/task-graph-fs/internal/state"
)

// loadState reads and parses the state file
func loadState(path string) (*state.StateFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state state.StateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// createTestWorkflow creates a test workflow with the given tasks
func createTestWorkflow(rootDir, workflowName string, tasks []string) error {
	workflowDir := filepath.Join(rootDir, workflowName)
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return err
	}

	for _, task := range tasks {
		content := []byte(fmt.Sprintf(`# %s

## Command
echo "test command"

## Dependencies
None

## Priority
medium

## Retries
1

## Timeout
30m
`, task))

		if err := os.WriteFile(
			filepath.Join(workflowDir, task+".md"),
			content,
			0o644,
		); err != nil {
			return err
		}
	}

	return nil
}

// createDependencyLink creates a symbolic link to represent task dependencies
func createDependencyLink(rootDir, workflowName, sourceTask, targetTask string) error {
	workflowDir := filepath.Join(rootDir, workflowName)
	return os.Symlink(
		filepath.Join(workflowDir, targetTask+".md"),
		filepath.Join(workflowDir, sourceTask+"_dependencies"),
	)
}

// verifyWorkflowState checks if the workflow state matches expected values
func verifyWorkflowState(t *testing.T, state *state.StateFile, workflowID string, expectedStatus string) {
	t.Helper()

	var workflow *state.WorkflowState
	for i := range state.Workflows {
		if state.Workflows[i].WorkflowID == workflowID {
			workflow = &state.Workflows[i]
			break
		}
	}

	if workflow == nil {
		t.Errorf("workflow %s not found in state", workflowID)
		return
	}

	if workflow.Status != expectedStatus {
		t.Errorf("expected workflow status %s, got %s", expectedStatus, workflow.Status)
	}
}

// verifyTaskState checks if a task's state matches expected values
func verifyTaskState(t *testing.T, state *state.StateFile, workflowID, taskID, expectedStatus string) {
	t.Helper()

	var task *state.TaskState
	for _, w := range state.Workflows {
		if w.WorkflowID == workflowID {
			for i := range w.Tasks {
				if w.Tasks[i].ID == taskID {
					task = &w.Tasks[i]
					break
				}
			}
		}
	}

	if task == nil {
		t.Errorf("task %s not found in workflow %s", taskID, workflowID)
		return
	}

	if task.Status != expectedStatus {
		t.Errorf("expected task status %s, got %s", expectedStatus, task.Status)
	}
}
