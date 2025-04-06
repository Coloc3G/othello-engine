package evaluation

import (
	"fmt"
	"time"

	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// PerformanceResults contains benchmark results
type PerformanceResults struct {
	BatchThroughput   float64 // positions/ms
	SingleEvalAvgTime time.Duration
	TransferTimeRatio float64 // percent of total time
	KernelTimeRatio   float64 // percent of total time
	GPUMemoryFree     uint64
	GPUMemoryTotal    uint64
	SuccessRate       float64 // percent
}

// RunGPUBenchmark runs a comprehensive GPU performance benchmark
func RunGPUBenchmark() *PerformanceResults {
	if !IsGPUAvailable() {
		fmt.Println("GPU acceleration not available")
		return nil
	}

	fmt.Println("Running GPU Performance Benchmark")
	fmt.Println("=================================")

	results := &PerformanceResults{}

	// Run basic memory transfer benchmark
	benchmarkResults := runMemoryTransferBenchmark()
	results.BatchThroughput = benchmarkResults.BatchThroughput
	results.TransferTimeRatio = benchmarkResults.TransferTimeRatio
	results.KernelTimeRatio = benchmarkResults.KernelTimeRatio

	// Test search success rate
	successResults := runSearchBenchmark(10) // Use 10 positions for testing
	results.SuccessRate = successResults

	fmt.Printf("GPU batch throughput: %.2f positions/ms\n", results.BatchThroughput)
	fmt.Printf("Search success rate: %.1f%%\n", results.SuccessRate)

	return results
}

// BenchmarkResults contains results from a single benchmark
type BenchmarkResults struct {
	BatchThroughput   float64
	TransferTimeRatio float64
	KernelTimeRatio   float64
}

// runMemoryTransferBenchmark measures memory transfer performance
func runMemoryTransferBenchmark() BenchmarkResults {
	fmt.Println("\nMemory Transfer Benchmark")
	fmt.Println("========================")

	var results BenchmarkResults
	batchSizes := []int{64, 256, 1024, 4096}

	var bestThroughput float64
	for _, size := range batchSizes {
		// Generate test data
		boards := make([]game.Board, size)
		players := make([]game.Piece, size)

		// Initialize with some random data
		for i := range boards {
			boards[i] = generateRandomBoard(i)
			players[i] = game.Black
		}

		// Warm up
		EvaluateStatesCUDA(boards[:10], players[:10])

		// Reset stats
		ResetBatchStats()

		// Benchmark
		fmt.Printf("Batch size %d: ", size)
		start := time.Now()
		scores := EvaluateStatesCUDA(boards, players)
		duration := time.Since(start)

		throughput := float64(len(scores)) / float64(duration.Milliseconds())
		fmt.Printf("%d positions in %s (%.2f pos/ms)\n",
			len(scores), duration, throughput)

		// Get stats
		batches, _, avgTransferMs, avgKernelMs := GetBatchStats()
		if batches > 0 {
			totalTime := avgTransferMs + avgKernelMs
			transferRatio := avgTransferMs / totalTime * 100
			kernelRatio := avgKernelMs / totalTime * 100

			fmt.Printf("  Transfer: %.2f ms (%.1f%%), Kernel: %.2f ms (%.1f%%)\n",
				avgTransferMs, transferRatio, avgKernelMs, kernelRatio)

			// Keep best results
			if throughput > bestThroughput {
				bestThroughput = throughput
				results.BatchThroughput = throughput
				results.TransferTimeRatio = transferRatio
				results.KernelTimeRatio = kernelRatio
			}
		}
	}

	return results
}

// runSearchBenchmark tests the success rate of GPU minimax
func runSearchBenchmark(numPositions int) float64 {
	fmt.Println("\nGPU Search Benchmark")
	fmt.Println("===================")

	// Create test cases from openings
	testCases := make([]struct {
		board  game.Board
		player game.Piece
		depth  int
	}, 0, numPositions)

	// Use openings to generate test cases
	for i, op := range opening.KNOWN_OPENINGS {
		if i >= numPositions {
			break
		}

		// Create game from opening
		g := game.NewGame("Test", "Test")

		// Apply opening moves
		transcript := op.Transcript
		for j := 0; j < len(transcript); j += 2 {
			if j+1 >= len(transcript) {
				break
			}

			move := utils.AlgebraicToPosition(transcript[j : j+2])
			g.Board, _ = game.ApplyMoveToBoard(g.Board, g.CurrentPlayer.Color, move)
			g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
		}

		testCases = append(testCases, struct {
			board  game.Board
			player game.Piece
			depth  int
		}{
			board:  g.Board,
			player: g.CurrentPlayer.Color,
			depth:  5,
		})
	}

	// Run the benchmark
	fmt.Printf("Testing GPU search with %d positions...\n", len(testCases))

	successes := 0
	for i, tc := range testCases {
		start := time.Now()
		pos, success := FindBestMoveCUDA(tc.board, tc.player, tc.depth)
		duration := time.Since(start)

		// Validate the move is valid
		validMove := false
		if success {
			validMove = game.IsValidMove(tc.board, tc.player, pos)
			if validMove {
				successes++
			}
		}

		fmt.Printf("  Position %d: %v in %s (valid: %v)\n",
			i, success, duration, validMove)
	}

	successRate := float64(successes) / float64(len(testCases)) * 100
	fmt.Printf("Success rate: %.1f%% (%d/%d)\n",
		successRate, successes, len(testCases))

	return successRate
}

// generateRandomBoard creates a board with some random pieces
func generateRandomBoard(seed int) game.Board {
	var board game.Board

	// Initialize empty board
	for i := range board {
		for j := range board[i] {
			board[i][j] = game.Empty
		}
	}

	// Add some pieces based on seed
	piecesToAdd := 20 + (seed % 40) // Between 20-60 pieces

	for i := 0; i < piecesToAdd; i++ {
		row := (seed + i) % 8
		col := (seed + i*i) % 8

		if board[row][col] == game.Empty {
			if i%2 == 0 {
				board[row][col] = game.Black
			} else {
				board[row][col] = game.White
			}
		}
	}

	return board
}
