package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/services"
)

// NewApplyCmd creates and returns the "apply" command.
func NewApplyCmd(parser *fsparse.Parser) *cobra.Command {
	var opts struct {
		autoApprove bool
		workflowDir string
	}

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the planned changes to workflows and tasks",
		Long: `The "apply" command executes the planned changes to the workflows and tasks.
It ensures that only approved changes are applied and supports interactive or
automatic approval modes.`,
		Args: cobra.NoArgs, // No positional arguments are expected
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runApply(ctx, parser, opts.workflowDir, opts.autoApprove)
		},
	}

	applyCmd.Flags().BoolVar(&opts.autoApprove, "auto-approve", false, "Skip interactive approval")
	applyCmd.Flags().StringVarP(&opts.workflowDir, "dir", "d", ".", "Directory containing workflows")

	return applyCmd
}

// runApply contains the core logic for the "apply" command.
func runApply(ctx context.Context, parser *fsparse.Parser, workflowDir string, autoApprove bool) error {
	applyService := services.NewApplyService(parser)

	// Check for changes first
	result, err := applyService.Plan(ctx, services.ApplyOptions{
		WorkflowDir: workflowDir,
		AutoApprove: autoApprove,
	})
	if err != nil {
		return fmt.Errorf("error during planning: %w", err)
	}

	if !result.HasChanges {
		fmt.Println("No changes to apply")
		return nil
	}

	// Confirm changes unless auto-approve is set
	if !autoApprove {
		if !confirmChanges() {
			fmt.Println("Apply cancelled")
			return nil
		}
	}

	// Set up cancellation context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	handleInterrupts(cancel)

	// Execute apply
	if err := applyService.Apply(ctx, services.ApplyOptions{
		WorkflowDir: workflowDir,
		AutoApprove: autoApprove,
	}); err != nil {
		return fmt.Errorf("error during apply: %w", err)
	}

	fmt.Println("\nApply complete!")
	return nil
}

// confirmChanges prompts the user for confirmation to apply changes.
func confirmChanges() bool {
	fmt.Printf("\nDo you want to apply these changes? [y/N] ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// handleInterrupts sets up a signal listener to handle graceful shutdown.
func handleInterrupts(cancelFunc context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, gracefully shutting down...")
		cancelFunc()
	}()
}
