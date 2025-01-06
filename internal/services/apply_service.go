package services

import (
	"context"
	"fmt"

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

func (s *ApplyService) Plan(opts ApplyOptions) (*ApplyResult, error) {
	workflows, err := s.parser.ParseWorkflows(opts.WorkflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflows: %w", err)
	}

	currentState, err := state.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	added, updated, removed := currentState.ComputeDiff(workflows)
	return &ApplyResult{
		Added:      added,
		Updated:    updated,
		Removed:    removed,
		HasChanges: len(added)+len(updated)+len(removed) > 0,
	}, nil
}

func (s *ApplyService) Apply(ctx context.Context, opts ApplyOptions) error {
	workflows, err := s.parser.ParseWorkflows(opts.WorkflowDir)
	if err != nil {
		return fmt.Errorf("failed to parse workflows: %w", err)
	}

	newState := &state.StateFile{}

	for _, workflow := range workflows {
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
			workflowState.Status = "failed"
			return fmt.Errorf("workflow %s failed: %w", workflow.Name, err)
		}
		workflowState.Status = "completed"

		newState.Workflows = append(newState.Workflows, workflowState)
	}

	if err := newState.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}
