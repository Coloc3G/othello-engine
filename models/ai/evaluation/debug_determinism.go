package evaluation

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include "othello_cuda.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Coloc3G/othello-engine/models/game"
)

// DebugEvaluationResult holds the breakdown of all evaluation components for debugging
type DebugEvaluationResult struct {
	// Configuration and raw scores
	Phase             int
	RawMaterialScore  int
	MaterialCoeff     int
	RawMobilityScore  int
	MobilityCoeff     int
	RawCornerScore    int
	CornerCoeff       int
	RawParityScore    int
	ParityCoeff       int
	RawStabilityScore int
	StabilityCoeff    int
	RawFrontierScore  int
	FrontierCoeff     int

	// Weighted contributions
	MaterialContribution  int
	MobilityContribution  int
	CornerContribution    int
	ParityContribution    int
	StabilityContribution int
	FrontierContribution  int

	// Final score
	FinalScore int
}

// DebugEvaluate performs a detailed evaluation of a board position, returning all components
func DebugEvaluate(board game.Board, player game.Piece) *DebugEvaluationResult {
	if !IsGPUAvailable() {
		return nil
	}

	// Flatten the board for C
	flatBoard := make([]C.int, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			flatBoard[i*8+j] = C.int(board[i][j])
		}
	}

	// Allocate memory for debug info
	debugInfo := make([]C.int, 20) // Array to hold all components

	// Call the debug evaluate function
	boardC := (*C.int)(unsafe.Pointer(&flatBoard[0]))
	playerC := C.int(int(player))
	debugInfoC := (*C.int)(unsafe.Pointer(&debugInfo[0]))

	// Call C function
	finalScore := C.debugEvaluateBoard(boardC, playerC, debugInfoC)

	// Create result object
	result := &DebugEvaluationResult{
		Phase:                 int(debugInfo[0]),
		RawMaterialScore:      int(debugInfo[1]),
		MaterialCoeff:         int(debugInfo[2]),
		RawMobilityScore:      int(debugInfo[3]),
		MobilityCoeff:         int(debugInfo[4]),
		RawCornerScore:        int(debugInfo[5]),
		CornerCoeff:           int(debugInfo[6]),
		RawParityScore:        int(debugInfo[7]),
		ParityCoeff:           int(debugInfo[8]),
		RawStabilityScore:     int(debugInfo[9]),
		StabilityCoeff:        int(debugInfo[10]),
		RawFrontierScore:      int(debugInfo[11]),
		FrontierCoeff:         int(debugInfo[12]),
		MaterialContribution:  int(debugInfo[13]),
		MobilityContribution:  int(debugInfo[14]),
		CornerContribution:    int(debugInfo[15]),
		ParityContribution:    int(debugInfo[16]),
		StabilityContribution: int(debugInfo[17]),
		FrontierContribution:  int(debugInfo[18]),
		FinalScore:            int(debugInfo[19]),
	}

	// Verify the sum matches the final score
	calculatedSum := result.MaterialContribution +
		result.MobilityContribution +
		result.CornerContribution +
		result.ParityContribution +
		result.StabilityContribution +
		result.FrontierContribution

	if calculatedSum != result.FinalScore || int(finalScore) != result.FinalScore {
		fmt.Printf("WARNING: Score mismatch - returned: %d, calculated: %d, debugInfo: %d\n",
			finalScore, calculatedSum, result.FinalScore)
	}

	return result
}

