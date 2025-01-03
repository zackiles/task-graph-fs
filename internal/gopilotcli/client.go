package gopilotcli

import (
	"fmt"
	"os/exec"
)

type RealGopilot struct{}

func NewRealGopilot() *RealGopilot {
	return &RealGopilot{}
}

func (g *RealGopilot) GenerateTaskProps(taskPath string) (string, []string, string, int, string, error) {
	cmd := exec.Command("gopilot", "create a task for this file", "--path", taskPath)
	output, err := cmd.Output()
	if err != nil {
		return "", nil, "", 0, "", fmt.Errorf("gopilot CLI error: %w", err)
	}

	// TODO: Parse gopilot output format
	// For now, return defaults
	return "echo test", []string{}, "medium", 1, "30m", nil
}
