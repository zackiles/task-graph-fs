package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/zackiles/task-graph-fs/internal/state"
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
	// Sanitize workflow name to match init behavior
	workflowName = sanitizeWorkflowName(workflowName)
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

func createDependencyLink(rootDir, workflowName, taskID, dependencyID string) error {
	workflowDir := filepath.Join(rootDir, workflowName)
	return os.Symlink(
		filepath.Join(workflowDir, dependencyID+".md"),
		filepath.Join(workflowDir, taskID+"_dependencies"),
	)
}

// verifyWorkflowState checks if a workflow has the expected status
func verifyWorkflowState(t *testing.T, stateFile *state.StateFile, workflowID, expectedStatus string) {
	var workflow *state.WorkflowState
	for i := range stateFile.Workflows {
		if stateFile.Workflows[i].WorkflowID == workflowID {
			workflow = &stateFile.Workflows[i]
			break
		}
	}

	if workflow == nil {
		t.Errorf("workflow %s not found", workflowID)
		return
	}

	if workflow.Status != expectedStatus {
		t.Errorf("expected workflow status %s, got %s", expectedStatus, workflow.Status)
	}
}

// verifyTaskState checks if a task has the expected status
func verifyTaskState(t *testing.T, stateFile *state.StateFile, workflowID, taskID, expectedStatus string) {
	var task *state.TaskState
	for _, w := range stateFile.Workflows {
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

// sanitizeWorkflowName sanitizes the workflow name
func sanitizeWorkflowName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]+")
	return reg.ReplaceAllString(name, "")
}
