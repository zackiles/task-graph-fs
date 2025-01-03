# TaskGraphFS

**The incredibly simple and astonishingly powerful task engine for humans.** Build and run complex workflows using only files, folders, symlinks, and natural language.

## Overview

TaskGraphFS lets you express complex graph-based workflows using nothing more than the filesystem. Define workflows as folders, tasks as files, and dependencies as symbolic links. Use plain English in markdown files to describe tasks, and TaskGraphFS's built-in LLM capabilities will handle the rest.

- No new languages to learn.
- No rigid schema formats.
- No complex user interfaces to master.

If you can use a file explorer, you already know how to use TaskGraphFS.

### Graph to Filesystem Mapping

TaskGraphFS maps directed acyclic graphs (DAGs) to your filesystem:
- **Workflows (Subgraphs)** → Directories
- **Tasks (Nodes)** → Markdown Files
- **Dependencies (Edges)** → Symbolic Links
- **Task Properties** → Markdown Content
- **Nested Workflows** → Subdirectories

## Why TaskGraphFS?

The rise of AI has brought a flood of new frameworks with complex abstractions for defining workflows and managing state. Most rely on proprietary formats or rigid schemas, creating unnecessary barriers. TaskGraphFS takes a simpler approach: it uses the filesystem as the interface and plain English for task definitions, with LLMs handling the complexity. It's intuitive, interoperable, and built to align with AI's promise of natural, human-first workflows.

## Getting Started

1. **Install TaskGraphFS:**
   ```bash
   go install github.com/company/task-graph-fs/cmd/tgfs@latest
   ```

2. **Initialize a Workflow:**
   ```bash
   tgfs init --workflow-name my-workflow --task-name first-task
   ```

3. **Run Your Workflow:**
   ```bash
   tgfs plan    # Preview changes
   tgfs apply   # Apply and execute workflow
   ```

## CLI Commands

### Initialize a New Workflow
```bash
tgfs init --workflow-name <workflow-name> --task-name <task-name>
```

### 1. Plan()

Preview the changes that will be made to your workflow.

```bash
tgfs plan
```

**Output Example:**

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

  ~ "workflows[1]" {
      "workflow_id": "ExistingWorkflow",
      "tasks": [
        ~ "tasks[0]" {
            "id": "TaskA",
            "command": "python updated_task_a.py",
            "dependencies": ["TaskB"],
            "priority": "high",
            "retries": 3,
            "status": "pending" -> "running"
          }
        + "tasks[2]" {
            "id": "TaskC",
            "command": "python task_c.py",
            "dependencies": [],
            "priority": "low",
            "retries": 0,
            "status": "pending"
          }
      ]
    }

  - "workflows[2]" {
      "workflow_id": "DeprecatedWorkflow",
      "status": "completed",
      "tasks": []
    }

Plan Summary:
- Workflows: 1 to add, 1 to update, 1 to remove.
- Tasks: 1 to update, 1 to add, 1 to remove.

Run `tgfs apply` to execute these changes.
```

### 2. Apply()

Execute your workflow and generate a statefile in the working directory.

```bash
tgfs apply
```

**Output Example:**

```
Executing Workflow: MyWorkflow...

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
    },
    {
      "id": "TaskA",
      "command": "python task_a.py",
      "dependencies": ["TaskB"],
      "priority": "high",
      "retries": 3,
      "status": "completed",
      "duration": "5s",
      "output": "Processed data successfully."
    }
  ]
}

Workflow Completed Successfully!
```

## Task Definition

Tasks are defined in markdown files. While you can write your tasks in natural language, here's an example of how to structure your markdown to make it easier for the LLM to extract the required properties:

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

These properties (Command, Dependencies, Priority, Retries, Timeout) are the schema that TaskGraphFS uses internally, but you're free to describe your tasks in natural language - our LLM will extract these properties from your description.

## Example Workflow Structure

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

This structure shows:
- Multiple top-level workflows (`data-pipeline` and `model-training`)
- Nested workflows (under `data-pipeline/nested`)
- Cross-workflow dependencies (model training depending on data pipeline)
- Proper relative symlinks for dependency edges

## State Management

TaskGraphFS maintains a state file (`tgfs-state.json`) that tracks:
- Workflow status
- Task completion state
- Execution history
- Error information
- Task outputs

## Error Handling

- Automatic retries with configurable attempts
- Task-level timeout enforcement
- Graceful workflow cancellation
- State recovery after interruption

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.