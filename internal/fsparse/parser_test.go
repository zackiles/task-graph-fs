package fsparse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorkflows(t *testing.T) {
	// Create test directory structure
	testDir := t.TempDir()

	// Create workflow directory
	workflowDir := filepath.Join(testDir, "workflow1")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create task files
	taskA := `# TaskA
## Command
echo "task A"
## Dependencies
None
## Priority
high
## Retries
2
## Timeout
5m`

	taskB := `# TaskB
## Command
echo "task B"
## Dependencies
TaskA
## Priority
medium
## Retries
1
## Timeout
10m`

	if err := os.WriteFile(filepath.Join(workflowDir, "taskA.md"), []byte(taskA), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "taskB.md"), []byte(taskB), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create dependency symlink
	if err := os.Symlink(
		filepath.Join(workflowDir, "taskA.md"),
		filepath.Join(workflowDir, "taskB_dependencies"),
	); err != nil {
		t.Fatal(err)
	}

	// Parse workflows
	workflows, err := ParseWorkflows(testDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results
	if len(workflows) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(workflows))
	}

	workflow := workflows[0]
	if workflow.Name != "workflow1" {
		t.Errorf("expected workflow name 'workflow1', got '%s'", workflow.Name)
	}

	if len(workflow.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(workflow.Tasks))
	}

	deps, ok := workflow.Dependencies["taskB"]
	if !ok {
		t.Error("expected dependencies for taskB")
	}
	if len(deps) != 1 || deps[0] != "taskA" {
		t.Errorf("expected taskB to depend on taskA, got %v", deps)
	}
}

func TestParseNestedWorkflows(t *testing.T) {
	// Create test directory structure
	testDir := t.TempDir()

	// Create nested workflow directories
	nestedPath := filepath.Join(testDir, "parent", "child", "workflow")
	if err := os.MkdirAll(nestedPath, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create task file in nested workflow
	taskContent := `# NestedTask
## Command
echo "nested task"
## Dependencies
None
## Priority
high
## Retries
2
## Timeout
5m`

	if err := os.WriteFile(filepath.Join(nestedPath, "task.md"), []byte(taskContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Parse workflows
	workflows, err := ParseWorkflows(testDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results
	if len(workflows) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(workflows))
	}

	workflow := workflows[0]
	expectedName := filepath.Join("parent", "child", "workflow")
	if workflow.Name != expectedName {
		t.Errorf("expected workflow name '%s', got '%s'", expectedName, workflow.Name)
	}

	if len(workflow.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(workflow.Tasks))
	}
}
