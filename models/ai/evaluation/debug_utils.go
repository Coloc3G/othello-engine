package evaluation

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// DebugGPUCPUConsistency runs test comparisons between GPU and CPU evaluations
func DebugGPUCPUConsistency() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	fmt.Println("\n=== Testing GPU-CPU consistency ===")

	// Create test games from openings
	testGames := createTestGames(10)
	totalMismatches := 0
	totalTests := 0

	for i, g := range testGames {
		fmt.Printf("\nTest game #%d:\n", i+1)

		// Print board state
		printBoard(g.Board)

		// Test evaluation for both players
		totalTests += 2

		// Set up evaluators with same coefficients
		gpuEval := NewGPUMixedEvaluation(V1Coeff)
		cpuEval := NewMixedEvaluationWithCoefficients(V1Coeff)

		// Black player evaluation
		gpuBlackScore := gpuEval.Evaluate(g, g.Board, game.Player{Color: game.Black})
		cpuBlackScore := cpuEval.Evaluate(g, g.Board, game.Player{Color: game.Black})
		blackDiff := abs(gpuBlackScore - cpuBlackScore)

		if blackDiff > 0 {
			fmt.Printf("BLACK evaluation mismatch: GPU=%d, CPU=%d (diff=%d)\n",
				gpuBlackScore, cpuBlackScore, blackDiff)
			totalMismatches++
		} else {
			fmt.Printf("BLACK evaluation match: %d\n", gpuBlackScore)
		}

		// White player evaluation
		gpuWhiteScore := gpuEval.Evaluate(g, g.Board, game.Player{Color: game.White})
		cpuWhiteScore := cpuEval.Evaluate(g, g.Board, game.Player{Color: game.White})
		whiteDiff := abs(gpuWhiteScore - cpuWhiteScore)

		if whiteDiff > 0 {
			fmt.Printf("WHITE evaluation mismatch: GPU=%d, CPU=%d (diff=%d)\n",
				gpuWhiteScore, cpuWhiteScore, whiteDiff)
			totalMismatches++
		} else {
			fmt.Printf("WHITE evaluation match: %d\n", gpuWhiteScore)
		}

		// Test move finding at different depths
		for depth := 1; depth <= 3; depth++ {
			totalTests += 2
			fmt.Printf("\nTesting minimax at depth %d:\n", depth)

			// Black player
			gpuBlackPos, _ := GPUSolve(g, game.Player{Color: game.Black}, depth)
			cpuBlackPos := Solve(g, game.Player{Color: game.Black}, depth, cpuEval)

			if gpuBlackPos.Row != cpuBlackPos.Row || gpuBlackPos.Col != cpuBlackPos.Col {
				fmt.Printf("BLACK move mismatch at depth %d: GPU=%v, CPU=%v\n",
					depth, gpuBlackPos, cpuBlackPos)
				totalMismatches++
			} else {
				fmt.Printf("BLACK move match at depth %d: %v\n", depth, gpuBlackPos)
			}

			// White player
			gpuWhitePos, _ := GPUSolve(g, game.Player{Color: game.White}, depth)
			cpuWhitePos := Solve(g, game.Player{Color: game.White}, depth, cpuEval)

			if gpuWhitePos.Row != cpuWhitePos.Row || gpuWhitePos.Col != cpuWhitePos.Col {
				fmt.Printf("WHITE move mismatch at depth %d: GPU=%v, CPU=%v\n",
					depth, gpuWhitePos, cpuWhitePos)
				totalMismatches++
			} else {
				fmt.Printf("WHITE move match at depth %d: %v\n", depth, gpuWhitePos)
			}
		}
	}

	// Summary
	matchPct := 100.0 * float64(totalTests-totalMismatches) / float64(totalTests)
	fmt.Printf("\n=== Results: %d/%d tests matched (%.1f%% consistency) ===\n",
		totalTests-totalMismatches, totalTests, matchPct)
}

// createTestGames creates test games from openings
func createTestGames(count int) []game.Game {
	games := make([]game.Game, 0, count)

	// Use some known openings
	for i, op := range opening.KNOWN_OPENINGS {
		if i >= count {
			break
		}

		// Create new game
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

		// Add some random legal moves to create more variation
		randomMoves := rand.Intn(5) // 0-4 random moves
		for k := 0; k < randomMoves; k++ {
			moves := game.ValidMoves(g.Board, g.CurrentPlayer.Color)
			if len(moves) == 0 {
				break
			}

			// Pick a random move
			move := moves[rand.Intn(len(moves))]
			g.Board, _ = game.ApplyMoveToBoard(g.Board, g.CurrentPlayer.Color, move)
			g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
		}

		games = append(games, *g)
	}

	return games
}

