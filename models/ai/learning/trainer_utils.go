package learning

import (
	"fmt"
	"sync"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
	"github.com/schollz/progressbar/v3"
)

// Consolidated type for training position tracking
type BatchPosition struct {
	board       game.Board
	player      game.Piece
	modelIndex  int
	openingIdx  int
	playerIndex int
}

// PlayMatchWithOpening plays a match between a model and a standard AI using a specific opening
// This is the central match playing function used by both tournament and evaluation
func PlayMatchWithOpening(
	model EvaluationModel,
	modelEval, standardEval evaluation.Evaluation,
	op opening.Opening,
	playerIndex, maxDepth int,
	collectPositions bool,
	modelIdx, openingIdx, playerPos int,
	positions *[]BatchPosition,
	mutex *sync.Mutex) (win, loss, draw bool) {

	// Create a new game
	g := game.NewGame("Model", "Standard")

	// Apply opening moves
	applyOpening(g, op)

	// Determine player model (alternate between games)
	playerModel := &g.Players[playerIndex]

	// Check if we're using a GPU evaluation
	useGPU := false
	if gpuEval, ok := modelEval.(*evaluation.GPUMixedEvaluation); ok {
		useGPU = true
		// Ensure coefficients are set in CUDA
		evaluation.SetCUDACoefficients(gpuEval.Coeffs)
	}

	// Play the game until completion
	for !game.IsGameFinished(g.Board) {
		if g.CurrentPlayer.Color == playerModel.Color {
			// Model player's turn
			if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
				var pos game.Position

				// If collecting positions for GPU batch processing
				if collectPositions && useGPU && mutex != nil {
					mutex.Lock()
					*positions = append(*positions, BatchPosition{
						board:       g.Board,
						player:      g.CurrentPlayer.Color,
						modelIndex:  modelIdx,
						openingIdx:  openingIdx,
						playerIndex: playerPos,
					})
					mutex.Unlock()
				}

				// Get the move based on evaluation type
				if useGPU {
					gpuPos, success := evaluation.GPUSolve(*g, g.CurrentPlayer, maxDepth)
					if success {
						pos = gpuPos
					} else {
						// Fall back to CPU search
						pos = evaluation.Solve(*g, g.CurrentPlayer, maxDepth, modelEval)
					}
				} else {
					// Use CPU search
					pos = evaluation.Solve(*g, g.CurrentPlayer, maxDepth, modelEval)
				}

				g.ApplyMove(pos)
			} else {
				// Skip turn if no valid moves
				g.CurrentPlayer = g.GetOtherPlayerMethod()
			}
		} else {
			// Standard player's turn
			if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
				// Always use CPU search for standard player for consistency
				pos := evaluation.Solve(*g, g.CurrentPlayer, maxDepth, standardEval)
				g.ApplyMove(pos)
			} else {
				// Skip turn if no valid moves
				g.CurrentPlayer = g.GetOtherPlayerMethod()
			}
		}
	}

	// Determine winner
	blackCount, whiteCount := game.CountPieces(g.Board)

	// Return result from model's perspective
	if playerModel.Color == game.Black {
		if blackCount > whiteCount {
			return true, false, false // Win
		} else if blackCount < whiteCount {
			return false, true, false // Loss
		}
		return false, false, true // Draw
	} else {
		if whiteCount > blackCount {
			return true, false, false // Win
		} else if whiteCount < blackCount {
			return false, true, false // Loss
		}
		return false, false, true // Draw
	}
}

