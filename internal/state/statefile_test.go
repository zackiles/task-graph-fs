package state

import (
	"os"
	"testing"

	"github.com/company/task-graph-fs/internal/fsparse"
)

func TestStateFileOperations(t *testing.T) {
	// Create test state
	testState := &StateFile{
		Workflows: []WorkflowState{
			{
				WorkflowID: "workflow1",
				Status:     "completed",
				Tasks: []TaskState{
					{
						ID:           "taskA",
						Command:      "echo test",
						Dependencies: []string{},
						Priority:     "high",
						Retries:      2,
						Status:       "completed",
						Output:       "test output",
					},
				},
			},
		},
	}

	// Save state
	if err := testState.Save(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(StateFileName)

	// Load state
	loaded, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}

	// Compare states
	if len(loaded.Workflows) != len(testState.Workflows) {
		t.Errorf("expected %d workflows, got %d", len(testState.Workflows), len(loaded.Workflows))
	}

	if loaded.Workflows[0].WorkflowID != testState.Workflows[0].WorkflowID {
		t.Errorf("expected workflow ID %s, got %s",
			testState.Workflows[0].WorkflowID,
			loaded.Workflows[0].WorkflowID)
	}
}

func TestComputeDiff(t *testing.T) {
	currentState := &StateFile{
		Workflows: []WorkflowState{
			{
				WorkflowID: "existing",
				Status:     "completed",
				Tasks: []TaskState{
					{
						ID:      "task1",
						Command: "echo old",
					},
				},
			},
		},
	}

	newWorkflows := []fsparse.Workflow{
		{
			Name: "existing",
			Tasks: []fsparse.Task{
				{
					ID:      "task1",
					Command: "echo new",
				},
			},
		},
		{
			Name: "new",
			Tasks: []fsparse.Task{
				{
					ID:      "task2",
					Command: "echo test",
				},
			},
		},
	}

	added, updated, removed := currentState.ComputeDiff(newWorkflows)

	if len(added) != 1 || added[0] != "new" {
		t.Errorf("expected one addition 'new', got %v", added)
	}

	if len(updated) != 1 || updated[0] != "existing" {
		t.Errorf("expected one update 'existing', got %v", updated)
	}

	if len(removed) != 0 {
		t.Errorf("expected no removals, got %v", removed)
	}
}
