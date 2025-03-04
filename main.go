package main

import (
	"fmt"

	"github.com/Coloc3G/othello-engine/models/game"
)

func RunGame() {
	// Initialize the game
	g := game.NewGame()

	// Main game loop
	for {
		// Display current game state
		fmt.Printf("\nCurrent player: %s\n", g.CurrentPlayer.Name)
		g.DisplayBoard()
		fmt.Printf("\nList of valid moves: %v\n", g.GetValidMovesForCurrentPlayer())
		// Get player move (temporary: just display a message)
		fmt.Printf("\nWaiting for %s's move...\n", g.CurrentPlayer.Name)

		// TODO: Implement move validation and execution
		// TODO: Implement game end conditions

		// Temporary: break after first iteration
		break
	}

	fmt.Println("\nGame ended!")
}

func main() {
	fmt.Println("Welcome to Othello!")
	RunGame()
}
