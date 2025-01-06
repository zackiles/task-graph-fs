package gopilotcli

import "sync"

// NOTE: This is not a real Gopilot integration
// Gopilot is still being developed and can be found here https://github.com/zackiles/gopilot
// For now, we're just mocking the expected interface and behavior
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

func (m *MockGopilot) SetResponse(path string, response TaskResponse) {
	m.responses[path] = response
}

func (m *MockGopilot) GenerateTaskProps(taskPath string) (string, []string, string, int, string, error) {
	if response, ok := m.responses[taskPath]; ok {
		if response.Error != nil {
			return "", nil, "", 0, "", response.Error
		}
		return response.Command, response.Dependencies, response.Priority, response.Retries, response.Timeout, nil
	}
	return "", nil, "", 0, "", nil
}