// CompareCPUGPUEvaluation evaluates the same position with both CPU and GPU and prints differences
func CompareCPUGPUEvaluation(board game.Board, player game.Player) {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for comparison")
		return
	}

	// First get the GPU detailed evaluation
	gpuDetail := DebugEvaluate(board, player.Color)
	if gpuDetail == nil {
		fmt.Println("Failed to get GPU evaluation details")
		return
	}

	// Calculate the same values on CPU side
	var g game.Game // Dummy game
	cpuEval := NewMixedEvaluationWithCoefficients(V1Coeff)

	// Calculate CPU scores - we have to do this manually to get the raw components
	pieceCount := 0
	for _, row := range board {
		for _, piece := range row {
			if piece != game.Empty {
				pieceCount++
			}
		}
	}

	// Determine phase
	phase := 0
	if pieceCount < 20 {
		phase = 0 // Early
	} else if pieceCount <= 58 {
		phase = 1 // Mid
	} else {
		phase = 2 // Late
	}

	// Get raw scores
	materialScore := cpuEval.MaterialEvaluation.rawEvaluate(board, player)
	mobilityScore := cpuEval.MobilityEvaluation.rawEvaluate(board, player)
	cornersScore := cpuEval.CornersEvaluation.rawEvaluate(board, player)
	parityScore := cpuEval.ParityEvaluation.Evaluate(g, board, player)
	stabilityScore := cpuEval.StabilityEvaluation.Evaluate(g, board, player)
	frontierScore := cpuEval.FrontierEvaluation.rawEvaluate(board, player)

	// Get coefficients
	materialCoeff := cpuEval.MaterialCoeff[phase]
	mobilityCoeff := cpuEval.MobilityCoeff[phase]
	cornersCoeff := cpuEval.CornersCoeff[phase]
	parityCoeff := cpuEval.ParityCoeff[phase]
	stabilityCoeff := cpuEval.StabilityCoeff[phase]
	frontierCoeff := cpuEval.FrontierCoeff[phase]

	// Calculate contributions
	materialContrib := materialCoeff * materialScore
	mobilityContrib := mobilityCoeff * mobilityScore
	cornersContrib := cornersCoeff * cornersScore
	parityContrib := parityCoeff * parityScore
	stabilityContrib := stabilityCoeff * stabilityScore
	frontierContrib := frontierCoeff * frontierScore

	// Calculate final score
	cpuFinalScore := materialContrib + mobilityContrib + cornersContrib +
		parityContrib + stabilityContrib + frontierContrib

	// Also get the standard CPU score
	cpuStandardScore := cpuEval.Evaluate(g, board, player)

	// Get standard GPU score
	gpuScores := EvaluateStatesCUDA([]game.Board{board}, []game.Piece{player.Color})
	gpuStandardScore := 0
	if len(gpuScores) > 0 {
		gpuStandardScore = gpuScores[0]
	}

	// Print comparison of all values
	fmt.Println("\n=== CPU vs GPU Evaluation Comparison ===")
	fmt.Printf("Board position:\n")
	printBoardForDebugging(board)

	fmt.Printf("\nPlayer: %s, Phase: %d\n", getPlayerName(player.Color), phase)

	// Print comparison of scores
	fmt.Printf("\nFinal scores: CPU=%d, GPU=%d, DetailGPU=%d (diff CPU-GPU: %d)\n",
		cpuStandardScore, gpuStandardScore, gpuDetail.FinalScore,
		cpuStandardScore-gpuStandardScore)

	// Print full comparison table
	fmt.Println("\nComponent       | Raw Score |   Coefficient  |  Contribution")
	fmt.Println("----------------|-----------|----------------|---------------")
	printCompRow("Material", materialScore, gpuDetail.RawMaterialScore,
		materialCoeff, gpuDetail.MaterialCoeff,
		materialContrib, gpuDetail.MaterialContribution)
	printCompRow("Mobility", mobilityScore, gpuDetail.RawMobilityScore,
		mobilityCoeff, gpuDetail.MobilityCoeff,
		mobilityContrib, gpuDetail.MobilityContribution)
	printCompRow("Corners", cornersScore, gpuDetail.RawCornerScore,
		cornersCoeff, gpuDetail.CornerCoeff,
		cornersContrib, gpuDetail.CornerContribution)
	printCompRow("Parity", parityScore, gpuDetail.RawParityScore,
		parityCoeff, gpuDetail.ParityCoeff,
		parityContrib, gpuDetail.ParityContribution)
	printCompRow("Stability", stabilityScore, gpuDetail.RawStabilityScore,
		stabilityCoeff, gpuDetail.StabilityCoeff,
		stabilityContrib, gpuDetail.StabilityContribution)
	printCompRow("Frontier", frontierScore, gpuDetail.RawFrontierScore,
		frontierCoeff, gpuDetail.FrontierCoeff,
		frontierContrib, gpuDetail.FrontierContribution)
	fmt.Println("----------------|-----------|----------------|---------------")
	fmt.Printf("TOTAL           |           |                | %5d vs %5d\n",
		cpuFinalScore, gpuDetail.FinalScore)

	// Identify mismatches
	if cpuFinalScore != gpuDetail.FinalScore {
		fmt.Println("\nMismatches found:")
		if materialScore != gpuDetail.RawMaterialScore || materialCoeff != gpuDetail.MaterialCoeff {
			fmt.Printf("- Material: score %d vs %d, coeff %d vs %d\n",
				materialScore, gpuDetail.RawMaterialScore,
				materialCoeff, gpuDetail.MaterialCoeff)
		}
		if mobilityScore != gpuDetail.RawMobilityScore || mobilityCoeff != gpuDetail.MobilityCoeff {
			fmt.Printf("- Mobility: score %d vs %d, coeff %d vs %d\n",
				mobilityScore, gpuDetail.RawMobilityScore,
				mobilityCoeff, gpuDetail.MobilityCoeff)
		}
		if cornersScore != gpuDetail.RawCornerScore || cornersCoeff != gpuDetail.CornerCoeff {
			fmt.Printf("- Corners: score %d vs %d, coeff %d vs %d\n",
				cornersScore, gpuDetail.RawCornerScore,
				cornersCoeff, gpuDetail.CornerCoeff)
		}
		if parityScore != gpuDetail.RawParityScore || parityCoeff != gpuDetail.ParityCoeff {
			fmt.Printf("- Parity: score %d vs %d, coeff %d vs %d\n",
				parityScore, gpuDetail.RawParityScore,
				parityCoeff, gpuDetail.ParityCoeff)
		}
		if stabilityScore != gpuDetail.RawStabilityScore || stabilityCoeff != gpuDetail.StabilityCoeff {
			fmt.Printf("- Stability: score %d vs %d, coeff %d vs %d\n",
				stabilityScore, gpuDetail.RawStabilityScore,
				stabilityCoeff, gpuDetail.StabilityCoeff)
		}
		if frontierScore != gpuDetail.RawFrontierScore || frontierCoeff != gpuDetail.FrontierCoeff {
			fmt.Printf("- Frontier: score %d vs %d, coeff %d vs %d\n",
				frontierScore, gpuDetail.RawFrontierScore,
				frontierCoeff, gpuDetail.FrontierCoeff)
		}
	}
}

