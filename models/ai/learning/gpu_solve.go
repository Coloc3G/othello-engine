package learning

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include "othello_cuda.h"
import "C"
import (
	"sync"
	"time"
	"unsafe"

	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
)

var (
	gpuSolveMutex sync.Mutex
)

// GPUSolveWithStats performs minimax search with GPU acceleration and collects performance statistics
func GPUSolveWithStats(g game.Game, player game.Player, depth int, perfStats *stats.PerformanceStats) (game.Position, bool) {
	totalStart := time.Now()

	// Record cache miss
	perfStats.RecordOperation("cache_miss", 0)

	// Only attempt GPU solve if there are valid moves
	validMoves := game.ValidMoves(g.Board, player.Color)
	if len(validMoves) == 0 {
		perfStats.RecordOperation("gpu_no_moves", 0)
		return game.Position{Row: -1, Col: -1}, false
	}

	// If only one move is available, return it immediately
	if len(validMoves) == 1 {
		perfStats.RecordOperation("gpu_single_move", 0)
		return validMoves[0], true
	}

	// Ensure exclusive access to GPU for this search
	gpuAccessStart := time.Now()
	gpuSolveMutex.Lock()
	defer gpuSolveMutex.Unlock()
	gpuAccessTime := time.Since(gpuAccessStart)
	perfStats.RecordOperation("gpu_access_wait", gpuAccessTime)

	// Convert the board to a flat array for C
	flatBoard := make([]C.int, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			flatBoard[i*8+j] = C.int(g.Board[i][j])
		}
	}

	// Prepare C variables
	boardC := (*C.int)(unsafe.Pointer(&flatBoard[0]))
	playerC := C.int(int(player.Color))
	depthC := C.int(depth)

	var bestRow, bestCol C.int
	bestRowPtr := (*C.int)(unsafe.Pointer(&bestRow))
	bestColPtr := (*C.int)(unsafe.Pointer(&bestCol))

	// Call C function to find best move with timing
	gpuComputeStart := time.Now()
	result := C.findBestMove(boardC, playerC, depthC, bestRowPtr, bestColPtr)
	gpuComputeTime := time.Since(gpuComputeStart)
	perfStats.RecordOperation("gpu_compute", gpuComputeTime)

	// Check if a valid move was found
	if result <= -1000000 || bestRow < 0 || bestRow >= 8 || bestCol < 0 || bestCol >= 8 {
		perfStats.RecordOperation("gpu_failure", 0)
		return game.Position{Row: -1, Col: -1}, false
	}

	// Verify that the move is valid
	foundMove := game.Position{Row: int(bestRow), Col: int(bestCol)}
	if !game.IsValidMove(g.Board, player.Color, foundMove) {
		perfStats.RecordOperation("gpu_invalid_move", 0)
		return game.Position{Row: -1, Col: -1}, false
	}

	// Record total time and success
	totalTime := time.Since(totalStart)
	perfStats.RecordOperation("gpu_solve_success", totalTime)
	perfStats.GPUSuccesses++

	// Return the best move
	return foundMove, true
}

// GPUSolve performs the minimax search using GPU acceleration
// It returns the best move and a boolean indicating if the operation was successful
func GPUSolve(g game.Game, player game.Player, depth int) (game.Position, bool) {
	// Create a dummy stats object if needed for detailed tracking
	perfStats := stats.NewPerformanceStats()
	return GPUSolveWithStats(g, player, depth, perfStats)
}
