package main

import (
	"fmt"
	"os"

	"github.com/zackiles/task-graph-fs/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
