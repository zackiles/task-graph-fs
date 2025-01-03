package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/company/task-graph-fs/internal/fsparse"
)

const StateFileName = "tgfs-state.json"

type StateFile struct {
	Workflows []WorkflowState `json:"workflows"`
}

type WorkflowState struct {
	WorkflowID string      `json:"workflow_id"`
	Status     string      `json:"status"`
	Tasks      []TaskState `json:"tasks"`
}

type TaskState struct {
	ID           string   `json:"id"`
	Command      string   `json:"command"`
	Dependencies []string `json:"dependencies"`
	Priority     string   `json:"priority"`
	Retries      int      `json:"retries"`
	Status       string   `json:"status"`
	Duration     string   `json:"duration,omitempty"`
	Output       string   `json:"output,omitempty"`
}

func LoadState() (*StateFile, error) {
	data, err := os.ReadFile(StateFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &StateFile{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state StateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

func (s *StateFile) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(StateFileName, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (s *StateFile) ComputeDiff(workflows []fsparse.Workflow) (added, updated, removed []string) {
	// Create maps for quick lookup
	currentWorkflows := make(map[string]WorkflowState)
	for _, w := range s.Workflows {
		currentWorkflows[w.WorkflowID] = w
	}

	newWorkflows := make(map[string]fsparse.Workflow)
	for _, w := range workflows {
		newWorkflows[w.Name] = w
	}

	// Find added and updated workflows
	for name, workflow := range newWorkflows {
		if _, exists := currentWorkflows[name]; !exists {
			added = append(added, name)
		} else {
			// Compare tasks to determine if workflow was updated
			if hasChanges(currentWorkflows[name], workflow) {
				updated = append(updated, name)
			}
		}
	}

	// Find removed workflows
	for name := range currentWorkflows {
		if _, exists := newWorkflows[name]; !exists {
			removed = append(removed, name)
		}
	}

	return
}

func hasChanges(current WorkflowState, new fsparse.Workflow) bool {
	// Compare task counts
	if len(current.Tasks) != len(new.Tasks) {
		return true
	}

	// Create map of current tasks
	currentTasks := make(map[string]TaskState)
	for _, t := range current.Tasks {
		currentTasks[t.ID] = t
	}

	// Compare each task
	for _, newTask := range new.Tasks {
		if currentTask, exists := currentTasks[newTask.ID]; !exists {
			return true
		} else if currentTask.Command != newTask.Command ||
			len(currentTask.Dependencies) != len(newTask.Dependencies) {
			return true
		}
	}

	return false
}
