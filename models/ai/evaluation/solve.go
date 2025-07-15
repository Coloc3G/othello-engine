package evaluation

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

type TTEntry struct {
	Score int16
	Depth int8
	Moves []game.Position
	Flag  int8 // 0: exact, 1: lower bound, 2: upper bound
}

type Cache struct {
	TTCache    map[string]TTEntry
	MaxEntries int
}

// NewCache creates a new cache with max entries limit
func NewCache() *Cache {
	return &Cache{
		TTCache:    make(map[string]TTEntry),
		MaxEntries: 20000000,
	}
}

func (c *Cache) cacheTTEntry(boardHash string, entry TTEntry) {
	if len(c.TTCache) >= c.MaxEntries {
		return
	}
	c.TTCache[boardHash] = entry
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

	cache.TTCache = make(map[string]TTEntry, 0)

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

	// Check transposition table first
	if ttEntry, exists := cache.TTCache[boardHash]; exists && ttEntry.Depth >= depth {
		ttHitStart := time.Now()

		switch ttEntry.Flag {
		case 0: // Exact value
			if perfStats != nil {
				perfStats.RecordOperation("tt_exact_hit", time.Since(ttHitStart), boardHash)
			}
			return ttEntry.Score, ttEntry.Moves
		case 1: // Lower bound
			if ttEntry.Score >= beta {
				if perfStats != nil {
					perfStats.RecordOperation("tt_lower_cutoff", time.Since(ttHitStart), boardHash)
				}
				return ttEntry.Score, ttEntry.Moves
			}
			if ttEntry.Score > alpha {
				alpha = ttEntry.Score
			}
		case 2: // Upper bound
			if ttEntry.Score <= alpha {
				if perfStats != nil {
					perfStats.RecordOperation("tt_upper_cutoff", time.Since(ttHitStart), boardHash)
				}
				return ttEntry.Score, ttEntry.Moves
			}
			if ttEntry.Score < beta {
				beta = ttEntry.Score
			}
		}

		if perfStats != nil {
			perfStats.RecordOperation("tt_partial_hit", time.Since(ttHitStart), boardHash)
		}
	}

	originalAlpha := alpha
	originalBeta := beta

	// Base case: leaf node or terminal position
	if depth == 0 {
		// Evaluate position
		var score int16
		pecTimeStart := time.Now()
		pec := PrecomputeEvaluationBitBoard(node)
		if perfStats != nil {
			perfStats.RecordOperation("pec", time.Since(pecTimeStart), boardHash)
		}
		evalStartTime := time.Now()
		score = eval.PECEvaluate(node, pec)
		if perfStats != nil {
			perfStats.RecordOperation("leaf_eval", time.Since(evalStartTime), boardHash)
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
		moveStart := time.Now()
		newNode, _ := game.GetNewBitBoardAfterMove(node, move, player)
		if perfStats != nil {
			perfStats.RecordOperation("move", time.Since(moveStart), algebraicMove+"-"+boardHash)
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

	// Store result in transposition table
	var flag int8
	if bestScore <= originalAlpha {
		flag = 2 // Upper bound
	} else if bestScore >= originalBeta {
		flag = 1 // Lower bound
	} else {
		flag = 0 // Exact value
	}

	cache.cacheTTEntry(boardHash, TTEntry{
		Score: bestScore,
		Depth: depth,
		Moves: bestMoves[:1],
		Flag:  flag,
	})

	return bestScore, bestMoves

}
