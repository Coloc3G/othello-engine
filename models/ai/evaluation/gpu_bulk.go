package evaluation

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include "othello_cuda.h"
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/Coloc3G/othello-engine/models/game"
)

var (
	bulkMutex sync.Mutex
)

// BulkEvaluationResult represents results from a bulk GPU evaluation
type BulkEvaluationResult struct {
	Scores       []int
	BestMoves    []game.Position
	ProcessedAt  time.Time
	ProcessingMs int64
}

// EvaluateAndFindBestMovesBulk performs bulk evaluation and move finding for multiple positions
// This is optimized for training, where we need to evaluate many positions at once
func EvaluateAndFindBestMovesBulk(boards []game.Board, players []game.Piece, depths []int) *BulkEvaluationResult {
	if !IsGPUAvailable() {
		return nil
	}

	numStates := len(boards)
	if numStates == 0 || numStates != len(players) || numStates != len(depths) {
		return nil
	}

	// Ensure exclusive access to GPU for bulk operations
	bulkMutex.Lock()
	defer bulkMutex.Unlock()

	startTime := time.Now()

	// Flatten boards for C
	flattenedBoards := make([]int, numStates*8*8)
	for s := 0; s < numStates; s++ {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				flattenedBoards[s*64+i*8+j] = int(boards[s][i][j])
			}
		}
	}

	// Convert player colors and depths to int arrays
	colorInts := make([]int, numStates)
	depthInts := make([]int, numStates)
	for i := 0; i < numStates; i++ {
		colorInts[i] = int(players[i])
		depthInts[i] = depths[i]
	}

	// Prepare C arrays for input and output
	boardsC := (*C.int)(unsafe.Pointer(&flattenedBoards[0]))
	colorsC := (*C.int)(unsafe.Pointer(&colorInts[0]))
	depthsC := (*C.int)(unsafe.Pointer(&depthInts[0]))

	// Allocate memory for results
	scoresC := (*C.int)(C.malloc(C.size_t(numStates * 4)))   // 4 bytes per int
	bestRowsC := (*C.int)(C.malloc(C.size_t(numStates * 4))) // 4 bytes per int
	bestColsC := (*C.int)(C.malloc(C.size_t(numStates * 4))) // 4 bytes per int
	defer C.free(unsafe.Pointer(scoresC))
	defer C.free(unsafe.Pointer(bestRowsC))
	defer C.free(unsafe.Pointer(bestColsC))

	// Call C function to evaluate and find best moves in bulk
	C.evaluateAndFindBestMoves(
		boardsC, colorsC, depthsC,
		scoresC, bestRowsC, bestColsC,
		C.int(numStates))

	// Convert results back to Go types
	scores := make([]int, numStates)
	bestMoves := make([]game.Position, numStates)

	for i := 0; i < numStates; i++ {
		scores[i] = int(*(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(scoresC)) + uintptr(i*4))))

		row := int(*(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(bestRowsC)) + uintptr(i*4))))
		col := int(*(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(bestColsC)) + uintptr(i*4))))

		// Ensure move is valid
		if row >= 0 && row < 8 && col >= 0 && col < 8 {
			bestMoves[i] = game.Position{Row: row, Col: col}
		} else {
			bestMoves[i] = game.Position{Row: -1, Col: -1} // Invalid move
		}
	}

	processingTime := time.Since(startTime)

	// Print some stats if processing a large batch
	if numStates > 1000 {
		fmt.Printf("GPU processed %d board evaluations in %s (%.2f boards/ms)\n",
			numStates, processingTime, float64(numStates)/float64(processingTime.Milliseconds()))
	}

	return &BulkEvaluationResult{
		Scores:       scores,
		BestMoves:    bestMoves,
		ProcessedAt:  time.Now(),
		ProcessingMs: processingTime.Milliseconds(),
	}
}

// ProcessBatchedMinimax processes a large batch of minimax operations in one GPU call
// This is more efficient than individual calls when training
func ProcessBatchedMinimax(positions []game.Board, players []game.Piece, searchDepth int) ([]game.Position, []int) {
	if !IsGPUAvailable() || len(positions) == 0 {
		return nil, nil
	}

	// Create depths array (all same depth)
	depths := make([]int, len(positions))
	for i := range depths {
		depths[i] = searchDepth
	}

	// Call bulk evaluation
	result := EvaluateAndFindBestMovesBulk(positions, players, depths)
	if result == nil {
		return nil, nil
	}

	return result.BestMoves, result.Scores
}
