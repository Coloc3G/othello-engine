package evaluation

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include "othello_cuda.h"
import "C"
import (
	"fmt"
	"math/rand"
	"sort"
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

	// Skip if GPU is not available or evaluation is not initialized
	if !IsGPUAvailable() {
		perfStats.RecordOperation("gpu_unavailable", 0)
		return game.Position{Row: -1, Col: -1}, false
	}

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

func GPUSolveCmpCPU(g game.Game, player game.Player, depth int, coefs EvaluationCoefficients) (game.Position, bool) {
	// Create a performance stats object for detailed tracking
	perfStats := stats.NewPerformanceStats()

	// Sort all valid moves for deterministic results
	validMoves := game.ValidMoves(g.Board, player.Color)
	if len(validMoves) > 0 {
		sort.Slice(validMoves, func(i, j int) bool {
			if validMoves[i].Row == validMoves[j].Row {
				return validMoves[i].Col < validMoves[j].Col
			}
			return validMoves[i].Row < validMoves[j].Row
		})
	}

	// Try GPU solve first
	gpuPos, success := GPUSolveWithStats(g, player, depth, perfStats)

	// Occasionally verify with CPU solve (1% chance for performance, 100% for debugging)
	verifyRate := 0.01 // 1% verification rate for normal operation
	// verifyRate := 1.0  // 100% for debugging - uncomment to always verify

	if success && rand.Float64() < verifyRate {
		// Create a CPU evaluator with the same coefficients
		var cpuEval Evaluation

		// Try to extract coefficients from current CUDA state
		if IsGPUAvailable() {
			// Use exact same coefficients for CPU
			cpuEval = NewMixedEvaluationWithCoefficients(coefs)

			// CPU minimax with same depth
			startCPU := time.Now()
			cpuPos := Solve(g, player, depth, cpuEval)
			cpuTime := time.Since(startCPU)

			// Log potential discrepancies
			if cpuPos.Row != gpuPos.Row || cpuPos.Col != gpuPos.Col {
				// Check both moves' validity
				gpuValid := game.IsValidMove(g.Board, player.Color, gpuPos)
				cpuValid := game.IsValidMove(g.Board, player.Color, cpuPos)

				// Get the evaluation of both positions to see if they have equal scores
				equalScores := false
				equalEvaluation := false
				if gpuValid && cpuValid {
					// Apply both moves and compare their evaluations
					gpuBoard, _ := game.ApplyMoveToBoard(g.Board, player.Color, gpuPos)
					cpuBoard, _ := game.ApplyMoveToBoard(g.Board, player.Color, cpuPos)

					// Get evaluation scores using CPU evaluation
					gpuScore := cpuEval.Evaluate(g, gpuBoard, player)
					cpuScore := cpuEval.Evaluate(g, cpuBoard, player)

					// If scores are exactly equal, consider them equal
					if gpuScore == cpuScore {
						equalScores = true
					}

					// Also directly check if they produce the same pure evaluation
					directGpuScore := cpuEval.Evaluate(g, g.Board, player)
					directCpuScore := cpuEval.Evaluate(g, g.Board, player)
					if directGpuScore == directCpuScore {
						equalEvaluation = true
					}

					if !equalScores || !equalEvaluation {
						// Print board state for easier debugging
						fmt.Printf("Board state causing mismatch:\n")
						printBoardForDebugging(g.Board)

						// Print all available moves and their sorted order
						fmt.Printf("Valid moves (sorted): %v\n", validMoves)

						// Print comparison of scores
						fmt.Printf("GPU move %v (score: %d) vs CPU move %v (score: %d), diff=%d\n",
							gpuPos, gpuScore, cpuPos, cpuScore, gpuScore-cpuScore)

						// Print direct evaluation comparison
						if !equalEvaluation {
							fmt.Printf("Direct evaluation comparison - GPU: %d, CPU: %d, diff=%d\n",
								directGpuScore, directCpuScore, directGpuScore-directCpuScore)
						}

						// Additional debug: trace minimax by printing search tree
						fmt.Printf("CPU Minimax decision path (depth %d):\n", depth)
						traceMinimax(g, player, depth, cpuEval)
					}
				}

				if !gpuValid {
					fmt.Printf("WARNING: GPU move is invalid!\n")
				} else if !cpuValid {
					fmt.Printf("WARNING: CPU move is invalid!\n")
				} else if equalScores {
					fmt.Printf("VERIFICATION: GPU/CPU move mismatch at depth %d - GPU: %v, CPU: %v (took %v)\n",
						depth, gpuPos, cpuPos, cpuTime)
					fmt.Printf("Both moves are valid with equal evaluation scores\n")
				} else {
					fmt.Printf("VERIFICATION: GPU/CPU move mismatch at depth %d - GPU: %v, CPU: %v (took %v)\n",
						depth, gpuPos, cpuPos, cpuTime)
					fmt.Printf("WARNING: Moves have different evaluations - potential algorithm inconsistency\n")
				}
			} else {
				fmt.Printf("VERIFICATION: GPU/CPU match confirmed at depth %d (CPU took %v)\n",
					depth, cpuTime)
			}
		}
	}

	return gpuPos, success
}

// Helper function to print a board state for debugging
func printBoardForDebugging(board game.Board) {
	fmt.Println("  0 1 2 3 4 5 6 7")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d ", i)
		for j := 0; j < 8; j++ {
			switch board[i][j] {
			case game.Black:
				fmt.Print("B ")
			case game.White:
				fmt.Print("W ")
			default:
				fmt.Print(". ")
			}
		}
		fmt.Println()
	}
}

// Helper to trace minimax decision tree (for debugging)
func traceMinimax(g game.Game, player game.Player, depth int, eval Evaluation) {
	validMoves := game.ValidMoves(g.Board, player.Color)
	if len(validMoves) == 0 {
		fmt.Println("  No valid moves")
		return
	}

	// Sort moves for consistent output
	sort.Slice(validMoves, func(i, j int) bool {
		if validMoves[i].Row == validMoves[j].Row {
			return validMoves[i].Col < validMoves[j].Col
		}
		return validMoves[i].Row < validMoves[j].Row
	})

	fmt.Println("  Available moves:")
	for _, move := range validMoves {
		// Apply the move
		newBoard, _ := game.ApplyMoveToBoard(g.Board, player.Color, move)

		// Evaluate the resulting position
		score := eval.Evaluate(g, newBoard, player)

		fmt.Printf("    Move: %v - Direct eval score: %d\n", move, score)
	}
}