// Helper to print a comparison row
func printCompRow(name string, cpuRaw, gpuRaw, cpuCoeff, gpuCoeff, cpuContrib, gpuContrib int) {
	rawMatch := cpuRaw == gpuRaw
	coeffMatch := cpuCoeff == gpuCoeff
	contribMatch := cpuContrib == gpuContrib

	rawStr := fmt.Sprintf("%4d vs %4d", cpuRaw, gpuRaw)
	if !rawMatch {
		rawStr += "*"
	}

	coeffStr := fmt.Sprintf("%6d vs %6d", cpuCoeff, gpuCoeff)
	if !coeffMatch {
		coeffStr += "*"
	}

	contribStr := fmt.Sprintf("%5d vs %5d", cpuContrib, gpuContrib)
	if !contribMatch {
		contribStr += "*"
	}

	fmt.Printf("%-15s| %10s | %14s | %14s\n", name, rawStr, coeffStr, contribStr)
}

// Helper to return the player name
func getPlayerName(player game.Piece) string {
	if player == game.Black {
		return "BLACK"
	} else if player == game.White {
		return "WHITE"
	}
	return "EMPTY"
}

// UpdateDeterminismTestsToInvestigate runs enhanced determinism tests on problematic positions
func RunDeterminismTestWithCoeffs() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	// Set standard coefficients on GPU
	SetCUDACoefficients(V1Coeff)

	fmt.Println("\n=== Running Determinism Tests With Coefficient Debugging ===")

	// Test case: problematic board position
	testBoard := game.NewGame("", "").Board
	// Set up the board position that was failing
	testBoard[2][2] = game.White
	testBoard[3][2] = game.Black
	testBoard[3][3] = game.White
	testBoard[3][4] = game.Black
	testBoard[4][3] = game.Black
	testBoard[4][4] = game.White

	player := game.Player{Color: game.Black}

	// Run enhanced comparison with debug info
	CompareCPUGPUEvaluation(testBoard, player)

	// Also test with white player
	playerWhite := game.Player{Color: game.White}
	CompareCPUGPUEvaluation(testBoard, playerWhite)

	// Also test some specific moves from the bug report
	validMoves := game.ValidMoves(testBoard, player.Color)
	fmt.Printf("\nTesting individual moves with detailed evaluations:\n")
	for _, move := range validMoves {
		// Apply the move
		newBoard, _ := game.ApplyMoveToBoard(testBoard, player.Color, move)

		fmt.Printf("\n=== After applying move %v ===\n", move)
		CompareCPUGPUEvaluation(newBoard, game.Player{Color: game.White})
	}
}

