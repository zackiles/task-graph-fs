package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/orchestration"
	"github.com/zackiles/task-graph-fs/internal/state"
)

type ApplyService struct {
	parser *fsparse.Parser
}

func NewApplyService(parser *fsparse.Parser) *ApplyService {
	return &ApplyService{
		parser: parser,
	}
}

type ApplyOptions struct {
	WorkflowDir string
	AutoApprove bool
}

type ApplyResult struct {
	Added      []string
	Updated    []string
	Removed    []string
	HasChanges bool
}

func (s *ApplyService) Plan(ctx context.Context, opts ApplyOptions) (*ApplyResult, error) {
	workflows, err := s.parser.ParseWorkflows(ctx, opts.WorkflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflows: %w", err)
	}

	currentState, err := state.LoadState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	added, updated, removed, err := currentState.ComputeDiff(ctx, workflows)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	return &ApplyResult{
		Added:      added,
		Updated:    updated,
		Removed:    removed,
		HasChanges: len(added)+len(updated)+len(removed) > 0,
	}, nil
}

func (s *ApplyService) Apply(ctx context.Context, opts ApplyOptions) error {
	// Create a new context with timeout for the entire apply operation
	// Use a shorter timeout for tests
	timeout := 30 * time.Second
	if os.Getenv("TEST_ENV") == "true" {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	workflows, err := s.parser.ParseWorkflows(ctx, opts.WorkflowDir)
	if err != nil {
		return fmt.Errorf("failed to parse workflows: %w", err)
	}

	newState := &state.StateFile{}

	for _, workflow := range workflows {
		// Check context before starting each workflow
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		workflowState := state.WorkflowState{
			WorkflowID: workflow.Name,
			Status:     "running",
			Tasks:      make([]state.TaskState, len(workflow.Tasks)),
		}

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

		orchestrator := orchestration.NewOrchestrator(workflow, &workflowState)
		if err := orchestrator.Execute(ctx); err != nil {
			if err == context.Canceled {
				workflowState.Status = "cancelled"
			} else {
				workflowState.Status = "failed"
			}
			// Save partial state before returning
			_ = newState.Save(ctx)
			return fmt.Errorf("workflow %s failed: %w", workflow.Name, err)
		}

		workflowState.Status = "completed"
		newState.Workflows = append(newState.Workflows, workflowState)
	}

	// Final state save
	if err := newState.Save(ctx); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}
