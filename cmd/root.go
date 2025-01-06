package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
)

func NewRootCommand() *cobra.Command {
	parser := fsparse.NewParser()

	rootCmd := &cobra.Command{
		Use:   "tgfs",
		Short: "Filesystem-based task orchestration",
		Long: `TaskGraphFS (tgfs) is a tool for defining and executing task workflows
using a filesystem-based approach with markdown files and symbolic links.`,
	}

	rootCmd.AddCommand(
		NewInitCmd(),
		NewPlanCmd(parser),
		NewApplyCmd(parser),
	)

	return rootCmd
}