// RunSignFixTest runs a test of the sign fix for GPU evaluations
func RunSignFixTest() {
	if !IsGPUAvailable() {
		fmt.Println("GPU not available for testing")
		return
	}

	fmt.Println("\n=== Running Sign Fix Test ===")

	// Create test boards with different configurations
	testBoards := createTestBoards()

	// For each test board, compare CPU and GPU evaluations
	for i, board := range testBoards {
		blackPlayer := game.Player{Color: game.Black}
		whitePlayer := game.Player{Color: game.White}

		// Set up evaluators
		cpuEval := NewMixedEvaluationWithCoefficients(V1Coeff)
		SetCUDACoefficients(V1Coeff)

		// Get CPU scores
		var g game.Game
		blackCPUScore := cpuEval.Evaluate(g, board, blackPlayer)
		whiteCPUScore := cpuEval.Evaluate(g, board, whitePlayer)

		// Get GPU scores
		blackGPUScores := EvaluateStatesCUDA([]game.Board{board}, []game.Piece{game.Black})
		whiteGPUScores := EvaluateStatesCUDA([]game.Board{board}, []game.Piece{game.White})

		// Also get debug scores for verification
		blackDebug := DebugEvaluate(board, game.Black)
		whiteDebug := DebugEvaluate(board, game.White)

		fmt.Printf("\nTest Board #%d:\n", i+1)
		printBoardForDebugging(board)

		// Check if scores match
		fmt.Printf("Black - CPU: %d, GPU: %d, Debug: %d, diff: %d\n",
			blackCPUScore, blackGPUScores[0], blackDebug.FinalScore,
			blackCPUScore-blackGPUScores[0])

		fmt.Printf("White - CPU: %d, GPU: %d, Debug: %d, diff: %d\n",
			whiteCPUScore, whiteGPUScores[0], whiteDebug.FinalScore,
			whiteCPUScore-whiteGPUScores[0])

		// Indicate if matches
		blackMatch := abs(blackCPUScore-blackGPUScores[0]) <= 5
		whiteMatch := abs(whiteCPUScore-whiteGPUScores[0]) <= 5

		if blackMatch && whiteMatch {
			fmt.Println("SUCCESS: CPU and GPU evaluations match!")
		} else {
			fmt.Println("ERROR: CPU and GPU evaluations don't match")
		}
	}
}

// createTestBoards creates a variety of test boards
func createTestBoards() []game.Board {
	boards := make([]game.Board, 3)

	// Standard opening position
	boards[0] = game.NewGame("", "").Board
	boards[0][3][3] = game.White
	boards[0][3][4] = game.Black
	boards[0][4][3] = game.Black
	boards[0][4][4] = game.White

	// Test board with edge position
	boards[1] = game.NewGame("", "").Board
	boards[1][2][2] = game.White
	boards[1][3][2] = game.Black
	boards[1][3][3] = game.White
	boards[1][3][4] = game.Black
	boards[1][4][3] = game.Black
	boards[1][4][4] = game.White

	// Test with more complex position
	boards[2] = game.NewGame("", "").Board
	boards[2][0][0] = game.White
	boards[2][0][1] = game.White
	boards[2][0][2] = game.White
	boards[2][0][3] = game.White
	boards[2][0][4] = game.White
	boards[2][1][0] = game.Black
	boards[2][1][1] = game.White
	boards[2][1][2] = game.Black
	boards[2][1][3] = game.Black
	boards[2][1][4] = game.Black
	boards[2][2][0] = game.White
	boards[2][2][1] = game.White
	boards[2][2][2] = game.White
	boards[2][2][3] = game.Black
	boards[2][3][0] = game.Black
	boards[2][3][1] = game.White
	boards[2][3][2] = game.Black
	boards[2][3][3] = game.Black
	boards[2][3][4] = game.Black
	boards[2][4][0] = game.Black
	boards[2][4][1] = game.Black
	boards[2][4][2] = game.White
	boards[2][4][3] = game.Black
	boards[2][4][4] = game.White

	return boards
}
