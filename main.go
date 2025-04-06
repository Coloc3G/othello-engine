package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Coloc3G/othello-engine/ui"
)

func main() {
	// Define minimal command line flags
	helpPtr := flag.Bool("help", false, "Show help information")
	flag.Parse()

	// Show help information if requested
	if *helpPtr {
		fmt.Println("Othello Game")
		fmt.Println("Launch the application with no arguments to start the UI")
		fmt.Println("Additional functionality is available in the cmd directory:")
		fmt.Println("  - For training: go run cmd/train/main.go")
		fmt.Println("  - For visualization: go run cmd/visualization/main.go")
		os.Exit(0)
	}

	// Launch the UI-based game
	fmt.Println("Starting Othello game...")
	ui.RunUI()
}
