package printutils

import (
	"fmt"
	"strings"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/state"
)

func PrintTaskChanges(workflow fsparse.Workflow, currentTasks []state.TaskState) {
	currentTaskMap := make(map[string]state.TaskState)
	for _, t := range currentTasks {
		currentTaskMap[t.ID] = t
	}

	printed := false
	for _, task := range workflow.Tasks {
		if current, exists := currentTaskMap[task.ID]; exists {
			if TaskHasChanges(task, current) {
				if printed {
					fmt.Printf(",\n")
				}
				PrintTaskUpdate(task, current, workflow.Dependencies[task.ID])
				printed = true
			}
		} else {
			if printed {
				fmt.Printf(",\n")
			}
			PrintTaskAddition(task, workflow.Dependencies[task.ID])
			printed = true
		}
	}
	fmt.Printf("\n")
}

func PrintTaskUpdate(new fsparse.Task, current state.TaskState, deps []string) {
	fmt.Printf("        ~ \"tasks[%s]\" {\n", new.ID)
	fmt.Printf("            \"id\": \"%s\",\n", new.ID)
	if new.Command != current.Command {
		fmt.Printf("            \"command\": \"%s\" -> \"%s\",\n", current.Command, new.Command)
	}
	if !StringSlicesEqual(deps, current.Dependencies) {
		fmt.Printf("            \"dependencies\": [%s] -> [%s],\n",
			FormatDependencies(current.Dependencies),
			FormatDependencies(deps))
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

func PrintTaskAddition(task fsparse.Task, deps []string) {
	fmt.Printf("        + \"tasks[%s]\" {\n", task.ID)
	fmt.Printf("            \"id\": \"%s\",\n", task.ID)
	fmt.Printf("            \"command\": \"%s\",\n", task.Command)
	fmt.Printf("            \"dependencies\": [%s],\n", FormatDependencies(deps))
	fmt.Printf("            \"priority\": \"%s\",\n", task.Priority)
	fmt.Printf("            \"retries\": %d,\n", task.Retries)
	fmt.Printf("            \"timeout\": \"%s\",\n", task.Timeout)
	fmt.Printf("            \"status\": \"pending\"\n")
	fmt.Printf("          }")
}

func FormatDependencies(deps []string) string {
	if len(deps) == 0 {
		return ""
	}
	quoted := make([]string, len(deps))
	for i, dep := range deps {
		quoted[i] = fmt.Sprintf("\"%s\"", dep)
	}
	return strings.Join(quoted, ", ")
}

func TaskHasChanges(newTask fsparse.Task, currentTask state.TaskState) bool {
	return newTask.Command != currentTask.Command ||
		!StringSlicesEqual(newTask.Dependencies, currentTask.Dependencies) ||
		newTask.Priority != currentTask.Priority ||
		newTask.Retries != currentTask.Retries
}

func StringSlicesEqual(a, b []string) bool {
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
