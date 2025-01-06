package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/state"
)

func TestOrchestrator(t *testing.T) {
	workflow := fsparse.Workflow{
		Name: "test",
		Tasks: []fsparse.Task{
			{
				ID:      "task1",
				Command: "echo success",
				Timeout: "1m",
				Retries: 1,
			},
			{
				ID:      "task2",
				Command: "echo dependent",
				Timeout: "1m",
				Retries: 1,
			},
		},
		Dependencies: map[string][]string{
			"task2": {"task1"},
		},
	}

	workflowState := &state.WorkflowState{
		WorkflowID: "test",
		Status:     "pending",
		Tasks: []state.TaskState{
			{
				ID:     "task1",
				Status: "pending",
			},
			{
				ID:     "task2",
				Status: "pending",
			},
		},
	}

	orchestrator := NewOrchestrator(workflow, workflowState)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orchestrator.Execute(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify execution order and status
	if workflowState.Tasks[0].Status != "completed" {
		t.Errorf("expected task1 status 'completed', got '%s'", workflowState.Tasks[0].Status)
	}

	if workflowState.Tasks[1].Status != "completed" {
		t.Errorf("expected task2 status 'completed', got '%s'", workflowState.Tasks[1].Status)
	}
}

func TestOrchestratorFailure(t *testing.T) {
	workflow := fsparse.Workflow{
		Name: "test",
		Tasks: []fsparse.Task{
			{
				ID:      "failing",
				Command: "exit 1",
				Timeout: "1m",
				Retries: 1,
			},
		},
	}

	workflowState := &state.WorkflowState{
		WorkflowID: "test",
		Status:     "pending",
		Tasks: []state.TaskState{
			{
				ID:     "failing",
				Status: "pending",
			},
		},
	}

	orchestrator := NewOrchestrator(workflow, workflowState)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orchestrator.Execute(ctx); err != nil {
		t.Fatal(err)
	}

	if workflowState.Tasks[0].Status != "failed" {
		t.Errorf("expected task status 'failed', got '%s'", workflowState.Tasks[0].Status)
	}
}
