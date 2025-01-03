package main

import (
	"fmt"
	"strings"

	"github.com/company/task-graph-fs/internal/fsparse"
	"github.com/company/task-graph-fs/internal/state"
)

func printWorkflowAddition(w fsparse.Workflow) {
	fmt.Printf("  + \"workflows[%s]\" {\n", w.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", w.Name)
	fmt.Printf("      \"status\": \"pending\",\n")
	fmt.Printf("      \"tasks\": [\n")

	for _, task := range w.Tasks {
		fmt.Printf("        {\n")
		fmt.Printf("          \"id\": \"%s\",\n", task.ID)
		fmt.Printf("          \"command\": \"%s\",\n", task.Command)
		fmt.Printf("          \"dependencies\": [%s],\n", formatDependencies(task.Dependencies))
		fmt.Printf("          \"priority\": \"%s\",\n", task.Priority)
		fmt.Printf("          \"retries\": %d,\n", task.Retries)
		fmt.Printf("          \"status\": \"pending\"\n")
		fmt.Printf("        }")
		if task != w.Tasks[len(w.Tasks)-1] {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("      ]\n")
	fmt.Printf("    }\n\n")
}

func printWorkflowUpdate(new fsparse.Workflow, current state.WorkflowState) {
	fmt.Printf("  ~ \"workflows[%s]\" {\n", new.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", new.Name)

	// Print task changes
	fmt.Printf("      \"tasks\": [\n")
	printTaskChanges(new.Tasks, current.Tasks)
	fmt.Printf("      ]\n")
	fmt.Printf("    }\n\n")
}

func printWorkflowRemoval(w state.WorkflowState) {
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

func printTaskChanges(newTasks []fsparse.Task, currentTasks []state.TaskState) {
	// Create maps for quick lookups
	currentTaskMap := make(map[string]state.TaskState)
	for _, t := range currentTasks {
		currentTaskMap[t.ID] = t
	}

	for i, task := range newTasks {
		if current, exists := currentTaskMap[task.ID]; exists {
			if taskHasChanges(task, current) {
				printTaskUpdate(task, current)
			}
		} else {
			printTaskAddition(task)
		}

		if i < len(newTasks)-1 {
			fmt.Printf(",\n")
		}
	}
}

func printTaskUpdate(new fsparse.Task, current state.TaskState) {
	fmt.Printf("        ~ \"tasks[%s]\" {\n", new.ID)
	fmt.Printf("            \"id\": \"%s\",\n", new.ID)
	if new.Command != current.Command {
		fmt.Printf("            \"command\": \"%s\" -> \"%s\",\n", current.Command, new.Command)
	}
	if !stringSlicesEqual(new.Dependencies, current.Dependencies) {
		fmt.Printf("            \"dependencies\": [%s] -> [%s],\n",
			formatDependencies(current.Dependencies),
			formatDependencies(new.Dependencies))
	}
	fmt.Printf("            \"status\": \"%s\"\n", current.Status)
	fmt.Printf("          }")
}

func printTaskAddition(task fsparse.Task) {
	fmt.Printf("        + \"tasks[%s]\" {\n", task.ID)
	fmt.Printf("            \"id\": \"%s\",\n", task.ID)
	fmt.Printf("            \"command\": \"%s\",\n", task.Command)
	fmt.Printf("            \"dependencies\": [%s],\n", formatDependencies(task.Dependencies))
	fmt.Printf("            \"priority\": \"%s\",\n", task.Priority)
	fmt.Printf("            \"retries\": %d,\n", task.Retries)
	fmt.Printf("            \"status\": \"pending\"\n")
	fmt.Printf("          }")
}
