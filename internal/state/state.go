package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zackiles/task-graph-fs/internal/fsparse"
)

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
	Output       string   `json:"output,omitempty"`
}

// LoadState loads the state from the state file
func LoadState() (*StateFile, error) {
	data, err := os.ReadFile("tgfs-state.json")
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

// Save writes the state to the state file
func (s *StateFile) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile("tgfs-state.json", data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// ComputeDiff compares the current state with the new workflows and returns
// lists of workflows that need to be added, updated, or removed
func (s *StateFile) ComputeDiff(workflows []fsparse.Workflow) (added, updated, removed []string) {
	// Create maps for quick lookups
	currentWorkflows := make(map[string]bool)
	newWorkflows := make(map[string]bool)

	// Track current workflows
	for _, w := range s.Workflows {
		currentWorkflows[w.WorkflowID] = true
	}

	// Track new workflows and find additions/updates
	for _, w := range workflows {
		newWorkflows[w.Name] = true
		if !currentWorkflows[w.Name] {
			added = append(added, w.Name)
		} else {
			updated = append(updated, w.Name)
		}
	}

	// Find removals
	for _, w := range s.Workflows {
		if !newWorkflows[w.WorkflowID] {
			removed = append(removed, w.WorkflowID)
		}
	}

	return added, updated, removed
}
