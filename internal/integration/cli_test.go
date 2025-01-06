package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zackiles/task-graph-fs/internal/commands"
	"github.com/zackiles/task-graph-fs/internal/gopilotcli"
	"github.com/zackiles/task-graph-fs/internal/state"
)

type testEnv struct {
	rootDir     string
	mockGopilot *gopilotcli.MockGopilot
}

func setupTest(t *testing.T) *testEnv {
	rootDir := t.TempDir()
	mockGopilot := gopilotcli.NewMockGopilot()

	return &testEnv{
		rootDir:     rootDir,
		mockGopilot: mockGopilot,
	}
}

func TestEndToEnd(t *testing.T) {
	env := setupTest(t)

	// Test init command
	t.Run("init command", func(t *testing.T) {
		// Create a buffer to simulate stdin with a workflow name
		input := bytes.NewBufferString("My Workflow 123!\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe and pass input
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		go func() {
			io.Copy(w, input)
			w.Close()
		}()

		// Execute init command
		args := []string{"init"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Verify directory and file creation
		// Note: The workflow name should be sanitized to "mock-workflow1"
		taskPath := filepath.Join(env.rootDir, "mock-workflow1", "mock-task.md")
		if _, err := os.Stat(taskPath); os.IsNotExist(err) {
			t.Error("example task file was not created")
		}

		// Read the created file and verify its contents
		content, err := os.ReadFile(taskPath)
		if err != nil {
			t.Fatal(err)
		}

		expectedContent := `# Example Task

## Command
python example_script.py

## Dependencies
None

## Priority
medium

## Retries
1

## Timeout
30m`

		if strings.TrimSpace(string(content)) != strings.TrimSpace(expectedContent) {
			t.Error("task file content does not match expected template")
		}
	})

	// Set up mock responses for plan and apply
	env.mockGopilot.SetResponse(
		filepath.Join(env.rootDir, "workflow1", "taskA.md"),
		gopilotcli.TaskResponse{
			Command:      "echo test",
			Dependencies: []string{},
			Priority:     "high",
			Retries:      2,
			Timeout:      "5m",
		},
	)

	// Test plan command
	t.Run("plan command", func(t *testing.T) {
		args := []string{"plan"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Verify plan output file exists
		if _, err := os.Stat(filepath.Join(env.rootDir, ".tgfs-plan")); os.IsNotExist(err) {
			t.Error("plan file was not created")
		}
	})

	// Test apply command
	t.Run("apply command", func(t *testing.T) {
		args := []string{"apply", "--auto-approve"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Verify state file exists and contains expected content
		stateFile := filepath.Join(env.rootDir, "tgfs-state.json")
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			t.Error("state file was not created")
		}

		// Verify task execution results
		state, err := loadState(stateFile)
		if err != nil {
			t.Fatal(err)
		}

		if len(state.Workflows) != 1 {
			t.Errorf("expected 1 workflow, got %d", len(state.Workflows))
		}

		workflow := state.Workflows[0]
		if workflow.Status != "completed" {
			t.Errorf("expected workflow status 'completed', got '%s'", workflow.Status)
		}

		if len(workflow.Tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(workflow.Tasks))
		}

		task := workflow.Tasks[0]
		if task.Status != "completed" {
			t.Errorf("expected task status 'completed', got '%s'", task.Status)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	env := setupTest(t)

	// Set up mock response with error
	env.mockGopilot.SetResponse(
		filepath.Join(env.rootDir, "workflow1", "taskA.md"),
		gopilotcli.TaskResponse{
			Error: fmt.Errorf("gopilot CLI error"),
		},
	)

	// Test error handling in plan
	t.Run("plan handles gopilot error", func(t *testing.T) {
		args := []string{"plan"}
		err := executeCommand(args...)
		if err == nil {
			t.Error("expected error from plan command")
		}
		if !strings.Contains(err.Error(), "gopilot CLI error") {
			t.Errorf("expected gopilot error, got: %v", err)
		}
	})
}

func TestComplexWorkflow(t *testing.T) {
	env := setupTest(t)

	// Create a workflow with multiple tasks and dependencies
	t.Run("complex workflow setup", func(t *testing.T) {
		// Create workflow with three tasks
		err := createTestWorkflow(env.rootDir, "complex", []string{"taskA", "taskB", "taskC"})
		if err != nil {
			t.Fatal(err)
		}

		// Create dependencies: B depends on A, C depends on B
		err = createDependencyLink(env.rootDir, "complex", "taskB", "taskA")
		if err != nil {
			t.Fatal(err)
		}
		err = createDependencyLink(env.rootDir, "complex", "taskC", "taskB")
		if err != nil {
			t.Fatal(err)
		}

		// Set up mock responses
		tasks := []string{"taskA", "taskB", "taskC"}
		for _, task := range tasks {
			env.mockGopilot.SetResponse(
				filepath.Join(env.rootDir, "complex", task+".md"),
				gopilotcli.TaskResponse{
					Command:      fmt.Sprintf("echo %s", task),
					Dependencies: []string{},
					Priority:     "medium",
					Retries:      1,
					Timeout:      "5m",
				},
			)
		}

		// Test plan
		args := []string{"plan"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Test apply
		args = []string{"apply", "--auto-approve"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Verify final state
		state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
		if err != nil {
			t.Fatal(err)
		}

		verifyWorkflowState(t, state, "complex", "completed")
		verifyTaskState(t, state, "complex", "taskA", "completed")
		verifyTaskState(t, state, "complex", "taskB", "completed")
		verifyTaskState(t, state, "complex", "taskC", "completed")
	})
}

func TestErrorHandlingScenarios(t *testing.T) {
	env := setupTest(t)

	t.Run("failing task", func(t *testing.T) {
		// Create workflow with failing task
		err := createTestWorkflow(env.rootDir, "failing", []string{"failingTask"})
		if err != nil {
			t.Fatal(err)
		}

		// Set up mock to simulate failure
		env.mockGopilot.SetResponse(
			filepath.Join(env.rootDir, "failing", "failingTask.md"),
			gopilotcli.TaskResponse{
				Command:      "exit 1",
				Dependencies: []string{},
				Priority:     "high",
				Retries:      0,
				Timeout:      "1m",
			},
		)

		// Test apply
		args := []string{"apply", "--auto-approve"}
		err = executeCommand(args...)
		if err != nil {
			t.Fatal(err)
		}

		// Verify failed state
		state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
		if err != nil {
			t.Fatal(err)
		}

		verifyWorkflowState(t, state, "failing", "failed")
		verifyTaskState(t, state, "failing", "failingTask", "failed")
	})

	t.Run("retry mechanism", func(t *testing.T) {
		// Create workflow with retrying task
		err := createTestWorkflow(env.rootDir, "retry", []string{"retryTask"})
		if err != nil {
			t.Fatal(err)
		}

		// Set up mock to simulate retry scenario
		env.mockGopilot.SetResponse(
			filepath.Join(env.rootDir, "retry", "retryTask.md"),
			gopilotcli.TaskResponse{
				Command:      "echo retry",
				Dependencies: []string{},
				Priority:     "high",
				Retries:      2,
				Timeout:      "1m",
			},
		)

		// Test apply
		args := []string{"apply", "--auto-approve"}
		err = executeCommand(args...)
		if err != nil {
			t.Fatal(err)
		}

		// Verify final state
		state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
		if err != nil {
			t.Fatal(err)
		}

		verifyWorkflowState(t, state, "retry", "completed")
		verifyTaskState(t, state, "retry", "retryTask", "completed")
	})
}

func TestConcurrentWorkflows(t *testing.T) {
	env := setupTest(t)

	// Create multiple independent workflows
	workflows := []string{"workflow1", "workflow2", "workflow3"}
	for _, w := range workflows {
		err := createTestWorkflow(env.rootDir, w, []string{"taskA"})
		if err != nil {
			t.Fatal(err)
		}

		env.mockGopilot.SetResponse(
			filepath.Join(env.rootDir, w, "taskA.md"),
			gopilotcli.TaskResponse{
				Command:      fmt.Sprintf("sleep 1 && echo %s", w),
				Dependencies: []string{},
				Priority:     "medium",
				Retries:      1,
				Timeout:      "5m",
			},
		)
	}

	// Test concurrent execution
	args := []string{"apply", "--auto-approve"}
	start := time.Now()
	if err := executeCommand(args...); err != nil {
		t.Fatal(err)
	}
	duration := time.Since(start)

	// Verify that workflows ran concurrently (total time should be less than sequential execution)
	if duration >= 3*time.Second {
		t.Errorf("workflows appear to run sequentially, took %v", duration)
	}

	// Verify all workflows completed
	state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
	if err != nil {
		t.Fatal(err)
	}

	for _, w := range workflows {
		verifyWorkflowState(t, state, w, "completed")
	}
}

func TestTimeoutHandling(t *testing.T) {
	env := setupTest(t)

	err := createTestWorkflow(env.rootDir, "timeout", []string{"longTask"})
	if err != nil {
		t.Fatal(err)
	}

	env.mockGopilot.SetResponse(
		filepath.Join(env.rootDir, "timeout", "longTask.md"),
		gopilotcli.TaskResponse{
			Command:      "sleep 5", // Task takes 5 seconds
			Dependencies: []string{},
			Priority:     "high",
			Retries:      0,
			Timeout:      "1s", // But timeout is 1 second
		},
	)

	args := []string{"apply", "--auto-approve"}
	err = executeCommand(args...)
	if err != nil {
		t.Fatal(err)
	}

	state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
	if err != nil {
		t.Fatal(err)
	}

	verifyWorkflowState(t, state, "timeout", "failed")
	verifyTaskState(t, state, "timeout", "longTask", "failed")
}

func TestStateRecovery(t *testing.T) {
	env := setupTest(t)

	// Create initial workflow
	err := createTestWorkflow(env.rootDir, "recovery", []string{"taskA", "taskB"})
	if err != nil {
		t.Fatal(err)
	}

	// Set up initial state with taskA completed
	initialState := &state.StateFile{
		Workflows: []state.WorkflowState{
			{
				WorkflowID: "recovery",
				Status:     "running",
				Tasks: []state.TaskState{
					{
						ID:      "taskA",
						Command: "echo taskA",
						Status:  "completed",
					},
					{
						ID:      "taskB",
						Command: "echo taskB",
						Status:  "pending",
					},
				},
			},
		},
	}

	// Save initial state
	stateData, _ := json.MarshalIndent(initialState, "", "  ")
	if err := os.WriteFile(filepath.Join(env.rootDir, "tgfs-state.json"), stateData, 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up mock responses
	env.mockGopilot.SetResponse(
		filepath.Join(env.rootDir, "recovery", "taskB.md"),
		gopilotcli.TaskResponse{
			Command:      "echo taskB",
			Dependencies: []string{"taskA"},
			Priority:     "medium",
			Retries:      1,
			Timeout:      "5m",
		},
	)

	// Run apply
	args := []string{"apply", "--auto-approve"}
	if err := executeCommand(args...); err != nil {
		t.Fatal(err)
	}

	// Verify final state
	finalState, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify taskA remained completed and taskB was executed
	verifyWorkflowState(t, finalState, "recovery", "completed")
	verifyTaskState(t, finalState, "recovery", "taskA", "completed")
	verifyTaskState(t, finalState, "recovery", "taskB", "completed")
}

func TestNestedWorkflows(t *testing.T) {
	env := setupTest(t)

	// Create nested workflow structure
	nestedWorkflows := []struct {
		path string
		task string
	}{
		{filepath.Join("parent", "child1", "workflow"), "taskA"},
		{filepath.Join("parent", "child2", "workflow"), "taskB"},
		{filepath.Join("parent", "child1", "child2", "workflow"), "taskC"},
	}

	for _, nw := range nestedWorkflows {
		// Create workflow directory
		workflowPath := filepath.Join(env.rootDir, nw.path)
		if err := os.MkdirAll(workflowPath, 0o755); err != nil {
			t.Fatal(err)
		}

		// Create task file
		err := createTestWorkflow(workflowPath, "", []string{nw.task})
		if err != nil {
			t.Fatal(err)
		}

		// Set up mock response
		env.mockGopilot.SetResponse(
			filepath.Join(workflowPath, nw.task+".md"),
			gopilotcli.TaskResponse{
				Command:      fmt.Sprintf("echo %s", nw.task),
				Dependencies: []string{},
				Priority:     "medium",
				Retries:      1,
				Timeout:      "5m",
			},
		)
	}

	// Test plan command
	t.Run("plan nested workflows", func(t *testing.T) {
		args := []string{"plan"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}
	})

	// Test apply command
	t.Run("apply nested workflows", func(t *testing.T) {
		args := []string{"apply", "--auto-approve"}
		if err := executeCommand(args...); err != nil {
			t.Fatal(err)
		}

		// Verify state file
		state, err := loadState(filepath.Join(env.rootDir, "tgfs-state.json"))
		if err != nil {
			t.Fatal(err)
		}

		// Verify each nested workflow
		for _, nw := range nestedWorkflows {
			verifyWorkflowState(t, state, nw.path, "completed")
			verifyTaskState(t, state, nw.path, nw.task, "completed")
		}
	})
}

// Helper function to execute CLI commands in tests
func executeCommand(args ...string) error {
	cmd := commands.NewRootCommand()
	cmd.SetArgs(args)
	return cmd.Execute()
}