// printBoard prints a board in a readable format
func printBoard(b game.Board) {
	fmt.Println("  0 1 2 3 4 5 6 7")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d ", i)
		for j := 0; j < 8; j++ {
			switch b[i][j] {
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

// TestDeterminism runs a thorough test of GPU/CPU determinism
func TestDeterminism() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	fmt.Println("\n=== Running Determinism Tests ===")

	// Test case: early game position from error logs
	testBoard := game.NewGame("Test", "Test").Board
	// Set up the problematic board position
	testBoard[2][2] = game.White
	testBoard[3][2] = game.Black
	testBoard[3][3] = game.White
	testBoard[3][4] = game.Black
	testBoard[4][3] = game.Black
	testBoard[4][4] = game.White

	// Create game from this board
	g := game.NewGame("Test", "Test")
	g.Board = testBoard
	player := game.Player{Color: game.Black}

	// Print the board
	fmt.Println("Testing problematic board:")
	printBoard(testBoard)

	// Set up evaluators
	cpuEval := NewMixedEvaluationWithCoefficients(V1Coeff)
	// Set coefficients in CUDA
	SetCUDACoefficients(V1Coeff)

	// Get valid moves
	validMoves := game.ValidMoves(testBoard, player.Color)
	fmt.Printf("Valid moves: %v\n", validMoves)

	// Sort moves for deterministic ordering
	sort.Slice(validMoves, func(i, j int) bool {
		if validMoves[i].Row == validMoves[j].Row {
			return validMoves[i].Col < validMoves[j].Col
		}
		return validMoves[i].Row < validMoves[j].Row
	})

	// Test each move evaluation
	fmt.Println("\nTesting each move evaluation:")
	for _, move := range validMoves {
		// Apply move
		newBoard, _ := game.ApplyMoveToBoard(testBoard, player.Color, move)

		// Get CPU and GPU evaluation
		cpuScore := cpuEval.Evaluate(*g, newBoard, player)

		// Convert board to array for GPU
		flatBoard := make([]int, 64)
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				flatBoard[i*8+j] = int(newBoard[i][j])
			}
		}

		// Get GPU score
		gpuScores := EvaluateStatesCUDA([]game.Board{newBoard}, []game.Piece{player.Color})
		var gpuScore int
		if len(gpuScores) > 0 {
			gpuScore = gpuScores[0]
		}

		fmt.Printf("Move %v: CPU=%d, GPU=%d, diff=%d\n",
			move, cpuScore, gpuScore, cpuScore-gpuScore)
	}

	// Test minimax at each depth
	for depth := 1; depth <= 5; depth++ {
		fmt.Printf("\nTesting minimax at depth %d:\n", depth)

		// CPU solve
		cpuStart := time.Now()
		cpuPos := Solve(*g, player, depth, cpuEval)
		cpuTime := time.Since(cpuStart)

		// GPU solve
		gpuStart := time.Now()
		gpuPos, _ := GPUSolve(*g, player, depth)
		gpuTime := time.Since(gpuStart)

		if cpuPos.Row != gpuPos.Row || cpuPos.Col != gpuPos.Col {
			fmt.Printf("FAIL: Minimax mismatch at depth %d - CPU: %v, GPU: %v\n",
				depth, cpuPos, gpuPos)

			// Apply both moves
			cpuBoard, _ := game.ApplyMoveToBoard(testBoard, player.Color, cpuPos)
			gpuBoard, _ := game.ApplyMoveToBoard(testBoard, player.Color, gpuPos)

			// Evaluate results
			cpuScore := cpuEval.Evaluate(*g, cpuBoard, player)
			gpuScores := EvaluateStatesCUDA([]game.Board{gpuBoard}, []game.Piece{player.Color})
			var gpuScore int
			if len(gpuScores) > 0 {
				gpuScore = gpuScores[0]
			}

			fmt.Printf("CPU move %v score: %d, GPU move %v score: %d\n",
				cpuPos, cpuScore, gpuPos, gpuScore)
		} else {
			fmt.Printf("PASS: Minimax match at depth %d: %v (CPU: %v, GPU: %v)\n",
				depth, cpuPos, cpuTime, gpuTime)
		}
	}
}

// Additional test function to diagnose complex positions
func TestComplexPosition() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	// Load a complex mid-game position
	testBoard := game.NewGame("Test", "Test").Board

	// Setup a more complex position with many pieces
	testBoard[0][0] = game.White
	testBoard[0][1] = game.White
	testBoard[0][2] = game.White
	testBoard[0][3] = game.White
	testBoard[0][4] = game.White
	testBoard[1][0] = game.Black
	testBoard[1][1] = game.White
	testBoard[1][2] = game.Black
	testBoard[1][3] = game.Black
	testBoard[1][4] = game.Black
	testBoard[2][0] = game.White
	testBoard[2][1] = game.White
	testBoard[2][2] = game.White
	testBoard[2][3] = game.Black
	testBoard[3][0] = game.Black
	testBoard[3][1] = game.White
	testBoard[3][2] = game.Black
	testBoard[3][3] = game.Black
	testBoard[3][4] = game.Black
	testBoard[4][0] = game.Black
	testBoard[4][1] = game.Black
	testBoard[4][2] = game.White
	testBoard[4][3] = game.Black
	testBoard[4][4] = game.White

	// Create game from this board
	g := game.NewGame("Test", "Test")
	g.Board = testBoard
	player := game.Player{Color: game.Black}

	fmt.Println("\n=== Testing Complex Position ===")
	printBoard(testBoard)

	// Set up evaluators
	cpuEval := NewMixedEvaluationWithCoefficients(V1Coeff)
	SetCUDACoefficients(V1Coeff)

	// Test minimax
	for depth := 1; depth <= 4; depth++ {
		fmt.Printf("Testing minimax at depth %d:\n", depth)

		// CPU solve
		cpuStart := time.Now()
		cpuPos := Solve(*g, player, depth, cpuEval)
		cpuTime := time.Since(cpuStart)

		// GPU solve
		gpuStart := time.Now()
		gpuPos, _ := GPUSolve(*g, player, depth)
		gpuTime := time.Since(gpuStart)

		if cpuPos.Row != gpuPos.Row || cpuPos.Col != gpuPos.Col {
			fmt.Printf("FAIL: Mismatch at depth %d - CPU: %v, GPU: %v\n",
				depth, cpuPos, gpuPos)
		} else {
			fmt.Printf("PASS: Match at depth %d: %v (CPU: %v, GPU: %v)\n",
				depth, cpuPos, cpuTime, gpuTime)
		}
	}
}

// RunGPUDeterminismTests runs all determinism tests
func RunGPUDeterminismTests() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	TestDeterminism()
	TestComplexPosition()
	DebugGPUCPUConsistency()
}
