package commands

import (
	"github.com/spf13/cobra"
	"github.com/zackiles/task-graph-fs/internal/fsparse"
	"github.com/zackiles/task-graph-fs/internal/gopilotcli"
)

func NewRootCommand() *cobra.Command {
	gopilot := gopilotcli.NewRealGopilot()
	parser := fsparse.NewParser(gopilot)

	rootCmd := &cobra.Command{
		Use:   "tgfs",
		Short: "Filesystem-based task orchestration",
		Long: `TaskGraphFS (tgfs) is a tool for defining and executing task workflows
using a filesystem-based approach with markdown files and symbolic links.`,
	}

	rootCmd.AddCommand(
		NewInitCommand(),
		newPlanCmd(parser),
		newApplyCmd(parser),
	)

	return rootCmd
}
