package gopilotcli

import (
	"fmt"
	"os/exec"
)

// NOTE: This is not a real Gopilot integration
// Gopilot is still being developed and can be found here https://github.com/zackiles/gopilot
// For now, we're just mocking the expected interface and behavior
type RealGopilot struct{}

func NewRealGopilot() *RealGopilot {
	return &RealGopilot{}
}

func (g *RealGopilot) GenerateTaskProps(taskPath string) (string, []string, string, int, string, error) {
	cmd := exec.Command("gopilot", "create a task for this file", "--path", taskPath)
	_, err := cmd.Output()
	if err != nil {
		return "", nil, "", 0, "", fmt.Errorf("gopilot CLI error: %w", err)
	}

	// TODO: Parse gopilot output format
	// For now, return defaults
	return "echo test", []string{}, "medium", 1, "30m", nil
}
