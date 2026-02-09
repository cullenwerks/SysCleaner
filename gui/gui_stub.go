//go:build !gui

package gui

import "fmt"

// Run prints an error when the GUI build tag is not enabled.
func Run() {
	fmt.Println("GUI mode is not available in this build.")
	fmt.Println("Rebuild with: go build -tags gui")
}
