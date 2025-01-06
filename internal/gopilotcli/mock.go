package gopilotcli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// NOTE: This is not a real Gopilot integration
// Gopilot is still being developed and can be found here https://github.com/zackiles/gopilot
// For now, we're just mocking the expected interface and behavior
type GopilotCLI interface {
	GenerateTaskProps(ctx context.Context, taskPath string) (command string, dependencies []string, priority string, retries int, timeout string, err error)
}

type MockGopilot struct {
	mu        sync.Mutex
	responses map[string]TaskResponse
}

type TaskResponse struct {
	Command      string
	Dependencies []string
	Priority     string
	Retries      int
	Timeout      string
	Error        error
}

func NewMockGopilot() *MockGopilot {
	return &MockGopilot{
		responses: make(map[string]TaskResponse),
	}
}

func normalizePath(path string) string {
	// Clean the path first to remove any redundant separators
	path = filepath.Clean(path)

	// Remove /private prefix on macOS
	path = strings.TrimPrefix(path, "/private")

	// Evaluate symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = realPath
	}

	return path
}

func (m *MockGopilot) SetResponse(path string, response TaskResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Always store using normalized absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Sprintf("failed to get absolute path for mock response: %v", err))
	}
	normalizedPath := normalizePath(absPath)
	fmt.Printf("Setting response for normalized path: %s\n", normalizedPath)
	m.responses[normalizedPath] = response
}

func (m *MockGopilot) GenerateTaskProps(ctx context.Context, taskPath string) (string, []string, string, int, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Always use normalized absolute path for lookup
	absPath, err := filepath.Abs(taskPath)
	if err != nil {
		return "", nil, "", 0, "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	normalizedPath := normalizePath(absPath)
	fmt.Printf("Looking up response for normalized path: %s\n", normalizedPath)

	// Check for configured response
	if response, ok := m.responses[normalizedPath]; ok {
		if response.Error != nil {
			fmt.Printf("Mock returning error for path %s: %v\n", normalizedPath, response.Error)
			return "", nil, "", 0, "", response.Error
		}
		return response.Command, response.Dependencies, response.Priority, response.Retries, response.Timeout, nil
	}

	// Return default values if no response is configured
	fmt.Printf("Mock using default response for path %s\n", normalizedPath)
	return "echo default", []string{}, "medium", 1, "30m", nil
}

func (m *MockGopilot) GetResponses() map[string]TaskResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy to avoid map races
	responses := make(map[string]TaskResponse)
	for k, v := range m.responses {
		responses[k] = v
	}
	return responses
}
