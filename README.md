# TaskGraphFS

**The incredibly simple and astonishingly powerful execution-agnostic workflow definition engine.** Build and run complex workflows using only files, folders, symlinks, and natural language.

## Overview

TaskGraphFS is an execution-agnostic workflow definition engine. Express complex graph-based workflows, along with their tasks, inputs, outputs, and depedencies, using only files, folders, symlinks, and natural language. Plan and apply your workflows as you would Terraform, passing it's configuration and state to an execution engine of your choice. No new DSLs, SDKs, schemas, formats, or user interfaces to learn. If you can use a file explorer, you already know how to use TaskGraphFS.

### File Systems as a Graph

If we consider a workflow as a graph, we can map the graph to the filesystem as follows:

- **Workflows (Subgraphs)** → Directories.
- **Tasks (Nodes)** → Files
- **Dependencies (Edges)** → Symbolic Links
- **Task Properties** → Natural language in the files
- **Nested Workflows** → Subdirectories

For an example of a workflow expressed in this way, see the [example workflow structure](#example-workflow-structure).

### Why TaskGraphFS?

Because running mkdir, touch, and then piping a few sentences into a file to create workflows is an unmatched user experience. The rise of AI is spwaning 1000s of new task and agentic workflow frameworks each week, but beyond a few attempting to standardize around projects like Langraph and BAML, most bring completely new concepts to the table. As sand castles, they have little interopbility at a time we need the most, and as such they're forced to bring the entire kitchen sync to compensate for that lack of integration. Given the promise that LLMs bring through the power of natural language, it's also surprising every abstraction for workflows so far has been anything but natural language. TaskGraphFS tries to do one thing and do it well: provide a simple, natural language interface for planning and managing complex workflows.

## Getting Started

1. **Install TaskGraphFS:**
   ```bash
   go install github.com/company/task-graph-fs/cmd/tgfs@latest
   ```

2. **Initialize a Workflow:**
   ```bash
   tgfs init
   ```
   You'll be prompted to enter a workflow name. The name will be automatically formatted to be lowercase, hyphen-separated, and alphanumeric.

3. **Run Your Workflow:**
   ```bash
   tgfs plan    # Preview changes
   tgfs apply   # Apply and execute workflow
   ```

## CLI Commands

### Initialize a New Workflow
```bash
tgfs init
```
Interactively prompts for a workflow name. The name will be automatically sanitized to be lowercase, hyphen-separated, and alphanumeric.

### Plan Changes
Preview the changes that will be made to your workflow.

```bash
tgfs plan [--dir <directory>]
```
By default, `plan` runs in the current directory if `--dir` is not specified.

### Apply Changes
Apply and execute the planned changes.

```bash
tgfs apply [--auto-approve]
```
Without `--auto-approve`, you'll be prompted to confirm the changes before execution.

### Command Output Examples

The plan and apply output examples in the README are accurate to the actual implementation in the code, but I would add a note about the interactive confirmation for apply:

```bash
$ tgfs apply

# Without --auto-approve, you'll see:
Do you want to apply these changes? [y/N]
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