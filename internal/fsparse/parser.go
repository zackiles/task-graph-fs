package fsparse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zackiles/task-graph-fs/internal/gopilotcli"
)

type Parser struct {
	gopilot gopilotcli.GopilotCLI
}

// NewParser creates a new parser using the current gopilot provider
func NewParser() *Parser {
	return &Parser{
		gopilot: gopilotcli.GetProvider(),
	}
}

// NewParserWithGopilot creates a new parser with a specific gopilot implementation
// This is deprecated in favor of using the provider pattern
func NewParserWithGopilot(gopilot gopilotcli.GopilotCLI) *Parser {
	return &Parser{
		gopilot: gopilot,
	}
}

// ParseWorkflows walks through the given base path and constructs Workflow objects
func (p *Parser) ParseWorkflows(ctx context.Context, basePath string) ([]Workflow, error) {
	var workflows []Workflow

	// Walk through all directories recursively
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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
				workflow, err := p.parseWorkflow(ctx, path)
				if err != nil {
					return fmt.Errorf("failed to parse workflow %s: %w", path, err)
				}
				// Set the relative path as the workflow name
				workflow.Name = strings.TrimPrefix(path, basePath+string(os.PathSeparator))
				workflows = append(workflows, workflow)
			}

			return nil
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return workflows, nil
}

func (p *Parser) parseWorkflow(ctx context.Context, workflowPath string) (Workflow, error) {
	select {
	case <-ctx.Done():
		return Workflow{}, ctx.Err()
	default:
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

			taskPath := filepath.Join(workflowPath, entry.Name())

			// Validate task by parsing its properties
			command, deps, priority, retries, timeout, err := p.gopilot.GenerateTaskProps(ctx, taskPath)
			if err != nil {
				return Workflow{}, fmt.Errorf("failed to parse task %s: %w", entry.Name(), err)
			}

			task := Task{
				ID:           strings.TrimSuffix(entry.Name(), ".md"),
				MarkdownPath: taskPath,
				Command:      command,
				Dependencies: deps,
				Priority:     priority,
				Retries:      retries,
				Timeout:      timeout,
				Status:       "pending",
			}
			workflow.Tasks = append(workflow.Tasks, task)
		}

		return workflow, nil
	}
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