// createProgressBar creates a standardized progress bar for training
func createProgressBar(totalMatches int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(totalMatches,
		progressbar.OptionSetDescription(description),
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
}

// applyOpening applies a predefined opening to a game
func applyOpening(g *game.Game, op opening.Opening) {
	transcript := op.Transcript

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

// crossoverCoefficients performs crossover on a specific coefficient array
func crossoverCoefficients(parent1, parent2 []int, pattern []bool) []int {
	result := make([]int, len(parent1))
	for i := range parent1 {
		if pattern[i%len(pattern)] {
			result[i] = parent1[i]
		} else {
			result[i] = parent2[i]
		}
	}
	return result
}

// SelectRandomOpenings selects a random subset of openings
func SelectRandomOpenings(numGames int) []opening.Opening {
	// Ensure we don't try to select more openings than available
	openingCount := len(opening.KNOWN_OPENINGS)
	if numGames > openingCount {
		numGames = openingCount
	}

	// For simplicity, just return the first numGames openings
	// In real implementation, you'd want to shuffle and select randomly
	return opening.KNOWN_OPENINGS[:numGames]
}

// evaluateModelsInParallel evaluates multiple models in parallel
func evaluateModelsInParallel(
	models []*EvaluationModel,
	createEvalFunc func(EvaluationModel) evaluation.Evaluation,
	maxDepth int,
	numGames int,
	stats *stats.PerformanceStats) {

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Select random openings based on numGames parameter
	selectedOpenings := SelectRandomOpenings(numGames)
	openingCount := len(selectedOpenings)

	// Calculate total number of matches to play (all models * selected openings * 2 player positions)
	totalMatches := len(models) * openingCount * 2

	// Create a single progress bar for all matches
	bar := createProgressBar(totalMatches, "Evaluating models")
	bar.RenderBlank()

	// Standard evaluation for opponent
	standardEval := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V1Coeff)

	// Prepare batch structures for GPU evaluation
	var positions []BatchPosition
	collectPositions := false

	// Check if we're using GPU
	if _, ok := createEvalFunc(*models[0]).(*evaluation.GPUMixedEvaluation); ok && evaluation.IsGPUAvailable() {
		collectPositions = true
		// Pre-allocate space for collecting positions
		positions = make([]BatchPosition, 0, totalMatches*30) // Estimate average positions per game
	}

	// Launch goroutines for each model
	for i := range models {
		wg.Add(1)
		go func(modelIdx int, model *EvaluationModel) {
			defer wg.Done()

			// Create a thread-local copy of performance stats for this goroutine
			localStats := NewPerformanceStats()

			// Reset statistics
			model.Wins = 0
			model.Losses = 0
			model.Draws = 0

			// Create custom evaluation function
			startEval := time.Now()
			evalFunc := createEvalFunc(*model)

			// Record evaluation creation time
			evalCreationTime := time.Since(startEval)
			localStats.RecordOperation("eval_creation", evalCreationTime)
			localStats.Counts["eval_created"] = 1

			// Play games against standard AI with selected openings
			for openingIdx, op := range selectedOpenings {
				for playerIdx := range 2 {
					startMatch := time.Now()

					// Play the match
					win, loss, draw := PlayMatchWithOpening(
						*model, evalFunc, standardEval, op, playerIdx, maxDepth,
						collectPositions, modelIdx, openingIdx, playerIdx, &positions, &mutex)

					// Record game result
					if win {
						model.Wins++
					} else if loss {
						model.Losses++
					} else if draw {
						model.Draws++
					}

					// Record match time in local stats
					matchTime := time.Since(startMatch)
					localStats.RecordOperation("match", matchTime)
					localStats.Counts["matches_played"] = 1

					// Update progress bar
					mutex.Lock()
					bar.Add(1)
					mutex.Unlock()
				}
			}

			// Calculate fitness score
			model.Fitness = float64(model.Wins) + float64(model.Draws)*0.5

			// Merge local stats into global stats - this requires mutex
			if stats != nil {
				mutex.Lock()
				stats.RecordOperation("eval_creation", localStats.OpTimes["eval_creation"])
				stats.RecordOperation("match", localStats.OpTimes["match"])
				stats.Counts["eval_created"] += localStats.Counts["eval_created"]
				stats.Counts["matches_played"] += totalMatches / len(models)
				mutex.Unlock()
			}
		}(i, models[i])
	}

	wg.Wait()
	fmt.Println() // Add newline after progress bar completes
}
