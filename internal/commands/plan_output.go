package commands

import (
	"fmt"
	"strings"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/state"
)

func printWorkflowAddition(w fsparse.Workflow) {
	fmt.Printf("  + \"workflows[%s]\" {\n", w.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", w.Name)
	fmt.Printf("      \"status\": \"pending\",\n")
	fmt.Printf("      \"tasks\": [\n")

	for i, task := range w.Tasks {
		fmt.Printf("        {\n")
		fmt.Printf("          \"id\": \"%s\",\n", task.ID)
		fmt.Printf("          \"command\": \"%s\",\n", task.Command)
		fmt.Printf("          \"dependencies\": [%s],\n", formatDependencies(w.Dependencies[task.ID]))
		fmt.Printf("          \"priority\": \"%s\",\n", task.Priority)
		fmt.Printf("          \"retries\": %d,\n", task.Retries)
		fmt.Printf("          \"timeout\": \"%s\",\n", task.Timeout)
		fmt.Printf("          \"status\": \"pending\"\n")
		fmt.Printf("        }")
		if i < len(w.Tasks)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("      ]\n")
	fmt.Printf("    }\n\n")
}

func printWorkflowUpdate(new fsparse.Workflow, current *state.WorkflowState) {
	if current == nil {
		return
	}
	fmt.Printf("  ~ \"workflows[%s]\" {\n", new.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", new.Name)

	// Print task changes
	fmt.Printf("      \"tasks\": [\n")
	printTaskChanges(new, current.Tasks)
	fmt.Printf("      ]\n")
	fmt.Printf("    }\n\n")
}

func printWorkflowRemoval(w *state.WorkflowState) {
	if w == nil {
		return
	}
	fmt.Printf("  - \"workflows[%s]\" {\n", w.WorkflowID)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", w.WorkflowID)
	fmt.Printf("      \"status\": \"%s\",\n", w.Status)
	fmt.Printf("      \"tasks\": []\n")
	fmt.Printf("    }\n\n")
}

func formatDependencies(deps []string) string {
	if len(deps) == 0 {
		return ""
	}
	quoted := make([]string, len(deps))
	for i, dep := range deps {
		quoted[i] = fmt.Sprintf("\"%s\"", dep)
	}
	return strings.Join(quoted, ", ")
}

func printTaskChanges(workflow fsparse.Workflow, currentTasks []state.TaskState) {
	// Create maps for quick lookups
	currentTaskMap := make(map[string]state.TaskState)
	for _, t := range currentTasks {
		currentTaskMap[t.ID] = t
	}

	printed := false
	for _, task := range workflow.Tasks {
		if current, exists := currentTaskMap[task.ID]; exists {
			if taskHasChanges(task, current) {
				if printed {
					fmt.Printf(",\n")
				}
				printTaskUpdate(task, current, workflow.Dependencies[task.ID])
				printed = true
			}
		} else {
			if printed {
				fmt.Printf(",\n")
			}
			printTaskAddition(task, workflow.Dependencies[task.ID])
			printed = true
		}
	}
	fmt.Printf("\n")
}

func printTaskUpdate(new fsparse.Task, current state.TaskState, deps []string) {
	fmt.Printf("        ~ \"tasks[%s]\" {\n", new.ID)
	fmt.Printf("            \"id\": \"%s\",\n", new.ID)
	if new.Command != current.Command {
		fmt.Printf("            \"command\": \"%s\" -> \"%s\",\n", current.Command, new.Command)
	}
	if !stringSlicesEqual(deps, current.Dependencies) {
		fmt.Printf("            \"dependencies\": [%s] -> [%s],\n",
			formatDependencies(current.Dependencies),
			formatDependencies(deps))
	}
	if new.Priority != current.Priority {
		fmt.Printf("            \"priority\": \"%s\" -> \"%s\",\n", current.Priority, new.Priority)
	}
	if new.Retries != current.Retries {
		fmt.Printf("            \"retries\": %d -> %d,\n", current.Retries, new.Retries)
	}
	if new.Timeout != "" {
		fmt.Printf("            \"timeout\": \"%s\",\n", new.Timeout)
	}
	fmt.Printf("            \"status\": \"%s\"\n", current.Status)
	fmt.Printf("          }")
}

func printTaskAddition(task fsparse.Task, deps []string) {
	fmt.Printf("        + \"tasks[%s]\" {\n", task.ID)
	fmt.Printf("            \"id\": \"%s\",\n", task.ID)
	fmt.Printf("            \"command\": \"%s\",\n", task.Command)
	fmt.Printf("            \"dependencies\": [%s],\n", formatDependencies(deps))
	fmt.Printf("            \"priority\": \"%s\",\n", task.Priority)
	fmt.Printf("            \"retries\": %d,\n", task.Retries)
	fmt.Printf("            \"timeout\": \"%s\",\n", task.Timeout)
	fmt.Printf("            \"status\": \"pending\"\n")
	fmt.Printf("          }")
}

func taskHasChanges(newTask fsparse.Task, currentTask state.TaskState) bool {
	return newTask.Command != currentTask.Command ||
		!stringSlicesEqual(newTask.Dependencies, currentTask.Dependencies) ||
		newTask.Priority != currentTask.Priority ||
		newTask.Retries != currentTask.Retries
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
