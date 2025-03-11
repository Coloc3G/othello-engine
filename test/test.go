package test

import (
	"fmt"
	"sync"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// GameResult stores the result of a single game
type GameResult struct {
	Version1Win bool
	Version2Win bool
	Draw        bool
	Duration    time.Duration
}

// ComparisonStats stores the statistics from comparing two versions
type ComparisonStats struct {
	Version1Wins   int
	Version2Wins   int
	Draws          int
	TotalGames     int
	AvgGameTime    time.Duration
	Version1WinPct float64
	Version2WinPct float64
	DrawPct        float64
}

// CompareCoefficients runs N games between two coefficient versions and returns statistics
func CompareCoefficients(version1 evaluation.EvaluationCoefficients, version2 evaluation.EvaluationCoefficients, numGames int, searchDepth int) ComparisonStats {
	var stats ComparisonStats
	stats.TotalGames = numGames

	results := make(chan GameResult, numGames)
	var wg sync.WaitGroup

	// Start half the games with version1 as white, half with version2 as white
	for i := 0; i < numGames; i++ {
		wg.Add(1)
		go func(gameNum int) {
			defer wg.Done()

			// Alternate which version plays white
			isVersion1White := gameNum%2 == 0

			start := time.Now()
			result := playGame(version1, version2, isVersion1White, searchDepth)
			duration := time.Since(start)

			results <- GameResult{
				Version1Win: result == 1,
				Version2Win: result == 2,
				Draw:        result == 0,
				Duration:    duration,
			}
		}(i)
	}

	// Close results channel when all games are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var totalDuration time.Duration
	for result := range results {
		if result.Version1Win {
			stats.Version1Wins++
		} else if result.Version2Win {
			stats.Version2Wins++
		} else {
			stats.Draws++
		}
		totalDuration += result.Duration
	}

	// Calculate statistics
	stats.AvgGameTime = totalDuration / time.Duration(numGames)
	stats.Version1WinPct = float64(stats.Version1Wins) * 100 / float64(numGames)
	stats.Version2WinPct = float64(stats.Version2Wins) * 100 / float64(numGames)
	stats.DrawPct = float64(stats.Draws) * 100 / float64(numGames)

	return stats
}

// playGame runs a single game between two coefficient versions
// Returns: 1 for version1 win, 2 for version2 win, 0 for draw
func playGame(version1 evaluation.EvaluationCoefficients, version2 evaluation.EvaluationCoefficients, isVersion1White bool, searchDepth int) int {
	// Create the game
	g := game.NewGame("AI", "AI")

	// Apply random opening
	applyRandomOpening(g)

	// Create the AI players
	eval1 := evaluation.NewMixedEvaluationWithCoefficients(version1)
	eval2 := evaluation.NewMixedEvaluationWithCoefficients(version2)

	var whiteEval, blackEval evaluation.Evaluation
	var version1Color game.Piece

	if isVersion1White {
		whiteEval = eval1
		blackEval = eval2
		version1Color = game.White
	} else {
		whiteEval = eval2
		blackEval = eval1
		version1Color = game.Black
	}

	// Play the game until it's finished
	for !game.IsGameFinished(g.Board) {
		var move game.Position

		if g.CurrentPlayer.Color == game.White {
			move = evaluation.Solve(*g, g.CurrentPlayer, searchDepth, whiteEval)
		} else {
			move = evaluation.Solve(*g, g.CurrentPlayer, searchDepth, blackEval)
		}

		if move.Row < 0 || !game.IsValidMove(g.Board, g.CurrentPlayer.Color, move) {
			// If no valid moves, pass the turn
			g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
			continue
		}

		// Make the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer)
		g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
	}

	// Determine the winner
	whiteCount, blackCount := 0, 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if g.Board[i][j] == game.White {
				whiteCount++
			} else if g.Board[i][j] == game.Black {
				blackCount++
			}
		}
	}

	if whiteCount > blackCount {
		if version1Color == game.White {
			return 1 // Version 1 wins
		} else {
			return 2 // Version 2 wins
		}
	} else if blackCount > whiteCount {
		if version1Color == game.Black {
			return 1 // Version 1 wins
		} else {
			return 2 // Version 2 wins
		}
	} else {
		return 0 // Draw
	}
}

// applyRandomOpening selects a random opening and applies it to the game
func applyRandomOpening(g *game.Game) {
	randomOpening := opening.SelectRandomOpening()
	transcript := randomOpening.Transcript

	// Apply the moves from the opening transcript
	for i := 0; i < len(transcript); i += 2 {
		if i+1 >= len(transcript) {
			break // Ensure we have a complete move (row and column)
		}

		move := utils.AlgebraicToPosition(transcript[i : i+2])

		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer)
		g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
	}
}

// PrintComparison prints the comparison statistics in a readable format
func PrintComparison(stats ComparisonStats) {
	fmt.Println("=== Coefficient Comparison Results ===")
	fmt.Printf("Total games: %d\n", stats.TotalGames)
	fmt.Printf("Version 1 wins: %d (%.2f%%)\n", stats.Version1Wins, stats.Version1WinPct)
	fmt.Printf("Version 2 wins: %d (%.2f%%)\n", stats.Version2Wins, stats.Version2WinPct)
	fmt.Printf("Draws: %d (%.2f%%)\n", stats.Draws, stats.DrawPct)
	fmt.Printf("Average game time: %v\n", stats.AvgGameTime)

	// Determine which version performed better
	if stats.Version1WinPct > stats.Version2WinPct {
		fmt.Printf("Version 1 outperformed Version 2 by %.2f percentage points\n",
			stats.Version1WinPct-stats.Version2WinPct)
	} else if stats.Version2WinPct > stats.Version1WinPct {
		fmt.Printf("Version 2 outperformed Version 1 by %.2f percentage points\n",
			stats.Version2WinPct-stats.Version1WinPct)
	} else {
		fmt.Println("Both versions performed equally")
	}
}

// CompareVersions creates a test function to compare the predefined coefficient versions
func CompareVersions(numGames int, searchDepth int) {
	fmt.Println("Comparing V1 vs V2 coefficients...")
	stats := CompareCoefficients(evaluation.V1Coeff, evaluation.V2Coeff, numGames, searchDepth)
	PrintComparison(stats)
}

// CompareVersionsWithOpenings allows comparing coefficient versions using random openings
func CompareVersionsWithOpenings(numGames int, searchDepth int) {
	fmt.Println("Comparing coefficient versions with random openings...")
	fmt.Println("Comparing V1 vs V2 coefficients...")
	stats := CompareCoefficients(evaluation.V1Coeff, evaluation.V2Coeff, numGames, searchDepth)
	PrintComparison(stats)
}
