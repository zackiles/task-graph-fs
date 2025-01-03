package apply

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/company/task-graph-fs/internal/fsparse"
	"github.com/company/task-graph-fs/internal/state"
)

type Orchestrator struct {
	workflow    fsparse.Workflow
	state       *state.WorkflowState
	inProgress  sync.Map
	maxWorkers  int
	baseRetryMs int
}

func NewOrchestrator(workflow fsparse.Workflow, state *state.WorkflowState) *Orchestrator {
	return &Orchestrator{
		workflow:    workflow,
		state:       state,
		maxWorkers:  4,
		baseRetryMs: 1000,
	}
}

func (o *Orchestrator) Execute(ctx context.Context) error {
	// Create dependency graph
	graph := buildDependencyGraph(o.workflow)

	// Create worker pool
	tasks := make(chan *fsparse.Task, len(o.workflow.Tasks))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < o.maxWorkers; i++ {
		wg.Add(1)
		go o.worker(ctx, tasks, &wg)
	}

	// Schedule tasks that have no dependencies
	readyTasks := graph.getTasksWithNoDeps()
	for _, task := range readyTasks {
		tasks <- task
	}

	// Monitor task completion and schedule new tasks
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				o.inProgress.Range(func(key, value interface{}) bool {
					taskID := key.(string)
					status := value.(string)
					if status == "completed" {
						// Get and schedule tasks that were waiting on this one
						nextTasks := graph.getUnblockedTasks(taskID)
						for _, task := range nextTasks {
							tasks <- task
						}
						o.inProgress.Delete(taskID)
					}
					return true
				})
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	close(tasks)
	return nil
}

func (o *Orchestrator) worker(ctx context.Context, tasks <-chan *fsparse.Task, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		o.inProgress.Store(task.ID, "running")

		err := o.executeWithRetries(ctx, task)
		if err != nil {
			o.updateTaskState(task, "failed", err.Error())
			continue
		}

		o.updateTaskState(task, "completed", "")
		o.inProgress.Store(task.ID, "completed")
	}
}

func (o *Orchestrator) executeWithRetries(ctx context.Context, task *fsparse.Task) error {
	var lastErr error
	for attempt := 0; attempt <= task.Retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(o.baseRetryMs*(1<<attempt)) * time.Millisecond
			time.Sleep(backoff)
		}

		// Parse timeout from task
		timeout, err := time.ParseDuration(task.Timeout)
		if err != nil {
			timeout = 30 * time.Minute // default timeout
		}

		// Create context with timeout
		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Execute command
		cmd := exec.CommandContext(execCtx, "sh", "-c", task.Command)
		output, err := cmd.CombinedOutput()

		if err == nil {
			o.updateTaskState(task, "running", string(output))
			return nil
		}

		lastErr = fmt.Errorf("attempt %d failed: %w\noutput: %s", attempt+1, err, string(output))
		o.updateTaskState(task, "retrying", lastErr.Error())
	}

	return lastErr
}

func (o *Orchestrator) updateTaskState(task *fsparse.Task, status, output string) {
	for i, t := range o.state.Tasks {
		if t.ID == task.ID {
			o.state.Tasks[i].Status = status
			o.state.Tasks[i].Output = output
			return
		}
	}
}

type dependencyGraph struct {
	nodes map[string]*graphNode
}

type graphNode struct {
	task         *fsparse.Task
	dependencies map[string]bool
	dependents   map[string]bool
}

func buildDependencyGraph(workflow fsparse.Workflow) *dependencyGraph {
	graph := &dependencyGraph{
		nodes: make(map[string]*graphNode),
	}

	// Create nodes
	for _, task := range workflow.Tasks {
		task := task // Create new variable for closure
		graph.nodes[task.ID] = &graphNode{
			task:         &task,
			dependencies: make(map[string]bool),
			dependents:   make(map[string]bool),
		}
	}

	// Add dependencies
	for taskID, deps := range workflow.Dependencies {
		for _, dep := range deps {
			graph.nodes[taskID].dependencies[dep] = true
			graph.nodes[dep].dependents[taskID] = true
		}
	}

	return graph
}

func (g *dependencyGraph) getTasksWithNoDeps() []*fsparse.Task {
	var tasks []*fsparse.Task
	for _, node := range g.nodes {
		if len(node.dependencies) == 0 {
			tasks = append(tasks, node.task)
		}
	}
	return tasks
}

func (g *dependencyGraph) getUnblockedTasks(completedTaskID string) []*fsparse.Task {
	var tasks []*fsparse.Task
	for depTaskID := range g.nodes[completedTaskID].dependents {
		node := g.nodes[depTaskID]
		delete(node.dependencies, completedTaskID)
		if len(node.dependencies) == 0 {
			tasks = append(tasks, node.task)
		}
	}
	return tasks
}
