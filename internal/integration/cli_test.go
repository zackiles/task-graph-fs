package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zackiles/task-graph-fs/cmd"
	"github.com/zackiles/task-graph-fs/internal/gopilotcli"
	"github.com/zackiles/task-graph-fs/internal/state"
	"github.com/zackiles/task-graph-fs/internal/testutils"
)

type testEnv struct {
	rootDir     string
	mockGopilot *gopilotcli.MockGopilot
	ctx         context.Context
	cancel      context.CancelFunc
}

func setupTest(t *testing.T) *testEnv {
	// Create temp directory with absolute path
	rootDir, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Set up mock gopilot
	mockGopilot := gopilotcli.NewMockGopilot()
	gopilotcli.SetProvider(mockGopilot)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Store original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	if err := os.Chdir(rootDir); err != nil {
		t.Fatal(err)
	}

	// Set up cleanup
	t.Cleanup(func() {
		cancel()
		gopilotcli.Reset()
		os.Chdir(originalWd)
		time.Sleep(100 * time.Millisecond)
	})

	return &testEnv{
		rootDir:     rootDir,
		mockGopilot: mockGopilot,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func TestCLI(t *testing.T) {
	testutils.RunTestWithName(t, "Basic Init", func(t *testing.T) {
		env := setupTest(t)

		input := bytes.NewBufferString("test-workflow\n")
		oldStdin := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		go func() {
			_, err := io.Copy(w, input)
			if err != nil {
				t.Error(err)
			}
			w.Close()
		}()

		err = executeCommand(env.ctx, "init")
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = oldStdin

		workflowDir := filepath.Join(env.rootDir, "test-workflow")
		if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
			t.Fatal("workflow directory was not created")
		}
	})

	testutils.RunTestWithName(t, "Complex Workflow", func(t *testing.T) {
		env := setupTest(t)

		err := createTestWorkflow(env.rootDir, "complex", []string{"taskA", "taskB", "taskC"})
		if err != nil {
			t.Fatal(err)
		}

		err = createDependencyLink(env.rootDir, "complex", "taskB", "taskA")
		if err != nil {
			t.Fatal(err)
		}
		err = createDependencyLink(env.rootDir, "complex", "taskC", "taskB")
		if err != nil {
			t.Fatal(err)
		}

		for _, task := range []string{"taskA", "taskB", "taskC"} {
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

		if err := executeCommand(env.ctx, "plan"); err != nil {
			t.Fatal(err)
		}

		if err := executeCommand(env.ctx, "apply", "--auto-approve"); err != nil {
			t.Fatal(err)
		}

		state, err := state.LoadState(env.ctx)
		if err != nil {
			t.Fatal(err)
		}

		verifyWorkflowState(t, state, "complex", "completed")
		for _, task := range []string{"taskA", "taskB", "taskC"} {
			verifyTaskState(t, state, "complex", task, "completed")
		}
	})

	testutils.RunTestWithName(t, "Error Handling", func(t *testing.T) {
		env := setupTest(t)

		err := createTestWorkflow(env.rootDir, "failing", []string{"failingTask"})
		if err != nil {
			t.Fatal(err)
		}

		taskPath := filepath.Join(env.rootDir, "failing", "failingTask.md")
		absPath, err := filepath.Abs(taskPath)
		if err != nil {
			t.Fatal(err)
		}

		env.mockGopilot.SetResponse(
			absPath,
			gopilotcli.TaskResponse{
				Command:      "exit 1",
				Dependencies: []string{},
				Priority:     "high",
				Retries:      0,
				Timeout:      "1m",
			},
		)

		err = executeCommand(env.ctx, "apply", "--auto-approve")
		if err == nil {
			t.Fatal("expected apply command to fail")
		}
		if !strings.Contains(err.Error(), "exit status 1") {
			t.Fatalf("expected exit status 1 error, got: %v", err)
		}
	})

	testutils.RunTestWithName(t, "Concurrent Workflows", func(t *testing.T) {
		env := setupTest(t)

		workflows := []string{"workflow1", "workflow2", "workflow3"}
		for _, w := range workflows {
			err := createTestWorkflow(env.rootDir, w, []string{"taskA"})
			if err != nil {
				t.Fatal(err)
			}

			taskPath := filepath.Join(env.rootDir, w, "taskA.md")
			absPath, err := filepath.Abs(taskPath)
			if err != nil {
				t.Fatal(err)
			}

			env.mockGopilot.SetResponse(
				absPath,
				gopilotcli.TaskResponse{
					Command:      fmt.Sprintf("sleep 0.5 && echo %s", w),
					Dependencies: []string{},
					Priority:     "medium",
					Retries:      1,
					Timeout:      "5m",
				},
			)
		}

		start := time.Now()
		if err := executeCommand(env.ctx, "apply", "--auto-approve"); err != nil {
			t.Fatal(err)
		}
		duration := time.Since(start)

		if duration >= 2*time.Second {
			t.Errorf("workflows appear to run sequentially, took %v", duration)
		}

		state, err := state.LoadState(env.ctx)
		if err != nil {
			t.Fatal(err)
		}

		for _, w := range workflows {
			verifyWorkflowState(t, state, w, "completed")
		}
	})
}

// Helper function to execute CLI commands in tests with better cancellation
func executeCommand(ctx context.Context, args ...string) error {
	cmd := cmd.NewRootCommand()
	cmd.SetContext(ctx)
	cmd.SetArgs(args)

	// Create channel to track command completion
	done := make(chan error, 1)

	// Run command in goroutine
	go func() {
		done <- cmd.Execute()
		close(done)
	}()

	// Wait for either command completion or context cancellation
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Give the command a moment to cleanup before returning
		select {
		case <-done:
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return fmt.Errorf("command failed to shutdown gracefully: %w", ctx.Err())
		}
	}
}

// Debug wrapper for MockGopilot
type debugMockGopilot struct {
	*gopilotcli.MockGopilot
	t *testing.T
}

func (d *debugMockGopilot) GenerateTaskProps(ctx context.Context, taskPath string) (string, []string, string, int, string, error) {
	absPath, err := filepath.Abs(taskPath)
	if err != nil {
		d.t.Logf("Error getting absolute path: %v", err)
		return "", nil, "", 0, "", err
	}

	d.t.Logf("Mock called with path: %s (abs: %s)", taskPath, absPath)
	cmd, deps, pri, ret, timeout, err := d.MockGopilot.GenerateTaskProps(ctx, absPath)
	d.t.Logf("Mock returned: cmd=%s, deps=%v, pri=%s, ret=%d, timeout=%s, err=%v",
		cmd, deps, pri, ret, timeout, err)
	return cmd, deps, pri, ret, timeout, err
}
