package printutils

import (
	"fmt"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/state"
)

func PrintWorkflowAddition(w fsparse.Workflow) {
	fmt.Printf("  + \"workflows[%s]\" {\n", w.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", w.Name)
	fmt.Printf("      \"status\": \"pending\",\n")
	fmt.Printf("      \"tasks\": [\n")

	for i, task := range w.Tasks {
		fmt.Printf("        {\n")
		fmt.Printf("          \"id\": \"%s\",\n", task.ID)
		fmt.Printf("          \"command\": \"%s\",\n", task.Command)
		fmt.Printf("          \"dependencies\": [%s],\n", FormatDependencies(w.Dependencies[task.ID]))
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

func PrintWorkflowUpdate(new fsparse.Workflow, current *state.WorkflowState) {
	if current == nil {
		return
	}
	fmt.Printf("  ~ \"workflows[%s]\" {\n", new.Name)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", new.Name)

	fmt.Printf("      \"tasks\": [\n")
	PrintTaskChanges(new, current.Tasks)
	fmt.Printf("      ]\n")
	fmt.Printf("    }\n\n")
}

func PrintWorkflowRemoval(w *state.WorkflowState) {
	if w == nil {
		return
	}
	fmt.Printf("  - \"workflows[%s]\" {\n", w.WorkflowID)
	fmt.Printf("      \"workflow_id\": \"%s\",\n", w.WorkflowID)
	fmt.Printf("      \"status\": \"%s\",\n", w.Status)
	fmt.Printf("      \"tasks\": []\n")
	fmt.Printf("    }\n\n")
}
