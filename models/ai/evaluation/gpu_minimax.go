package evaluation

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello -L${SRCDIR}/. -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include <string.h>
// #include "othello_cuda.h"
import "C"
import (
	"sync"
	"time"
	"unsafe"

	"github.com/Coloc3G/othello-engine/models/ai/cache"
	"github.com/Coloc3G/othello-engine/models/game"
)

// PerformanceTracker is a minimal internal tracker for GPU performance
type PerformanceTracker struct {
	SuccessCount   int
	FailureCount   int
	TransferTime   time.Duration
	ComputeTime    time.Duration
	BatchesCount   int
	PositionsCount int
}

var (
	gpuPerfTracker PerformanceTracker
	perfMutex      sync.Mutex
)

// Maximum batch size for GPU minimax operations
const MaxMinimaxBatchSize = 64

var (
	gpuMinimaxBatchMutex   sync.Mutex
	gpuMinimaxBatchBoards  []game.Board
	gpuMinimaxBatchPlayers []game.Piece
	gpuMinimaxBatchDepths  []int
	gpuMinimaxBatchResults []gpuMinimaxResult
	gpuMinimaxIsRunning    bool
)

type gpuMinimaxResult struct {
	score     int
	bestMove  game.Position
	completed bool
}

// GetGPUPerformanceStats returns current GPU performance statistics
func GetGPUPerformanceStats() (successes, failures int, avgTransferMs, avgComputeMs float64) {
	perfMutex.Lock()
	defer perfMutex.Unlock()

	if gpuPerfTracker.SuccessCount+gpuPerfTracker.FailureCount == 0 {
		return 0, 0, 0, 0
	}

	totalOps := gpuPerfTracker.SuccessCount + gpuPerfTracker.FailureCount
	avgTransfer := float64(gpuPerfTracker.TransferTime.Milliseconds()) / float64(totalOps)
	avgCompute := float64(gpuPerfTracker.ComputeTime.Milliseconds()) / float64(totalOps)

	return gpuPerfTracker.SuccessCount, gpuPerfTracker.FailureCount, avgTransfer, avgCompute
}

// GPUMinimaxSolve tries to find the best move using GPU-accelerated minimax
// It returns the best move, score, and whether the operation was successful
func GPUMinimaxSolve(g game.Game, player game.Player, depth int) (game.Position, int, bool) {
	if !IsGPUAvailable() {
		recordPerformance(false, 0, 0)
		return game.Position{}, 0, false
	}

	// Get the evaluation cache
	boardCache := cache.GetGlobalCache()

	// Check if the position is already cached
	score, bestMove, _, found := boardCache.Lookup(g.Board, player.Color, depth)
	if found && bestMove.Row >= 0 && bestMove.Col < 8 {
		// Valid cache hit - count as success
		return bestMove, score, true
	}

	// First try direct GPU solve with FindBestMoveCUDA
	transferStart := time.Now()
	bestMove, success := FindBestMoveCUDA(g.Board, player.Color, depth)
	transferTime := time.Since(transferStart)

	if success {
		// Get score from CUDA evaluation
		computeStart := time.Now()
		scores := EvaluateStatesCUDA([]game.Board{g.Board}, []game.Piece{player.Color})
		computeTime := time.Since(computeStart)

		if len(scores) > 0 {
			score = scores[0]
			// Cache the result
			boardCache.Store(g.Board, player.Color, depth, score, bestMove, cache.SourceGPU)

			// Record performance
			recordPerformance(true, transferTime, computeTime)

			return bestMove, score, true
		}
	}

	// If direct solve failed, try batch processing
	return queueForBatchMinimax(g.Board, player.Color, depth)
}

// recordPerformance tracks GPU performance stats
func recordPerformance(success bool, transferTime, computeTime time.Duration) {
	perfMutex.Lock()
	defer perfMutex.Unlock()

	if success {
		gpuPerfTracker.SuccessCount++
	} else {
		gpuPerfTracker.FailureCount++
	}

	gpuPerfTracker.TransferTime += transferTime
	gpuPerfTracker.ComputeTime += computeTime
}

