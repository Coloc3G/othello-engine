package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// parseMove parses user input to extract the row and column coordinates
// Input is expected in chess notation (e.g., "E4")
func parseMove(input string) (game.Position, error) {
	// Remove any spaces
	input = strings.TrimSpace(input)
	input = strings.ToUpper(input)

	if len(input) != 2 {
		return game.Position{}, fmt.Errorf("invalid input format: please use chess notation (e.g. 'E4')")
	}

	// Parse column (letter A-H)
	col := int(input[0] - 'A')
	if col < 0 || col >= 8 {
		return game.Position{}, fmt.Errorf("invalid column: must be between A and H")
	}

	// Parse row (number 1-8)
	row, err := strconv.Atoi(string(input[1]))
	if err != nil || row < 1 || row > 8 {
		return game.Position{}, fmt.Errorf("invalid row: must be between 1 and 8")
	}

	// Convert to 0-based index
	row--

	return game.Position{Row: row, Col: col}, nil
}

func RunGame() {
	// Initialize the game
	g := game.NewGame()
	reader := bufio.NewReader(os.Stdin)

	// Main game loop
	gameOver := false
	skipTurn := false

	for !gameOver {
		// Display current game state
		fmt.Printf("\nCurrent player: %s\n", g.CurrentPlayer.Name)
		g.DisplayBoard(g.Board)

		// Get valid moves for current player
		validMoves := g.GetValidMovesForCurrentPlayer()

		// Check if the player has any valid moves
		if len(validMoves) == 0 {
			if skipTurn {
				// Both players had to skip, game is over
				gameOver = true
				fmt.Println("Neither player has any valid moves. Game over!")
				continue
			} else {
				// Current player has to skip
				fmt.Printf("\n%s has no valid moves. Turn passes to %s.\n",
					g.CurrentPlayer.Name,
					game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color).Name)
				g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
				skipTurn = true
				continue
			}
		}

		// Reset skip flag since current player has moves
		skipTurn = false

		// Display valid moves in chess notation
		fmt.Printf("\nValid moves: ")
		for i, pos := range validMoves {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%c%d", 'A'+pos.Col, pos.Row+1)
		}
		fmt.Println()

		// Get player move
		var movePos game.Position
		validInput := false

		for !validInput {
			var pos game.Position
			var err error
			if g.CurrentPlayer.Name == "AI" {
				// AI player
				openingFound := false
				if matches := opening.MatchOpening(utils.PositionsToAlgebraic(g.History)); len(matches) > 0 {
					maxL := 0
					for _, match := range matches {
						if len(match.Transcript) > 2*len(g.History) && len(match.Transcript) > maxL {
							maxL = len(match.Transcript)
							pos = utils.AlgebraicToPositions(match.Transcript)[len(g.History)]
							openingFound = true
						}
					}
				}
				if !openingFound {
					eval := evaluation.NewMixedEvaluation()
					pos = evaluation.Solve(g, g.CurrentPlayer, 5, eval)
				}
				fmt.Println("AI move: ", utils.PositionToAlgebraic(pos))
			} else {
				fmt.Printf("\nEnter your move in chess notation (e.g. 'E4'): ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				// Check for quit command
				if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
					fmt.Println("Game aborted.")
					return
				}

				// Parse the input
				pos, err = parseMove(input)
				if err != nil {
					fmt.Printf("Error: %s. Please try again.\n", err)
					continue
				}
			}

			// Check if the move is valid
			isValid := false
			for _, validPos := range validMoves {
				if validPos.Row == pos.Row && validPos.Col == pos.Col {
					isValid = true
					break
				}
			}

			if !isValid {
				fmt.Println("Error: That position is not a valid move. Please try again.")
				continue
			}

			movePos = pos
			validInput = true
		}

		// Apply the move
		success := g.ApplyMove(movePos)
		if !success {
			fmt.Println("Error: Failed to apply move. This shouldn't happen!")
			return
		}

		// Check if game is finished after this move
		if game.IsGameFinished(g.Board) {
			gameOver = true
		}
	}

	// Game is over, display final state and results
	fmt.Println("\nGame is over!")
	g.DisplayBoard(g.Board)

	// Count pieces and determine winner
	blackCount, whiteCount := game.CountPieces(g.Board)
	fmt.Printf("\nFinal score:\n")
	fmt.Printf("Black: %d\n", blackCount)
	fmt.Printf("White: %d\n", whiteCount)

	winner := game.GetWinner(g.Board)
	switch winner {
	case game.Black:
		fmt.Println("Black wins!")
	case game.White:
		fmt.Println("White wins!")
	default:
		fmt.Println("The game is a tie!")
	}
}

func main() {
	fmt.Println("Welcome to Othello!")
	fmt.Println("Enter moves in chess notation (e.g. 'E4')")
	fmt.Println("Type 'quit' or 'exit' to end the game")
	RunGame()
}
