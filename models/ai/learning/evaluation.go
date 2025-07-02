package learning

import (
	"fmt"
	"sync"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
	"github.com/schollz/progressbar/v3"
)

// PlayMatchWithOpening plays a match between a model and a standard AI using a specific opening
// This is the central match playing function used by evaluation
func PlayMatchWithOpening(
	modelEval, standardEval evaluation.Evaluation,
	op opening.Opening,
	playerIndex int, maxDepth int8) (win, loss, draw bool, history []game.Position) {
	// Create a new game
	g := game.NewGame("Black", "White")
	var blackCount, whiteCount int
	modelColor := game.Black
	if playerIndex == 1 {
		modelColor = game.White
	}

	// Apply opening moves
	applyOpening(g, op)

	for !game.IsGameFinished(g.Board) {
		// Determine which evaluation to use
		var currentEval evaluation.Evaluation

		if g.CurrentPlayer.Color == modelColor {
			currentEval = modelEval
		} else {
			currentEval = standardEval
		}

		// Check if current player has valid moves
		validMoves := game.ValidMoves(g.Board, g.CurrentPlayer.Color)
		if len(validMoves) > 0 {
			// Get the best move using minimax search
			pos, _ := evaluation.Solve(g.Board, g.CurrentPlayer.Color, maxDepth, currentEval)
			g.ApplyMove(pos)
		} else {
			// Skip turn if no valid moves
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		}
	}

	blackCount, whiteCount = game.CountPieces(g.Board)

	// Return result from model's perspective
	if modelColor == game.Black {
		if blackCount > whiteCount {
			return true, false, false, g.History // Win
		} else if blackCount < whiteCount {
			return false, true, false, g.History // Loss
		}
		return false, false, true, g.History // Draw
	} else {
		if whiteCount > blackCount {
			return true, false, false, g.History // Win
		} else if whiteCount < blackCount {
			return false, true, false, g.History // Loss
		}
		return false, false, true, g.History // Draw
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
	// Apply the moves from the opening transcript
	for _, move := range utils.AlgebraicToPositions(op.Transcript) {
		g.ApplyMove(move)
	}
}

// evaluateModelsInParallel evaluates multiple models in parallel
func evaluateModelsInParallel(
	models []*EvaluationModel,
	baseModel evaluation.EvaluationCoefficients,
	maxDepth int8,
	numGames int) {

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Calculate total number of matches to play (all models * selected openings * 2 player positions)
	openingCount := min(numGames, len(opening.KNOWN_OPENINGS))
	selectedOpenings := opening.SelectRandomOpenings(openingCount)
	totalMatches := len(models) * openingCount * 2

	// Create a single progress bar for all matches
	bar := createProgressBar(totalMatches, "Evaluating models")
	bar.RenderBlank()

	standardEval := evaluation.NewMixedEvaluation(baseModel)

	// Launch goroutines for each model
	for i := range models {
		wg.Add(1)
		go func(modelIdx int, model *EvaluationModel) {
			defer wg.Done()

			// Reset statistics
			model.Wins = 0
			model.Losses = 0
			model.Draws = 0
			model.BlackGames = make(map[string]string, 0)
			model.WhiteGames = make(map[string]string, 0)
			evalFunc := evaluation.NewMixedEvaluation(model.Coeffs)

			// Play games against standard AI with selected openings
			for _, op := range selectedOpenings {
				for playerIdx := range 2 {

					// Play the match
					win, loss, draw, history := PlayMatchWithOpening(
						evalFunc, standardEval, op, playerIdx, maxDepth)

					// Store the game history
					historyString := utils.PositionsToAlgebraic(history)
					if playerIdx == 0 {
						model.BlackGames[op.Name] = historyString
					} else {
						model.WhiteGames[op.Name] = historyString
					}

					// Record game result
					if win {
						model.Wins++
					} else if loss {
						model.Losses++
					} else if draw {
						model.Draws++
					}
					// Update progress bar
					mutex.Lock()
					bar.Add(1)
					mutex.Unlock()
				}
			}

			// Calculate fitness score
			model.Fitness = float64(model.Wins) + float64(model.Draws)*0.5

		}(i, models[i])
	}

	wg.Wait()
	fmt.Println() // Add newline after progress bar completes
}