// queueForBatchMinimax adds a position to the GPU minimax batch
// and returns the result if batch processing is successful
func queueForBatchMinimax(board game.Board, player game.Piece, depth int) (game.Position, int, bool) {
	gpuMinimaxBatchMutex.Lock()

	// Add to batch
	index := len(gpuMinimaxBatchBoards)
	gpuMinimaxBatchBoards = append(gpuMinimaxBatchBoards, board)
	gpuMinimaxBatchPlayers = append(gpuMinimaxBatchPlayers, player)
	gpuMinimaxBatchDepths = append(gpuMinimaxBatchDepths, depth)

	// Initialize empty result
	gpuMinimaxBatchResults = append(gpuMinimaxBatchResults, gpuMinimaxResult{
		completed: false,
	})

	// Process batch if it's full and not already running
	processBatch := len(gpuMinimaxBatchBoards) >= MaxMinimaxBatchSize && !gpuMinimaxIsRunning

	if processBatch {
		gpuMinimaxIsRunning = true
	}

	gpuMinimaxBatchMutex.Unlock()

	// Process batch if needed
	if processBatch {
		processBatchMinimax()
	}

	// Check if our request was processed
	gpuMinimaxBatchMutex.Lock()
	result := gpuMinimaxBatchResults[index]
	success := result.completed
	gpuMinimaxBatchMutex.Unlock()

	if success {
		// If the batch processing gave us a result, use it
		return result.bestMove, result.score, true
	}

	// Not processed, return failure
	return game.Position{}, 0, false
}

// processBatchMinimax processes the current batch of minimax requests
func processBatchMinimax() {
	gpuMinimaxBatchMutex.Lock()

	// Get current batch
	boardsCopy := make([]game.Board, len(gpuMinimaxBatchBoards))
	playersCopy := make([]game.Piece, len(gpuMinimaxBatchPlayers))
	depthsCopy := make([]int, len(gpuMinimaxBatchDepths))

	copy(boardsCopy, gpuMinimaxBatchBoards)
	copy(playersCopy, gpuMinimaxBatchPlayers)
	copy(depthsCopy, gpuMinimaxBatchDepths)

	// Process each request
	for i := range boardsCopy {
		// Call FindBestMoveCUDA for each board in batch
		bestMove, success := FindBestMoveCUDA(boardsCopy[i], playersCopy[i], depthsCopy[i])

		if success {
			// Get score
			scores := EvaluateStatesCUDA([]game.Board{boardsCopy[i]}, []game.Piece{playersCopy[i]})
			if len(scores) > 0 {
				gpuMinimaxBatchResults[i].score = scores[0]
				gpuMinimaxBatchResults[i].bestMove = bestMove
				gpuMinimaxBatchResults[i].completed = true

				// Cache result
				boardCache := cache.GetGlobalCache()
				boardCache.Store(boardsCopy[i], playersCopy[i], depthsCopy[i],
					scores[0], bestMove, cache.SourceGPU)
			}
		}
	}

	// Update stats
	perfMutex.Lock()
	gpuPerfTracker.BatchesCount++
	gpuPerfTracker.PositionsCount += len(boardsCopy)
	perfMutex.Unlock()

	// Clear batch
	gpuMinimaxBatchBoards = nil
	gpuMinimaxBatchPlayers = nil
	gpuMinimaxBatchDepths = nil
	gpuMinimaxBatchResults = nil
	gpuMinimaxIsRunning = false

	gpuMinimaxBatchMutex.Unlock()
}

// GPUMinimax is a GPU-specific implementation of minimax
// It uses direct GPU call instead of the recursive CPU version
func GPUMinimax(board game.Board, player game.Piece, depth int) (int, game.Position) {
	if !IsGPUAvailable() {
		return 0, game.Position{Row: -1, Col: -1}
	}

	// Flatten the board for C
	flatBoard := make([]C.int, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			flatBoard[i*8+j] = C.int(board[i][j])
		}
	}

	// Prepare C variables
	boardC := (*C.int)(unsafe.Pointer(&flatBoard[0]))
	playerC := C.int(int(player))
	depthC := C.int(depth)

	var bestRow, bestCol C.int
	bestRowPtr := (*C.int)(unsafe.Pointer(&bestRow))
	bestColPtr := (*C.int)(unsafe.Pointer(&bestCol))

	// Call C function to find best move
	result := C.findBestMove(boardC, playerC, depthC, bestRowPtr, bestColPtr)

	// Convert back to Go
	return int(result), game.Position{Row: int(bestRow), Col: int(bestCol)}
}

// HasValidMovesCUDA checks if a player has valid moves directly on GPU
func HasValidMovesCUDA(board game.Board, player game.Piece) bool {
	if !IsGPUAvailable() {
		return false
	}

	// Flatten the board for C
	flatBoard := make([]C.int, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			flatBoard[i*8+j] = C.int(board[i][j])
		}
	}

	// Prepare C variables
	boardC := (*C.int)(unsafe.Pointer(&flatBoard[0]))
	playerC := C.int(int(player))

	// Call C function to check for valid moves
	result := C.hasValidMoves(boardC, playerC)

	return int(result) > 0
}

// IsGameFinishedCUDA checks if the game is finished on GPU
func IsGameFinishedCUDA(board game.Board) bool {
	if !IsGPUAvailable() {
		return false
	}

	// For now this is a simple wrapper:
	blackMoves := HasValidMovesCUDA(board, game.Black)
	whiteMoves := HasValidMovesCUDA(board, game.White)

	return !blackMoves && !whiteMoves
}
