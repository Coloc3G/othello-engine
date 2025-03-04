package game

import "fmt"

func (g *Game) DisplayBoard() {
	// Display column numbers
	fmt.Print("  ")
	for i := 0; i < 8; i++ {
		fmt.Printf(" %d", i)
	}
	fmt.Println()

	// Display board with row numbers
	for i := range g.Board {
		fmt.Printf("%d |", i)
		for j := range g.Board[i] {
			switch g.Board[i][j] {
			case Empty:
				fmt.Print(" ·")
			case Black:
				fmt.Print(" ●")
			case White:
				fmt.Print(" ○")
			}
		}
		fmt.Println()
	}
}
