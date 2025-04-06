package main

import (
	"fmt"
	"math/rand"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/schollz/progressbar/v3"
)

// ComparisonStats holds statistics for AI version comparisons
type ComparisonStats struct {
	Version1Name   string
	Version2Name   string
	Version1Wins   int
	Version2Wins   int
	Draws          int
	TotalGames     int
	Version1WinPct float64
	Version2WinPct float64
	DrawPct        float64
}

// CompareCoefficients compares two sets of evaluation coefficients
func CompareCoefficients(coeff1, coeff2 evaluation.EvaluationCoefficients, numGames, searchDepth int) ComparisonStats {

	// Create stats object
	stats := ComparisonStats{
		Version1Name: coeff1.Name,
		Version2Name: coeff2.Name,
		TotalGames:   numGames,
	}

	// Create two evaluation functions with different coefficients
	eval1 := evaluation.NewMixedEvaluationWithCoefficients(coeff1)
	eval2 := evaluation.NewMixedEvaluationWithCoefficients(coeff2)

	// Create progress bar
	bar := progressbar.NewOptions(numGames,
		progressbar.OptionSetDescription("Playing tournament games"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Play games between the two AIs
	for i := 0; i < numGames; i++ {
		// Create a new game
		g := game.NewGame("Model1", "Model2")

		// Apply a random opening if available
		if len(opening.KNOWN_OPENINGS) > 0 {
			openingIndex := rand.Intn(len(opening.KNOWN_OPENINGS))
			selectedOpening := opening.KNOWN_OPENINGS[openingIndex]

			// Apply opening moves
			transcript := selectedOpening.Transcript
			for j := 0; j < len(transcript); j += 2 {
				if j+1 >= len(transcript) {
					break
				}

				// Parse algebraic notation (e.g., "c4")
				col := int(transcript[j] - 'a')
				row := int(transcript[j+1] - '1')
				pos := game.Position{Row: row, Col: col}

				// Apply the move
				g.Board, _ = game.ApplyMoveToBoard(g.Board, g.CurrentPlayer.Color, pos)
				g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
			}
		}

		// Alternate which AI goes first
		var firstEval, secondEval evaluation.Evaluation
		if i%2 == 0 {
			firstEval = eval1
			secondEval = eval2
		} else {
			firstEval = eval2
			secondEval = eval1
		}

		// Play the game
		var winner game.Piece
		gameOver := false

		for !gameOver {
			// Determine which evaluation to use
			var currentEval evaluation.Evaluation
			if g.CurrentPlayer.Color == game.Black {
				currentEval = firstEval
			} else {
				currentEval = secondEval
			}

			// Check if current player has valid moves
			validMoves := game.ValidMoves(g.Board, g.CurrentPlayer.Color)
			if len(validMoves) > 0 {
				// Get the best move using minimax search
				pos := evaluation.Solve(*g, g.CurrentPlayer, searchDepth, currentEval)
				g.ApplyMove(pos)
			} else {
				// Skip turn if no valid moves
				g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
			}

			// Check if game is over
			if game.IsGameFinished(g.Board) {
				gameOver = true
				winner = game.GetWinner(g.Board)
			}
		}

		// Record the result
		if winner == game.Empty {
			stats.Draws++
		} else {
			// Determine which version won
			version1Color := game.Black
			if i%2 != 0 {
				version1Color = game.White
			}

			if winner == version1Color {
				stats.Version1Wins++
			} else {
				stats.Version2Wins++
			}
		}

		// Update progress bar
		bar.Add(1)

		// Print progress (we'll keep this as a backup but the progress bar handles it)
		if (i+1)%100 == 0 || i == numGames-1 {
			fmt.Printf("Progress: %d/%d games completed\n", i+1, numGames)
		}
	}

	// Calculate percentages
	stats.Version1WinPct = float64(stats.Version1Wins) * 100.0 / float64(numGames)
	stats.Version2WinPct = float64(stats.Version2Wins) * 100.0 / float64(numGames)
	stats.DrawPct = float64(stats.Draws) * 100.0 / float64(numGames)

	return stats
}

// PrintComparison prints comparison statistics
func PrintComparison(stats ComparisonStats) {
	fmt.Println("\n==== Comparison Results ====")
	fmt.Printf("%s vs %s (Total games: %d)\n", stats.Version1Name, stats.Version2Name, stats.TotalGames)
	fmt.Printf("%s wins: %d (%.1f%%)\n", stats.Version1Name, stats.Version1Wins, stats.Version1WinPct)
	fmt.Printf("%s wins: %d (%.1f%%)\n", stats.Version2Name, stats.Version2Wins, stats.Version2WinPct)
	fmt.Printf("Draws: %d (%.1f%%)\n", stats.Draws, stats.DrawPct)

	// Print summary judgment
	fmt.Print("Conclusion: ")
	if stats.Version1Wins > stats.Version2Wins {
		diff := stats.Version1Wins - stats.Version2Wins
		fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
			stats.Version1Name, diff, float64(diff)*100.0/float64(stats.TotalGames))
	} else if stats.Version2Wins > stats.Version1Wins {
		diff := stats.Version2Wins - stats.Version1Wins
		fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
			stats.Version2Name, diff, float64(diff)*100.0/float64(stats.TotalGames))
	} else {
		fmt.Println("Both versions appear equally matched")
	}
	fmt.Println("===========================")
}

// CompareVersionsWithOpenings runs a comparison using random openings
func CompareVersionsWithOpenings(numGames, searchDepth int) {
	fmt.Println("Comparing AI versions with random openings...")
	stats := CompareCoefficients(evaluation.V1Coeff, evaluation.V2Coeff, numGames, searchDepth)
	PrintComparison(stats)
}

// CompareVersions runs a comparison without using openings
func CompareVersions(numGames, searchDepth int) {
	// Temporarily disable openings
	originalOpenings := opening.KNOWN_OPENINGS
	opening.KNOWN_OPENINGS = []opening.Opening{}

	fmt.Println("Comparing AI versions without openings...")
	stats := CompareCoefficients(evaluation.V1Coeff, evaluation.V2Coeff, numGames, searchDepth)
	PrintComparison(stats)

	// Restore openings
	opening.KNOWN_OPENINGS = originalOpenings
}
