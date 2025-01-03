package gopilotcli

import "sync"

type GopilotCLI interface {
	GenerateTaskProps(taskPath string) (command string, dependencies []string, priority string, retries int, timeout string, err error)
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

func (m *MockGopilot) SetResponse(taskPath string, response TaskResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[taskPath] = response
}

func (m *MockGopilot) GenerateTaskProps(taskPath string) (string, []string, string, int, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if response, ok := m.responses[taskPath]; ok {
		return response.Command, response.Dependencies, response.Priority, response.Retries, response.Timeout, response.Error
	}

	return "", nil, "medium", 1, "30m", nil
}
