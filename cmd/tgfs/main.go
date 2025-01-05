package main

import (
	"os"

	"github.com/company/task-graph-fs/internal/commands"
)

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
