package evaluation

import (
	"fmt"
	"sync"

	"github.com/Coloc3G/othello-engine/models/ai/cache"
	"github.com/Coloc3G/othello-engine/models/game"
)

// GPUMixedEvaluation is a version of MixedEvaluation that uses GPU acceleration
type GPUMixedEvaluation struct {
	// The evaluation components (for fallback)
	MaterialEvaluation  *MaterialEvaluation
	MobilityEvaluation  *MobilityEvaluation
	CornersEvaluation   *CornersEvaluation
	ParityEvaluation    *ParityEvaluation
	StabilityEvaluation *StabilityEvaluation
	FrontierEvaluation  *FrontierEvaluation

	// Coefficients for different game phases
	Coeffs EvaluationCoefficients

	// Batch evaluation management
	batchSize      int
	batchMutex     sync.Mutex
	boardBatch     []game.Board
	playerBatch    []game.Piece
	positionHashes map[string]int      // Position hash -> batch index
	resultCache    map[string]int      // Cache to avoid redundant transfers
	pendingEvals   map[string]chan int // Position hash -> result channel
	batchIsRunning bool

	// Stats
	hitCount  int
	missCount int
}

// NewGPUMixedEvaluation creates a new GPU-accelerated mixed evaluation
func NewGPUMixedEvaluation(coeffs EvaluationCoefficients) *GPUMixedEvaluation {
	// Initialize CUDA
	gpuAvailable := InitCUDA()

	// Set coefficients if GPU is available
	if gpuAvailable {
		SetCUDACoefficients(coeffs)
	}

	return &GPUMixedEvaluation{
		MaterialEvaluation:  NewMaterialEvaluation(),
		MobilityEvaluation:  NewMobilityEvaluation(),
		CornersEvaluation:   NewCornersEvaluation(),
		ParityEvaluation:    NewParityEvaluation(),
		StabilityEvaluation: NewStabilityEvaluation(),
		FrontierEvaluation:  NewFrontierEvaluation(),
		Coeffs:              coeffs,
		batchSize:           8192, // Large batch size for modern GPUs
		positionHashes:      make(map[string]int),
		resultCache:         make(map[string]int),
		pendingEvals:        make(map[string]chan int),
		batchIsRunning:      false,
		hitCount:            0,
		missCount:           0,
	}
}

// SetBatchSize sets the batch size for evaluation
func (e *GPUMixedEvaluation) SetBatchSize(size int) {
	// For larger GPU memory (>8GB), allow larger batch sizes
	if size < 1 {
		size = 1
	} else if size > 65536 {
		size = 65536 // Upper limit for safety
	}
	e.batchSize = size
}

// GetBatchSize returns the current batch size
func (e *GPUMixedEvaluation) GetBatchSize() int {
	return e.batchSize
}

// GetCacheStats returns cache hit/miss statistics
func (e *GPUMixedEvaluation) GetCacheStats() (hits, misses int) {
	e.batchMutex.Lock()
	defer e.batchMutex.Unlock()
	return e.hitCount, e.missCount
}

// Evaluate the given board state using GPU if available, fallback to CPU if not
func (e *GPUMixedEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	if !IsGPUAvailable() {
		// Fallback to CPU evaluation if GPU is not available
		return e.evaluateCPU(g, b, player)
	}

	// Create a cache key for this position
	boardKey := fmt.Sprintf("%v-%d", b, player.Color)

	// Check the cache to avoid recalculations
	e.batchMutex.Lock()
	if score, found := e.resultCache[boardKey]; found {
		e.hitCount++
		e.batchMutex.Unlock()
		return score
	}
	e.missCount++
	e.batchMutex.Unlock()

	// Get the global evaluation cache
	boardCache := cache.GetGlobalCache()

	// Check if the position is already cached
	score, _, _, found := boardCache.Lookup(b, player.Color, 0)
	if found {
		// Cache hit! Store in local cache too
		e.batchMutex.Lock()
		e.resultCache[boardKey] = score
		e.batchMutex.Unlock()
		return score
	}

	// Add to batch for GPU evaluation
	shouldFlush, _ := boardCache.AddToBatch(b, player.Color)

	// If batch processing was triggered, check cache again after processing
	if shouldFlush {
		// Force flush
		FlushBatch()

		// Check cache again
		score, _, _, found = boardCache.Lookup(b, player.Color, 0)
		if found {
			// Cache hit after flush! Store in local cache too
			e.batchMutex.Lock()
			e.resultCache[boardKey] = score
			e.batchMutex.Unlock()
			return score
		}
	}

	// Add to batch through the evaluation system
	e.batchMutex.Lock()
	// Create a result channel if this is a new evaluation
	resultChan, exists := e.pendingEvals[boardKey]
	if !exists {
		resultChan = make(chan int, 1)
		e.pendingEvals[boardKey] = resultChan

		// Add to batch
		idx := len(e.boardBatch)
		e.boardBatch = append(e.boardBatch, b)
		e.playerBatch = append(e.playerBatch, player.Color)
		e.positionHashes[boardKey] = idx

		// Process batch if full
		if len(e.boardBatch) >= e.batchSize && !e.batchIsRunning {
			e.processBatchAsync()
		}
	}
	e.batchMutex.Unlock()

	// If we still don't have a result, use CPU as a fallback
	// This is more responsive than waiting for GPU
	score = e.evaluateCPU(g, b, player)

	// Cache the CPU result
	e.batchMutex.Lock()
	e.resultCache[boardKey] = score
	e.batchMutex.Unlock()

	// Store in global cache too
	boardCache.Store(b, player.Color, 0, score, game.Position{Row: -1, Col: -1}, cache.SourceCPU)

	return score
}

