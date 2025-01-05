package commands

import (
	"github.com/company/task-graph-fs/internal/fsparse"
	"github.com/company/task-graph-fs/internal/gopilotcli"
	"github.com/spf13/cobra"
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
