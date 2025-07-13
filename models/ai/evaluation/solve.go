package evaluation

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

type Cache struct {
	PECCache   map[string]PreEvaluationComputation
	ScoreCache map[string]int16
	MovesCache map[string]map[string]game.BitBoard
}

// NewCache creates a new cache with PEC priority
func NewCache() *Cache {

	return &Cache{
		PECCache:   make(map[string]PreEvaluationComputation),
		ScoreCache: make(map[string]int16),
		MovesCache: make(map[string]map[string]game.BitBoard),
	}
}

func Solve(b game.Board, player game.Piece, depth int8, eval Evaluation) ([]game.Position, int16) {
	return SolveWithStats(b, player, depth, eval, nil)
}

// Solve finds the best move for a player using minimax with alpha-beta pruning
func SolveWithStats(b game.Board, player game.Piece, depth int8, eval Evaluation, perfStats *stats.PerformanceStats) ([]game.Position, int16) {
	bb := utils.BoardToBits(b)
	validMoves := game.ValidMovesBitBoard(bb, player)
	if len(validMoves) == 0 {
		return []game.Position{{Row: -1, Col: -1}}, -1
	}

	// If only one move is available, return it immediately
	if len(validMoves) == 1 {
		bestMove := validMoves[0]
		newBoard, _ := game.GetNewBitBoardAfterMove(bb, bestMove, player)
		bestScore := eval.Evaluate(newBoard)
		return []game.Position{bestMove}, bestScore
	}

	var bestMoves []game.Position
	bestScore := MIN_EVAL - 64
	if player == game.Black {
		bestScore = MAX_EVAL + 64
	}
	alpha := MIN_EVAL - 64
	beta := MAX_EVAL + 64
	opponent := game.GetOtherPlayer(player).Color
	cache := NewCache() // Cache optimisé avec priorité PEC

	for _, move := range validMoves {
		newBoard, _ := game.GetNewBitBoardAfterMove(bb, move, player)
		childScore, childMoves := MMAB(newBoard, opponent, depth-1, alpha, beta, eval, cache, perfStats)

		if player == game.White {
			// Maximizing white player
			if childScore > bestScore {
				bestScore = childScore
				bestMoves = []game.Position{move}
				if childMoves != nil {
					bestMoves = append(bestMoves, childMoves...)
				}
			}

			if childScore > alpha {
				alpha = childScore
			}
		} else {
			// Minimizing black player
			if childScore < bestScore {
				bestScore = childScore
				bestMoves = []game.Position{move}
				if childMoves != nil {
					bestMoves = append(bestMoves, childMoves...)
				}
			}

			if childScore < beta {
				beta = childScore
			}
		}

	}
	return bestMoves, bestScore
}

// MMAB performs minimax search with alpha-beta pruning
func MMAB(node game.BitBoard, player game.Piece, depth int8, alpha, beta int16, eval Evaluation, cache *Cache, perfStats *stats.PerformanceStats) (score int16, path []game.Position) {

	hashStart := time.Now()
	boardHash := utils.HashBitBoard(node)
	if perfStats != nil {
		pecTime := time.Since(hashStart)
		perfStats.RecordOperation("hashBoard", pecTime, boardHash)
	}

	// Base case: leaf node or terminal position
	if depth == 0 {
		// Evaluate position
		var score int16
		evalStartTime := time.Now()
		if cachedScore, exists := cache.ScoreCache[boardHash]; exists {
			score = cachedScore
			if perfStats != nil {
				perfStats.RecordOperation("leaf_eval_cache_hit", time.Since(evalStartTime), boardHash)
			}
		} else {
			var pec PreEvaluationComputation
			pecTimeStart := time.Now()
			if c, exists := cache.PECCache[boardHash]; exists {
				pec = c
				if perfStats != nil {
					perfStats.RecordOperation("pec_cache_hit", time.Since(pecTimeStart), boardHash)
				}
			} else {
				pec = PrecomputeEvaluationBitBoard(node)
				if perfStats != nil {
					perfStats.RecordOperation("pec", time.Since(pecTimeStart), boardHash)
				}
				cache.PECCache[boardHash] = pec
				cache.MovesCache[boardHash] = make(map[string]game.BitBoard)
			}

			evalStartTime = time.Now()
			score = eval.PECEvaluate(node, pec)

			// Track evaluation time
			if perfStats != nil {
				perfStats.RecordOperation("leaf_eval", time.Since(evalStartTime), boardHash)
			}

			cache.ScoreCache[boardHash] = score
		}

		return score, nil
	}

	// Determine current player
	opponent := game.GetOtherPlayer(player).Color
	moves := game.ValidMovesBitBoard(node, player)

	// If no valid moves, pass turn
	if len(moves) == 0 {
		return MMAB(node, opponent, depth-1, alpha, beta, eval, cache, perfStats)
	}
	var bestMoves []game.Position
	bestScore := MIN_EVAL - 64
	if player == game.Black {
		bestScore = MAX_EVAL + 64
	}

	for _, move := range moves {
		algebraicMove := utils.PositionToAlgebraic(move)
		var newNode game.BitBoard
		moveStart := time.Now()
		if movesMap, exists := cache.MovesCache[boardHash]; exists {
			if b, exists := movesMap[algebraicMove]; exists {
				newNode = b
				if perfStats != nil {
					perfStats.RecordOperation("move_cache_hit", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
			} else {
				newNode, _ = game.GetNewBitBoardAfterMove(node, move, player)
				if perfStats != nil {
					perfStats.RecordOperation("move", time.Since(moveStart), algebraicMove+"-"+boardHash)
				}
				movesMap[algebraicMove] = newNode

			}
		} else {
			newNode, _ = game.GetNewBitBoardAfterMove(node, move, player)
			if perfStats != nil {
				perfStats.RecordOperation("move", time.Since(moveStart), algebraicMove+"-"+boardHash)
			}
			cache.MovesCache[boardHash] = map[string]game.BitBoard{algebraicMove: newNode}

		}
		// Recursive evaluation
		score, childMoves := MMAB(newNode, opponent, depth-1, alpha, beta, eval, cache, perfStats)

		if player == game.White {
			if score > bestScore {
				bestScore = score
				bestMoves = []game.Position{move}
				if childMoves != nil {
					bestMoves = append(bestMoves, childMoves...)
				}
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
		} else {
			if score < bestScore {
				bestScore = score
				bestMoves = []game.Position{move}
				if childMoves != nil {
					bestMoves = append(bestMoves, childMoves...)
				}
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

	}
	return bestScore, bestMoves

}
