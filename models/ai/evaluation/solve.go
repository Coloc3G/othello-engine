package evaluation

import (
	"sort"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
)

// Solve finds the best move for a player using minimax with alpha-beta pruning
func Solve(g game.Game, player game.Player, depth int, eval Evaluation) game.Position {
	bestScore := -1 << 31
	var bestMove game.Position

	validMoves := game.ValidMoves(g.Board, player.Color)
	if len(validMoves) == 0 {
		return game.Position{Row: -1, Col: -1}
	}

	// If only one move is available, return it immediately
	if len(validMoves) == 1 {
		return validMoves[0]
	}

	// Sort moves by row,col for deterministic ordering that matches CUDA implementation
	sort.Slice(validMoves, func(i, j int) bool {
		if validMoves[i].Row == validMoves[j].Row {
			return validMoves[i].Col < validMoves[j].Col
		}
		return validMoves[i].Row < validMoves[j].Row
	})

	// Use same alpha-beta bounds as CUDA implementation
	alpha := -1 << 31
	beta := 1<<31 - 1

	for _, move := range validMoves {
		newBoard, _ := game.GetNewBoardAfterMove(g.Board, move, player)
		childScore := MMAB(g, newBoard, player, depth-1, false, alpha, beta, eval, nil)

		if childScore > bestScore {
			bestScore = childScore
			bestMove = move
		}

		// Update alpha for pruning - must match CUDA implementation
		if childScore > alpha {
			alpha = childScore
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

	// Determine current player
	var moves []game.Position
	if max {
		// Maximizing player (our player)
		moves = game.ValidMoves(node, player.Color)
	} else {
		// Minimizing player (opponent)
		opponent := game.GetOtherPlayer(g.Players, player.Color)
		moves = game.ValidMoves(node, opponent.Color)
	}

	// If no valid moves, pass turn
	if len(moves) == 0 {
		return MMAB(g, node, player, depth-1, !max, alpha, beta, eval, perfStats)
	}

	// Sort moves by row,col for deterministic ordering that matches CUDA implementation
	sort.Slice(moves, func(i, j int) bool {
		if moves[i].Row == moves[j].Row {
			return moves[i].Col < moves[j].Col
		}
		return moves[i].Row < moves[j].Row
	})

	if max {
		// Maximizing player (our player)
		bestScore := -1 << 31
		for _, move := range moves {
			// Create new board with this move
			newNode, _ := game.GetNewBoardAfterMove(node, move, player)

			// Recursive evaluation
			score := MMAB(g, newNode, player, depth-1, false, alpha, beta, eval, perfStats)

			// Update best score
			if score > bestScore {
				bestScore = score
			}

			// Update alpha for pruning
			if score > alpha {
				alpha = score
			}

			// Alpha-beta pruning
			if beta <= alpha {
				break
			}
		}
		return bestScore
	} else {
		// Minimizing player (opponent)
		bestScore := 1<<31 - 1
		opponent := game.GetOtherPlayer(g.Players, player.Color)

		for _, move := range moves {
			// Create new board with this move
			newNode, _ := game.GetNewBoardAfterMove(node, move, opponent)

			// Recursive evaluation
			score := MMAB(g, newNode, player, depth-1, true, alpha, beta, eval, perfStats)

			// Update best score
			if score < bestScore {
				bestScore = score
			}

			// Update beta for pruning
			if score < beta {
				beta = score
			}

			// Alpha-beta pruning
			if beta <= alpha {
				break
			}
		}
		return bestScore
	}
}
