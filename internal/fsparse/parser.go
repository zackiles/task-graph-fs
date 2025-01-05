package fsparse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/company/task-graph-fs/internal/gopilotcli"
)

type Parser struct {
	gopilot gopilotcli.GopilotCLI
}

func NewParser(gopilot gopilotcli.GopilotCLI) *Parser {
	return &Parser{gopilot: gopilot}
}

// ParseWorkflows walks through the given base path and constructs Workflow objects
func (p *Parser) ParseWorkflows(basePath string) ([]Workflow, error) {
	var workflows []Workflow

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workflow, err := p.parseWorkflow(filepath.Join(basePath, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse workflow %s: %w", entry.Name(), err)
		}
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func (p *Parser) parseWorkflow(workflowPath string) (Workflow, error) {
	workflow := Workflow{
		Name:         filepath.Base(workflowPath),
		Dependencies: make(map[string][]string),
	}

	entries, err := os.ReadDir(workflowPath)
	if err != nil {
		return Workflow{}, fmt.Errorf("failed to read workflow directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			// Check if it's a symlink representing dependencies
			if entry.Type()&os.ModeSymlink != 0 {
				sourceTask := strings.TrimSuffix(entry.Name(), "_dependencies")
				target, err := os.Readlink(filepath.Join(workflowPath, entry.Name()))
				if err != nil {
					return Workflow{}, fmt.Errorf("failed to read symlink: %w", err)
				}
				targetTask := strings.TrimSuffix(filepath.Base(target), ".md")
				workflow.Dependencies[sourceTask] = append(
					workflow.Dependencies[sourceTask],
					targetTask,
				)
			}
			continue
		}

		task := Task{
			ID:           strings.TrimSuffix(entry.Name(), ".md"),
			MarkdownPath: filepath.Join(workflowPath, entry.Name()),
			Status:       "pending",
		}
		workflow.Tasks = append(workflow.Tasks, task)
	}

	return workflow, nil
}

func ParseWorkflows(basePath string) ([]Workflow, error) {
	// Create a default parser with a real gopilot client
	parser := NewParser(gopilotcli.NewRealGopilot())
	return parser.ParseWorkflows(basePath)
}
