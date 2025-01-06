package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/vendor/gopilotcli"
)

var rootCmd = &cobra.Command{
	Use:   "tgfs",
	Short: "Filesystem-based task orchestration",
	Long: `TaskGraphFS (tgfs) is a tool for defining and executing task workflows
using a filesystem-based approach with markdown files and symbolic links.`,
}

func NewRootCommand() *cobra.Command {
	gopilot := gopilotcli.NewRealGopilot()
	parser := fsparse.NewParser(gopilot)

	rootCmd.AddCommand(
		NewInitCmd(),
		NewPlanCmd(parser)
		NewApplyCmd(parser),
	)

	return rootCmd
}
