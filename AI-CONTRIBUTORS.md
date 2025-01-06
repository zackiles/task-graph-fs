# AI Contributors Guide

## Purpose
TaskGraphFS is a filesystem-based task orchestration tool that allows users to define workflows using markdown files and symbolic links. It follows a Terraform-like plan/apply pattern, where workflows are first planned and then executed with proper dependency management and state tracking.

## Implementation

1. **Filesystem as Graph Structure**
   - Workflows are represented as directories
   - Tasks are markdown files with structured sections (Command, Dependencies, Priority, Retries, Timeout)
   - Dependencies are established through symbolic links (`taskB_dependencies -> taskA.md`)
   - State is tracked in `tgfs-state.json`

2. **Orchestration Engine**
   - Concurrent task execution with worker pools
   - Dependency graph resolution
   - Retry mechanism with exponential backoff
   - Task timeout management
   - State recovery after interruption

3. **AI Integration**
   - Uses GopilotCLI interface for task parsing
   - Supports both structured and natural language task definitions
   - Mockable for testing

## Task Definition
Tasks are defined in markdown files with the following structure:
```markdown
# ProcessData

## Command
python process_data.py

## Dependencies
None

## Priority
high

## Retries
2

## Timeout
30m
```

## State Management
The state file (`tgfs-state.json`) tracks workflow and task status:
```json
{
  "workflow_id": "MyWorkflow",
  "status": "in_progress",
  "tasks": [
    {
      "id": "TaskC",
      "command": "python task_c.py",
      "dependencies": [],
      "priority": "low",
      "retries": 0,
      "status": "completed",
      "duration": "2s",
      "output": "Initial setup completed."
    },
    {
      "id": "TaskB",
      "command": "python task_b.py",
      "dependencies": ["TaskC"],
      "priority": "medium",
      "retries": 1,
      "status": "completed",
      "duration": "3s",
      "output": "Generated intermediate results."
    }
  ]
}
```

## Error Handling
1. **Retry Mechanism**
   - Exponential backoff between retries
   - Configurable retry count per task
   - State preservation between attempts

2. **Timeout Management**
   - Per-task timeout configuration
   - Default 30-minute timeout
   - Clean termination of timed-out processes

3. **State Recovery**
   - Persistent state file
   - Recovery from interruption
   - Maintains task output history

## Structure
Example workflow structure:
```
workflows/
├── data-pipeline/
│   ├── fetch-data.md
│   ├── clean-data.md
│   ├── clean-data_dependencies -> fetch-data.md
│   ├── transform-data.md
│   ├── transform-data_dependencies -> clean-data.md
│   └── nested/
│       ├── aggregate-daily.md
│       ├── aggregate-daily_dependencies -> ../transform-data.md
│       ├── compute-metrics.md
│       └── compute-metrics_dependencies -> aggregate-daily.md
│
└── model-training/
    ├── prepare-features.md
    ├── prepare-features_dependencies -> ../data-pipeline/transform-data.md
    ├── train-model.md
    ├── train-model_dependencies -> prepare-features.md
    ├── evaluate-model.md
    └── evaluate-model_dependencies -> train-model.md
```

## Important Public Methods

### CLI Commands
1. `tgfs init`
   - Creates new workflow directory with example task
   - Args: None (interactive workflow name prompt)
   - Returns: Creates directory structure

2. `tgfs plan [--dir <path>]`
   - Shows planned workflow changes
   - Args: Optional workflow directory path
   - Returns: Detailed plan output:
```
TaskGraphFS Plan:

Workflow actions are indicated with the following symbols:
  + add (new workflow/task)
  ~ update (modified workflow/task)
  - remove (deleted workflow/task)

The following statefile changes will be made:

  + "workflows[0]" {
      "workflow_id": "NewWorkflow",
      "status": "pending",
      "tasks": [
        {
          "id": "TaskX",
          "command": "python task_x.py",
          "dependencies": [],
          "priority": "medium",
          "retries": 2,
          "status": "pending"
        }
      ]
    }
```

3. `tgfs apply [--auto-approve]`
   - Executes planned workflow changes
   - Args: Optional auto-approve flag
   - Returns: Execution results and updates state file

## Important Internal Methods

1. `Parser.ParseWorkflows(basePath string) ([]Workflow, error)`
   - Parses filesystem structure into workflow objects
   - Handles nested workflows and dependencies
   - Location: `internal/fsparse/parser.go`

2. `Orchestrator.Execute(ctx context.Context) error`
   - Executes workflow tasks with dependency ordering
   - Manages concurrent execution and retries
   - Location: `internal/apply/orchestrator.go`

3. `StateFile.ComputeDiff(workflows []fsparse.Workflow) (added, updated, removed []string)`
   - Computes changes between current and desired state
   - Location: `internal/state/state.go`

## Building & Testing

### Build Commands
```bash
make build              # Build production binary
make install           # Install globally
make dev               # Development build
make test              # Run all tests
make test-integration  # Run integration tests
make test-unit         # Run unit tests
make test-coverage     # Run unit tests with coverage
make test-race         # Run unit tests with race detection
make test-short        # Run unit tests without race detection
make test-verbose      # Run unit tests with verbose output
make test-watch        # Run unit tests and watch for changes
```

### Test Files
1. `internal/fsparse/parser_test.go`
   - Tests workflow parsing
   - Tests dependency resolution
   - Tests nested workflows

2. `internal/apply/orchestrator_test.go`
   - Tests task execution
   - Tests failure handling
   - Tests dependency ordering

3. `internal/integration/cli_test.go`
   - End-to-end CLI tests
   - Tests error scenarios
   - Tests concurrent workflows
   - Tests timeout handling
   - Tests state recovery

### Testing with Mock Gopilot
Example mock setup for integration tests:
```go
mockGopilot.SetResponse(
    filepath.Join(rootDir, "workflow", "task.md"),
    gopilotcli.TaskResponse{
        Command:      "echo test",
        Dependencies: []string{"taskA"},
        Priority:     "medium",
        Retries:      1,
        Timeout:      "5m",
    },
)
```

## TODOs

1. Actually implement real task parsing on `tgfs plan` which utilizes a real gopilot client to parse the tasks and return a structure task object representing what was defined in the unstructured markdown files written by a human.
2. Actually implement and test task running in `tgfs apply` so that we can trigger the tasks defined in the structured task objects returned by gopilot inthe `tgfs plan` command.
3. Add workflow versioning and rollback support
4. Decide on if a metafile or the statefile should be written to the root of the workspace.