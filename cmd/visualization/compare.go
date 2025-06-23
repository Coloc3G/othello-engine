package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/schollz/progressbar/v3"
)

// CompareCoefficients compares two sets of evaluation coefficients concurrently
func CompareCoefficients(coeff1, coeff2 evaluation.EvaluationCoefficients, numGames, searchDepth int) PerformanceResult {

	// Create stats object
	stats := PerformanceResult{
		Version1Name: coeff1.Name,
		Version2Name: coeff2.Name,
		TotalGames:   numGames,
	}

	// Create two evaluation functions with different coefficients
	eval1 := evaluation.NewMixedEvaluation(coeff1)
	eval2 := evaluation.NewMixedEvaluation(coeff2)

	// Create progress bar
	bar := progressbar.NewOptions(numGames,
		progressbar.OptionSetDescription("Playing games"),
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
	bar.RenderBlank()

	// Set up job and result channels and a worker pool
	jobsCh := make(chan int, numGames)
	resultsCh := make(chan int, numGames)

	for i := range numGames {
		jobsCh <- i
	}
	close(jobsCh)

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobsCh {
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
						pos, _ := evaluation.Solve(*g, g.CurrentPlayer, searchDepth, currentEval)
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

				// Determine outcome: 0 = draw, 1 = Version1 win, 2 = Version2 win
				var outcome int
				version1Color := game.Black
				if i%2 != 0 {
					version1Color = game.White
				}
				if winner == game.Empty {
					outcome = 0
				} else if winner == version1Color {
					outcome = 1
				} else {
					outcome = 2
				}
				bar.Add(1)
				resultsCh <- outcome
			}
		}()
	}

	wg.Wait()
	close(resultsCh)

	// Collect results and update statistics with progress bar updates
	for outcome := range resultsCh {
		if outcome == 0 {
			stats.Draws++
		} else if outcome == 1 {
			stats.Version1Wins++
		} else {
			stats.Version2Wins++
		}
		bar.Add(1)
	}

	// Calculate percentages
	stats.Version1WinPct = float64(stats.Version1Wins) * 100.0 / float64(numGames)
	stats.Version2WinPct = float64(stats.Version2Wins) * 100.0 / float64(numGames)
	stats.DrawPct = float64(stats.Draws) * 100.0 / float64(numGames)

	return stats
}

// PrintComparison prints comparison statistics
func PrintComparison(stats PerformanceResult) {
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

func CompareVersions(numGames, searchDepth int) (results []PerformanceResult) {

	stats := CompareCoefficients(evaluation.V4Coeff, evaluation.V1Coeff, numGames, searchDepth)
	results = append(results, stats)

	return
}