// processBatchAsync processes a batch asynchronously
func (e *GPUMixedEvaluation) processBatchAsync() {
	if e.batchIsRunning || len(e.boardBatch) == 0 {
		return
	}

	e.batchIsRunning = true

	// Clone the current batch
	boards := make([]game.Board, len(e.boardBatch))
	players := make([]game.Piece, len(e.playerBatch))
	hashes := make(map[string]int)
	resultChans := make(map[string]chan int)

	copy(boards, e.boardBatch)
	copy(players, e.playerBatch)

	// Copy maps to avoid race conditions
	for k, v := range e.positionHashes {
		hashes[k] = v
	}
	for k, v := range e.pendingEvals {
		resultChans[k] = v
	}

	// Clear batch
	e.boardBatch = nil
	e.playerBatch = nil
	e.positionHashes = make(map[string]int)

	// Launch asynchronous processing
	go func() {
		scores := EvaluateStatesCUDA(boards, players)

		e.batchMutex.Lock()

		// Process results
		if len(scores) == len(boards) {
			for hash, idx := range hashes {
				if idx < len(scores) {
					// Cache the result
					e.resultCache[hash] = scores[idx]

					// Signal waiting goroutines
					if ch, exists := resultChans[hash]; exists {
						// Non-blocking send
						select {
						case ch <- scores[idx]:
						default:
						}

						// Remove from pending
						delete(e.pendingEvals, hash)
					}
				}
			}
		}

		e.batchIsRunning = false

		// If we have more items in batch, process them
		if len(e.boardBatch) > 0 {
			e.processBatchAsync()
		}

		e.batchMutex.Unlock()
	}()
}

// evaluateBatch evaluates all accumulated states in a single GPU call
func (e *GPUMixedEvaluation) evaluateBatch() {
	e.batchMutex.Lock()

	if len(e.boardBatch) == 0 {
		e.batchMutex.Unlock()
		return
	}

	// Clone the current states for evaluation
	boards := make([]game.Board, len(e.boardBatch))
	players := make([]game.Piece, len(e.playerBatch))
	hashes := make(map[string]int)

	copy(boards, e.boardBatch)
	copy(players, e.playerBatch)

	// Copy position hashes
	for k, v := range e.positionHashes {
		hashes[k] = v
	}

	// Reset the batches
	e.boardBatch = nil
	e.playerBatch = nil
	e.positionHashes = make(map[string]int)

	e.batchMutex.Unlock()

	// Evaluate on GPU
	scores := EvaluateStatesCUDA(boards, players)

	// Store the results in the cache
	e.batchMutex.Lock()
	if len(scores) == len(boards) {
		for hash, idx := range hashes {
			if idx < len(scores) {
				e.resultCache[hash] = scores[idx]
			}
		}
	} else {
		// GPU evaluation failed, evaluate individually on CPU
		for i, board := range boards {
			hash := fmt.Sprintf("%v-%d", board, players[i])
			e.resultCache[hash] = e.evaluateCPU(game.Game{}, board, game.Player{Color: players[i]})
		}
	}
	e.batchMutex.Unlock()
}

// evaluateCPU is a fallback CPU implementation
func (e *GPUMixedEvaluation) evaluateCPU(g game.Game, b game.Board, player game.Player) int {
	materialCoeff, mobilityCoeff, cornersCoeff, parityCoeff, stabilityCoeff, frontierCoeff := e.computeCoefficients(b)

	materialScore := e.MaterialEvaluation.Evaluate(g, b, player)
	mobilityScore := e.MobilityEvaluation.Evaluate(g, b, player)
	cornersScore := e.CornersEvaluation.Evaluate(g, b, player)
	parityScore := e.ParityEvaluation.Evaluate(g, b, player)
	stabilityScore := e.StabilityEvaluation.Evaluate(g, b, player)
	frontierScore := e.FrontierEvaluation.Evaluate(g, b, player)

	return materialCoeff*materialScore +
		mobilityCoeff*mobilityScore +
		cornersCoeff*cornersScore +
		parityCoeff*parityScore +
		stabilityCoeff*stabilityScore +
		frontierCoeff*frontierScore
}

// computeCoefficients returns the coefficients based on game phase
func (e *GPUMixedEvaluation) computeCoefficients(board game.Board) (int, int, int, int, int, int) {
	pieceCount := 0
	for _, row := range board {
		for _, piece := range row {
			if piece != game.Empty {
				pieceCount++
			}
		}
	}

	var phase int
	if pieceCount < 20 {
		phase = 0 // Early game
	} else if pieceCount <= 58 {
		phase = 1 // Mid game
	} else {
		phase = 2 // Late game
	}

	return e.Coeffs.MaterialCoeffs[phase],
		e.Coeffs.MobilityCoeffs[phase],
		e.Coeffs.CornersCoeffs[phase],
		e.Coeffs.ParityCoeffs[phase],
		e.Coeffs.StabilityCoeffs[phase],
		e.Coeffs.FrontierCoeffs[phase]
}
