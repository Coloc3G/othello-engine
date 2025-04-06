package evaluation

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/cache"
	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
)

// SolveWithStats is like Solve but also collects performance statistics
func SolveWithStats(g game.Game, player game.Player, depth int, eval Evaluation, perfStats *stats.PerformanceStats) game.Position {
	startTime := time.Now()

	// Get the evaluation cache
	boardCache := cache.GetGlobalCache()

	// Try to find a cached result for this board
	_, cachedMove, source, found := boardCache.Lookup(g.Board, player.Color, depth)
	if found && cachedMove.Row >= 0 && cachedMove.Row < 8 && cachedMove.Col >= 0 && cachedMove.Col < 8 {
		// Verify that the cached move is still valid
		if game.IsValidMove(g.Board, player.Color, cachedMove) {
			// Record cache hit performance
			if perfStats != nil {
				cacheHitTime := time.Since(startTime)
				perfStats.RecordOperation("cache_hit", cacheHitTime)

				// Record source of the cache hit
				if source == cache.SourceCPU {
					perfStats.RecordOperation("cpu_cache_hit", 0)
				} else if source == cache.SourceGPU {
					perfStats.RecordOperation("gpu_cache_hit", 0)
				}
			}

			return cachedMove
		}
	}

	// Log cache miss if stats available
	if perfStats != nil {
		perfStats.RecordOperation("cache_miss", 0)
	}

	// No valid cache hit, compute the best move
	bestScore := -1 << 31
	var bestMove game.Position

	// Track which source actually computed this evaluation
	var evaluationSource cache.CacheSource = cache.SourceCPU // Default to CPU

	// Check if we're using a GPU evaluation
	gpuEval, isGPUEval := eval.(*GPUMixedEvaluation)
	cpuEvalTime := time.Duration(0)

	validMoves := game.ValidMoves(g.Board, player.Color)
	for _, move := range validMoves {
		newBoard, _ := game.GetNewBoardAfterMove(g.Board, move, player)

		var childScore int

		// Use GPU or CPU evaluation as appropriate
		if isGPUEval && IsGPUAvailable() {
			// Try GPU evaluation first
			gpuStartTime := time.Now()

			// Try to use GPU acceleration
			scores := EvaluateStatesCUDA([]game.Board{newBoard}, []game.Piece{player.Color})

			if len(scores) > 0 {
				// GPU evaluation succeeded
				childScore = scores[0]
				evaluationSource = cache.SourceGPU

				// Log GPU performance if stats available
				if perfStats != nil {
					gpuTime := time.Since(gpuStartTime)
					perfStats.RecordOperation("gpu_eval", gpuTime)
					perfStats.GPUSuccesses++
				}
			} else {
				// GPU evaluation failed, fall back to CPU
				fallbackStartTime := time.Now()
				childScore = MMAB(g, newBoard, player, depth-1, false, -1<<31, 1<<31-1, gpuEval, perfStats)
				evaluationSource = cache.SourceCPU

				// Log fallback performance if stats available
				if perfStats != nil {
					fallbackTime := time.Since(fallbackStartTime)
					perfStats.RecordOperation("gpu_fallback", fallbackTime)
					cpuEvalTime += fallbackTime
				}
			}
		} else {
			// Use CPU evaluation
			cpuStartTime := time.Now()
			childScore = MMAB(g, newBoard, player, depth-1, false, -1<<31, 1<<31-1, eval, perfStats)

			// Log CPU performance if stats available
			if perfStats != nil {
				cpuTime := time.Since(cpuStartTime)
				perfStats.RecordOperation("cpu_eval", cpuTime)
				cpuEvalTime += cpuTime
			}
		}

		// Update best move
		if childScore > bestScore {
			bestScore = childScore
			bestMove = move
		}
	}

	// Cache the result with the correct source
	boardCache.Store(g.Board, player.Color, depth, bestScore, bestMove, evaluationSource)

	// Log total solve time if stats available
	if perfStats != nil {
		totalTime := time.Since(startTime)
		perfStats.RecordOperation("solve_total", totalTime)

		// Record which was dominant in this evaluation
		if isGPUEval && cpuEvalTime < totalTime/2 {
			perfStats.RecordOperation("solve_gpu_dominant", 0)
		} else {
			perfStats.RecordOperation("solve_cpu_dominant", 0)
		}
	}

	return bestMove
}

// Solve finds the best move for a player using minimax with alpha-beta pruning
func Solve(g game.Game, player game.Player, depth int, eval Evaluation) game.Position {
	// This is a simpler version without performance tracking
	// Get the evaluation cache
	boardCache := cache.GetGlobalCache()

	// Try to find a cached result for this board
	_, cachedMove, _, found := boardCache.Lookup(g.Board, player.Color, depth)
	if found && cachedMove.Row >= 0 && cachedMove.Row < 8 && cachedMove.Col >= 0 && cachedMove.Col < 8 {
		// Verify that the cached move is still valid
		if game.IsValidMove(g.Board, player.Color, cachedMove) {
			return cachedMove
		}
	}

	// No valid cache hit, compute the best move
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

	// Cache the result
	boardCache.Store(g.Board, player.Color, depth, bestScore, bestMove, cache.SourceCPU)

	return bestMove
}

// MMAB performs minimax search with alpha-beta pruning
func MMAB(g game.Game, node game.Board, player game.Player, depth int, max bool, alpha, beta int, eval Evaluation, perfStats *stats.PerformanceStats) int {
	// Get the evaluation cache
	boardCache := cache.GetGlobalCache()

	// Check if position is already cached with sufficient depth
	score, _, source, found := boardCache.Lookup(node, player.Color, depth)
	if found {
		// Track cache hits if stats available
		if perfStats != nil {
			if source == cache.SourceCPU {
				perfStats.RecordOperation("cpu_cache_hit", 0)
			} else if source == cache.SourceGPU {
				perfStats.RecordOperation("gpu_cache_hit", 0)
			}
		}
		return score
	}

	// Track cache misses if stats available
	if perfStats != nil {
		perfStats.RecordOperation("cache_miss", 0)
	}

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

	// If no valid moves, pass turn
	oplayer := game.GetOtherPlayer(g.Players, player.Color)
	if (max && !game.HasAnyMoves(node, player.Color)) || (!max && !game.HasAnyMoves(node, oplayer.Color)) {
		return MMAB(g, node, player, depth-1, !max, alpha, beta, eval, perfStats)
	}

	var bestMove game.Position
	// Track which source actually computed this evaluation
	var evaluationSource cache.CacheSource = cache.SourceCPU

	if max {
		score = -1 << 31
		for _, move := range game.ValidMoves(node, player.Color) {
			newNode, _ := game.GetNewBoardAfterMove(node, move, player)
			childScore := MMAB(g, newNode, player, depth-1, false, alpha, beta, eval, perfStats)
			if childScore > score {
				score = childScore
				bestMove = move
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
				bestMove = move
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
	}

	// Check if this was evaluated with GPU
	if _, isGPUEval := eval.(*GPUMixedEvaluation); isGPUEval && IsGPUAvailable() {
		evaluationSource = cache.SourceGPU
	}

	// Cache the result
	boardCache.Store(node, player.Color, depth, score, bestMove, evaluationSource)

	return score
}
