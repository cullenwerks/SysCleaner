//go:build gui

package main

import (
	"syscleaner/gui"
)

func main() {
	// Always launch GUI - this is a GUI-only application
	gui.Run()
}
