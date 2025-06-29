package evaluation

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

type MMABCacheNode struct {
	PrecomputedEvaluation PreEvaluationComputation
	Leaf                  bool
	Score                 int16
	Moves                 map[string]game.BitBoard
	RecursiveWhite        map[int8]int16
	RecursiveBlack        map[int8]int16
}

func Solve(b game.Board, player game.Piece, depth int8, eval Evaluation) (game.Position, int16) {
	return SolveWithStats(b, player, depth, eval, nil)
}

// Solve finds the best move for a player using minimax with alpha-beta pruning
func SolveWithStats(b game.Board, player game.Piece, depth int8, eval Evaluation, perfStats *stats.PerformanceStats) (game.Position, int16) {
	bb := utils.BoardToBits(b)
	validMoves := game.ValidMovesBitBoard(bb, player)
	if len(validMoves) == 0 {
		return game.Position{Row: -1, Col: -1}, -1
	}

	// If only one move is available, return it immediately
	if len(validMoves) == 1 {
		bestMove := validMoves[0]
		newBoard, _ := game.GetNewBitBoardAfterMove(bb, bestMove, player)
		bestScore := eval.Evaluate(newBoard)
		return bestMove, bestScore
	}

	var bestMove game.Position
	bestScore := MIN_EVAL - 1
	if player == game.Black {
		bestScore = MAX_EVAL + 1
	}
	alpha := MIN_EVAL - 1
	beta := MAX_EVAL + 1
	opponent := game.GetOtherPlayer(player).Color
	cache := make(map[string]MMABCacheNode)

	for _, move := range validMoves {
		newBoard, _ := game.GetNewBitBoardAfterMove(bb, move, player)
		childScore := MMAB(newBoard, opponent, depth-1, alpha, beta, eval, cache, perfStats)

		if player == game.White {
			// Maximizing white player
			if childScore > bestScore {
				bestScore = childScore
				bestMove = move
			}

			if childScore > alpha {
				alpha = childScore
			}
		} else {
			// Minimizing black player
			if childScore < bestScore {
				bestScore = childScore
				bestMove = move
			}

			if childScore < beta {
				beta = childScore
			}
		}

	}

	return bestMove, bestScore
}

// MMAB performs minimax search with alpha-beta pruning
func MMAB(node game.BitBoard, player game.Piece, depth int8, alpha, beta int16, eval Evaluation, cache map[string]MMABCacheNode, perfStats *stats.PerformanceStats) int16 {

	var cachedNode *MMABCacheNode

	hashStart := time.Now()
	boardHash := utils.HashBitBoard(node)
	if perfStats != nil {
		pecTime := time.Since(hashStart)
		perfStats.RecordOperation("hashBoard", pecTime, boardHash)
	}

	var pec PreEvaluationComputation
	pecTimeStart := time.Now()
	if c, exists := cache[boardHash]; exists {
		pec = c.PrecomputedEvaluation
		if perfStats != nil {
			perfStats.RecordOperation("pec_cache_hit", time.Since(pecTimeStart), boardHash)
		}
		cachedNode = &c
	} else {
		pec = PrecomputeEvaluationBitBoard(node)
		if perfStats != nil {
			perfStats.RecordOperation("pec", time.Since(pecTimeStart), boardHash)
		}
		cachedNode = &MMABCacheNode{
			PrecomputedEvaluation: pec,
			Leaf:                  false,
			Score:                 0,
			Moves:                 make(map[string]game.BitBoard),
		}
		cache[boardHash] = *cachedNode
	}
	// Base case: leaf node or terminal position
	if depth == 0 || pec.IsGameOver {
		// Evaluate position
		var score int16
		evalStartTime := time.Now()
		if cachedNode.Leaf {
			score = cachedNode.Score
			if perfStats != nil {
				perfStats.RecordOperation("leaf_eval_cache_hit", time.Since(evalStartTime), boardHash)
			}
		} else {
			score = eval.PECEvaluate(node, pec)

			// Track evaluation time
			if perfStats != nil {
				perfStats.RecordOperation("leaf_eval", time.Since(evalStartTime), boardHash)
			}

			cachedNode.Leaf = true
			cachedNode.Score = score
			cache[boardHash] = *cachedNode
		}

		return score
	}

	// Determine current player
	opponent := game.GetOtherPlayer(player).Color
	var moves []game.Position
	if player == game.White {
		moves = pec.WhiteValidMoves
	} else {
		moves = pec.BlackValidMoves
	}

	// If no valid moves, pass turn
	if len(moves) == 0 {
		return MMAB(node, opponent, depth-1, alpha, beta, eval, cache, perfStats)
	}

	if player == game.White {
		// Maximizing white
		bestScore := MIN_EVAL - 1
		for _, move := range moves {
			algebraicMove := utils.PositionToAlgebraic(move)
			var newNode game.BitBoard
			moveStart := time.Now()
			if b, exists := cachedNode.Moves[algebraicMove]; exists {
				newNode = b
				if perfStats != nil {
					perfStats.RecordOperation("move_cache_hit", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
			} else {
				newNode, _ = game.GetNewBitBoardAfterMove(node, move, player)
				if perfStats != nil {
					perfStats.RecordOperation("move", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
				cachedNode.Moves[algebraicMove] = newNode
				cache[boardHash] = *cachedNode
			}
			// Recursive evaluation
			score := MMAB(newNode, opponent, depth-1, alpha, beta, eval, cache, perfStats)

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
				if perfStats != nil {
					perfStats.RecordOperation("prune", 0, "")
				}
				break
			}
		}
		return bestScore
	} else {
		bestScore := MAX_EVAL + 1

		for _, move := range moves {

			algebraicMove := utils.PositionToAlgebraic(move)
			var newNode game.BitBoard
			moveStart := time.Now()
			if b, exists := cachedNode.Moves[algebraicMove]; exists {
				newNode = b
				if perfStats != nil {
					perfStats.RecordOperation("move_cache_hit", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
			} else {
				newNode, _ = game.GetNewBitBoardAfterMove(node, move, player)
				if perfStats != nil {
					perfStats.RecordOperation("move", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
				cachedNode.Moves[algebraicMove] = newNode
				cache[boardHash] = *cachedNode
			}

			// Recursive evaluation
			score := MMAB(newNode, opponent, depth-1, alpha, beta, eval, cache, perfStats)

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
				if perfStats != nil {
					perfStats.RecordOperation("prune", 0, "")
				}
				break
			}
		}
		return bestScore
	}
}
