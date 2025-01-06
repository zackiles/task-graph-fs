package fsparse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zackiles/task-graph-fs/internal/gopilotcli"
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

	// Walk through all directories recursively
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory or if it's the base path
		if !info.IsDir() || path == basePath {
			return nil
		}

		// Skip hidden directories and non-workflow directories
		if strings.HasPrefix(info.Name(), ".") || isProjectDirectory(info.Name()) {
			return filepath.SkipDir
		}

		// Check if directory contains any .md files
		if containsMarkdownFiles(path) {
			workflow, err := p.parseWorkflow(path)
			if err != nil {
				return fmt.Errorf("failed to parse workflow %s: %w", path, err)
			}
			// Set the relative path as the workflow name
			workflow.Name = strings.TrimPrefix(path, basePath+string(os.PathSeparator))
			workflows = append(workflows, workflow)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
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

// Helper function to identify project-specific directories
func isProjectDirectory(name string) bool {
	projectDirs := map[string]bool{
		"cmd":      true,
		"internal": true,
		"bin":      true,
		"dist":     true,
		"vendor":   true,
	}
	return projectDirs[name]
}

// Helper function to check if a directory contains markdown files
func containsMarkdownFiles(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			return true
		}
	}
	return false
}

func ParseWorkflows(basePath string) ([]Workflow, error) {
	// Create a default parser with a real gopilot client
	parser := NewParser(gopilotcli.NewRealGopilot())
	return parser.ParseWorkflows(basePath)
}
