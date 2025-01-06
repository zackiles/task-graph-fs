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
   - Dependency graph resolution via symlinks
   - Task execution with state tracking
   - Task timeout management
   - State recovery after interruption
   - Clean termination via context cancellation
   - Concurrent task execution with sync.WaitGroup
   - Task status tracking with sync.Map
   - Context-aware execution with cancellation
   - Task-specific timeout contexts
   - Error propagation through channels
   - Graceful shutdown on interrupts

3. **AI Integration**
   - Uses provider pattern for GopilotCLI interface:
     ```go
     // Provider singleton pattern
     var provider GopilotCLI = NewRealGopilot()

     // Support for mock injection in tests
     func SetProvider(p GopilotCLI)
     func GetProvider() GopilotCLI
     func Reset()
     ```
   - Supports both structured and natural language task definitions
   - Context-aware operations for cancellation support
   - Mockable for testing with path normalization

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

The Task object includes:
```go
type Task struct {
    ID           string   // Task identifier (filename without .md)
    MarkdownPath string   // Full path to markdown file
    Command      string   // Command to execute
    Dependencies []string // List of dependent task IDs
    Priority     string   // Task priority level
    Retries      int     // Number of retry attempts
    Timeout      string  // Task timeout duration
    Status       string  // Current execution status
    Output       string  // Task execution output
    Duration     string  // Task execution duration
}
```

## State Management
The state file (`tgfs-state.json`) tracks workflow and task status with context-aware operations:
- Context-aware save operations
- Atomic file writes
- Proper error handling with context cancellation
- State diffing with context support
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
   - Global apply timeout (30s default, 5s in test environment)
   - Per-task timeout configuration
   - Default 30-minute task timeout
   - Clean termination of timed-out processes
   - Context-aware cancellation

3. **State Recovery**
   - Persistent state file
   - Recovery from interruption
   - Maintains task output history

## Project File Structure
Example workflow structure:
```
project/
├── cmd/
│   ├── apply.go
│   ├── init.go
│   ├── plan.go
│   └── root.go
├── internal/
│   ├── fsparse/
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── types.go
│   ├── gopilotcli/
│   │   ├── mock.go
│   │   └── real.go
│   ├── integration/
│   │   ├── cli_test.go
│   │   └── helpers.go
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   └── orchestrator_test.go
│   ├── printutils/
│   │   ├── tasks.go
│   │   └── workflows.go
│   └── state/
│       ├── statefile.go
│       └── statefile_test.go
├── main.go
├── go.mod
└── go.sum
```

## Important Public Methods

### CLI Commands
1. `tgfs init`
   - Creates new workflow directory with example task
   - Args: None (interactive workflow name prompt)
   - Returns: Creates directory structure with sanitized workflow name

2. `tgfs plan [--dir <path>]`
   - Shows planned workflow changes
   - Args: 
     - `--dir, -d`: Directory containing workflows (default: ".")
   - Returns: Detailed plan output and creates `.tgfs-plan` file

3. `tgfs apply [--dir <path>] [--auto-approve]`
   - Executes planned workflow changes
   - Args:
     - `--dir, -d`: Directory containing workflows (default: ".")
     - `--auto-approve`: Skip interactive approval prompt
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
make test-short        # Run u
make test-verbose      # Run unit tests with verbose output
make test-watch        # Run unit tests and watch for changes
```

### Test Files
1. `internal/fsparse/parser_test.go`
   - Tests workflow parsing
   - Tests dependency resolution
   - Tests nested workflows

2. `internal/orchestration/orchestrator_test.go`
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

### Testing Utilities
1. `internal/testutils/helper.go`
   - Provides test running utilities with proper naming and logging
   - Captures caller information for better test output

2. `internal/integration/helpers.go`
   - Test workflow creation helpers
   - State verification utilities
   - Dependency link creation
   - Workflow name sanitization

## TODOs

1. Actually implement real task parsing on `tgfs plan` which utilizes a real gopilot client to parse the tasks and return a structure task object representing what was defined in the unstructured markdown files written by a human.
2. Actually implement and test task running in `tgfs apply` so that we can trigger the tasks defined in the structured task objects returned by gopilot inthe `tgfs plan` command.
3. Add workflow versioning and rollback support
4. Decide on if a metafile or the statefile should be written to the root of the workspace.

### Print Utilities
Located in `internal/printutils`:
- Task change detection and formatting
- Workflow state diff visualization
- Dependency formatting
- Plan output formatting in Terraform-like style