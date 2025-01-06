package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/services"
)

// NewPlanCmd creates and returns the "plan" command.
func NewPlanCmd(parser *fsparse.Parser) *cobra.Command {
	if parser == nil {
		// Return a command that will fail when executed
		return &cobra.Command{
			Use:   "plan",
			Short: "Plan workflow execution",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("parser is required")
			},
		}
	}

	var opts struct {
		workflowDir string
	}

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan workflow execution",
		Long:  `The "plan" command analyzes workflows and creates an execution plan.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runPlan(ctx, parser, opts.workflowDir)
		},
	}

	planCmd.Flags().StringVarP(&opts.workflowDir, "dir", "d", ".", "Directory containing workflows")
	return planCmd
}

// runPlan contains the core logic for the "plan" command.
func runPlan(ctx context.Context, parser *fsparse.Parser, workflowDir string) error {
	if parser == nil {
		return fmt.Errorf("parser is required")
	}

	applyService := services.NewApplyService(parser)
	if applyService == nil {
		return fmt.Errorf("failed to create apply service")
	}

	result, err := applyService.Plan(ctx, services.ApplyOptions{
		WorkflowDir: workflowDir,
	})
	if err != nil {
		// Don't create plan file if there's an error
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Only create plan file if planning was successful
	planData := map[string]interface{}{
		"added":      result.Added,
		"updated":    result.Updated,
		"removed":    result.Removed,
		"hasChanges": result.HasChanges,
	}

	planJSON, err := json.MarshalIndent(planData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan data: %w", err)
	}

	if err := os.WriteFile(".tgfs-plan", planJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	// Print plan summary
	fmt.Println("\nPlan Summary:")
	fmt.Printf("  Added:   %d\n", len(result.Added))
	fmt.Printf("  Updated: %d\n", len(result.Updated))
	fmt.Printf("  Removed: %d\n", len(result.Removed))

	if !result.HasChanges {
		fmt.Println("\nNo changes to apply")
	}

	return nil
}
