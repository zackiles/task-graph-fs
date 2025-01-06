package orchestration

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/state"
)

type Orchestrator struct {
	workflow   *fsparse.Workflow
	state      *state.WorkflowState
	inProgress sync.Map
}

func NewOrchestrator(workflow fsparse.Workflow, state *state.WorkflowState) *Orchestrator {
	return &Orchestrator{
		workflow: &workflow,
		state:    state,
	}
}

func (o *Orchestrator) Execute(ctx context.Context) error {
	// Create a new context with cancellation for task management
	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create error channel for collecting task errors
	errChan := make(chan error, len(o.workflow.Tasks))
	var wg sync.WaitGroup

	// Start task execution
	for i, task := range o.workflow.Tasks {
		select {
		case <-taskCtx.Done():
			return taskCtx.Err()
		default:
			wg.Add(1)
			o.inProgress.Store(task.ID, struct{}{})
			go func(taskIndex int, t fsparse.Task) {
				defer wg.Done()
				defer o.inProgress.Delete(t.ID)

				if err := o.executeTask(taskCtx, t); err != nil {
					errChan <- fmt.Errorf("task %s failed: %w", t.ID, err)
					cancel() // Now cancel() is defined
				}
			}(i, task)
		}
	}

	// Wait for all tasks to complete or context to be cancelled
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-taskCtx.Done():
		return taskCtx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}

func (o *Orchestrator) executeTask(ctx context.Context, task fsparse.Task) error {
	// Update task status
	for i := range o.state.Tasks {
		if o.state.Tasks[i].ID == task.ID {
			o.state.Tasks[i].Status = "running"
			break
		}
	}

	// Parse the timeout duration from the task
	timeout, err := time.ParseDuration(task.Timeout)
	if err != nil {
		timeout = 30 * time.Second // fallback to default timeout
	}

	// Create command with task-specific timeout context
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(taskCtx, "sh", "-c", task.Command)

	// Run command
	err = cmd.Run()

	// Update task status based on result
	for i := range o.state.Tasks {
		if o.state.Tasks[i].ID == task.ID {
			if err != nil {
				if taskCtx.Err() == context.DeadlineExceeded {
					o.state.Tasks[i].Status = "timeout"
					err = fmt.Errorf("task %s timed out after %s: %w", task.ID, timeout, err)
				} else {
					o.state.Tasks[i].Status = "failed"
				}
			} else {
				o.state.Tasks[i].Status = "completed"
			}
			break
		}
	}

	return err
}
