package gopilotcli

import (
	"context"
	"os/exec"
)

// NOTE: This is not a real Gopilot integration
// Gopilot is still being developed and can be found here https://github.com/zackiles/gopilot
// For now, we're just mocking the expected interface and behavior
type RealGopilot struct{}

func NewRealGopilot() *RealGopilot {
	return &RealGopilot{}
}

func (g *RealGopilot) GenerateTaskProps(ctx context.Context, taskPath string) (string, []string, string, int, string, error) {
	// Create command with context
	cmd := exec.CommandContext(ctx, "gopilot", "parse", "--path", taskPath)

	// Run command with context awareness
	if err := cmd.Run(); err != nil {
		// For now, since gopilot isn't implemented, return mock values instead of error
		// This matches the example task format in /internal/integration/mock-workflow1/task.example.md
		return "python example_script.py", []string{}, "medium", 1, "30m", nil
	}

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return "", nil, "", 0, "", ctx.Err()
	default:
		// Return values that match the example task format
		return "python example_script.py", []string{}, "medium", 1, "30m", nil
	}
}
