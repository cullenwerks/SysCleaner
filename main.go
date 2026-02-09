package main

import (
	"os"

	"syscleaner/cmd"
	"syscleaner/gui"
)

func main() {
	// Check if GUI flag is present
	if len(os.Args) > 1 && os.Args[1] == "gui" {
		gui.Run()
		return
	}

	// Otherwise run CLI
	cmd.Execute()
}
