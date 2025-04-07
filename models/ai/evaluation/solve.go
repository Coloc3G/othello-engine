package evaluation

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
)

// Solve finds the best move for a player using minimax with alpha-beta pruning
func Solve(g game.Game, player game.Player, depth int, eval Evaluation) game.Position {
	// This is a simpler version without performance tracking
	bestScore := -1 << 31
	var bestMove game.Position

	validMoves := game.ValidMoves(g.Board, player.Color)
	for _, move := range validMoves {
		newBoard, _ := game.GetNewBoardAfterMove(g.Board, move, player)
		childScore := MMAB(g, newBoard, player, depth-1, false, -1<<31, 1<<31-1, eval, nil)
		if childScore > bestScore {
			bestScore = childScore
			bestMove = move
		}
	}

	return bestMove
}

// MMAB performs minimax search with alpha-beta pruning
func MMAB(g game.Game, node game.Board, player game.Player, depth int, max bool, alpha, beta int, eval Evaluation, perfStats *stats.PerformanceStats) int {
	// Base case: leaf node or terminal position
	if depth == 0 || game.IsGameFinished(node) {
		// Evaluate position
		evalStartTime := time.Now()
		score := eval.Evaluate(g, node, player)

		// Track evaluation time
		if perfStats != nil {
			evalTime := time.Since(evalStartTime)
			_, isGPUEval := eval.(*GPUMixedEvaluation)
			if isGPUEval && IsGPUAvailable() {
				perfStats.RecordOperation("gpu_leaf_eval", evalTime)
			} else {
				perfStats.RecordOperation("cpu_leaf_eval", evalTime)
			}
		}

		return score
	}

	var score int

	// If no valid moves, pass turn
	oplayer := game.GetOtherPlayer(g.Players, player.Color)
	if (max && !game.HasAnyMoves(node, player.Color)) || (!max && !game.HasAnyMoves(node, oplayer.Color)) {
		return MMAB(g, node, player, depth-1, !max, alpha, beta, eval, perfStats)
	}

	if max {
		score = -1 << 31
		for _, move := range game.ValidMoves(node, player.Color) {
			newNode, _ := game.GetNewBoardAfterMove(node, move, player)
			childScore := MMAB(g, newNode, player, depth-1, false, alpha, beta, eval, perfStats)
			if childScore > score {
				score = childScore
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
	} else {
		score = 1<<31 - 1
		for _, move := range game.ValidMoves(node, oplayer.Color) {
			newNode, _ := game.GetNewBoardAfterMove(node, move, oplayer)
			childScore := MMAB(g, newNode, player, depth-1, true, alpha, beta, eval, perfStats)
			if childScore < score {
				score = childScore
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
	}

	return score
}
