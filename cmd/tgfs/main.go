package main

import (
	"os"

	"github.com/zackiles/task-graph-fs/internal/commands"
)

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
