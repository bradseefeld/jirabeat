package main

import (
	"os"

	"github.com/bradseefeld/jirabeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
