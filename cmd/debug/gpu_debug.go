package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
)

func main() {
	// Parse command line flags
	runTests := flag.Bool("test", false, "Run GPU/CPU determinism tests")
	runConsistency := flag.Bool("consistency", false, "Run GPU/CPU consistency checks")
	runPerf := flag.Bool("perf", false, "Run GPU performance benchmark")
	runAll := flag.Bool("all", false, "Run all tests")
	runCoeffDebug := flag.Bool("coeffs", false, "Run coefficient debug tests")
	fixSign := flag.Bool("fixsign", false, "Test and fix sign issue between GPU and CPU")
	flag.Parse()

	// Check if GPU is available
	gpuAvailable := evaluation.InitCUDA()
	if !gpuAvailable {
		fmt.Println("ERROR: GPU acceleration not available")
		os.Exit(1)
	}

	fmt.Println("=== Othello GPU Debug Tool ===")
	fmt.Println("GPU acceleration available")

	// Run determinism tests if requested
	if *runTests || *runAll {
		evaluation.TestDeterminism()
		evaluation.TestComplexPosition()
	}

	// Run consistency checks if requested
	if *runConsistency || *runAll {
		evaluation.DebugGPUCPUConsistency()
	}

	// Run performance benchmark if requested
	if *runPerf || *runAll {
		evaluation.RunGPUBenchmark()
	}

	// Run coefficient debug tests if requested
	if *runCoeffDebug || *runAll {
		evaluation.RunDeterminismTestWithCoeffs()
	}

	// Test sign fix
	if *fixSign || *runAll {
		testSignFix()
		evaluation.RunSignFixTest()
	}

	// If no specific tests were requested and not running all, show help
	if !*runTests && !*runConsistency && !*runPerf && !*runAll && !*runCoeffDebug && !*fixSign {
		fmt.Println("\nNo tests specified. Use flags to run tests:")
		fmt.Println("  -test        Run GPU/CPU determinism tests")
		fmt.Println("  -consistency Run GPU/CPU consistency checks")
		fmt.Println("  -perf        Run GPU performance benchmark")
		fmt.Println("  -coeffs      Run coefficient debug tests")
		fmt.Println("  -fixsign     Test sign fix between GPU and CPU")
		fmt.Println("  -all         Run all tests")
	}

	// Clean up CUDA resources
	evaluation.CleanupCUDA()
}

// Test function for sign issue fix
func testSignFix() {
	fmt.Println("\n=== Testing Sign Fix ===")

	// Create a test board
	testBoard := game.NewGame("", "").Board
	testBoard[2][2] = game.White
	testBoard[3][2] = game.Black
	testBoard[3][3] = game.White
	testBoard[3][4] = game.Black
	testBoard[4][3] = game.Black
	testBoard[4][4] = game.White

	// Set standard coefficients
	evaluation.SetCUDACoefficients(evaluation.V1Coeff)

	// Create players
	blackPlayer := game.Player{Color: game.Black}
	whitePlayer := game.Player{Color: game.White}

	// Evaluate with CPU
	cpuEval := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V1Coeff)
	var g game.Game // dummy game
	blackCPUScore := cpuEval.Evaluate(g, testBoard, blackPlayer)
	whiteCPUScore := cpuEval.Evaluate(g, testBoard, whitePlayer)

	// Evaluate with GPU - get debug scores first which are reliable
	blackDebug := evaluation.DebugEvaluate(testBoard, blackPlayer.Color)
	whiteDebug := evaluation.DebugEvaluate(testBoard, whitePlayer.Color)

	// Then test the normal evaluation path that should be fixed
	gpuScoresBlack := evaluation.EvaluateStatesCUDA([]game.Board{testBoard}, []game.Piece{blackPlayer.Color})
	gpuScoresWhite := evaluation.EvaluateStatesCUDA([]game.Board{testBoard}, []game.Piece{whitePlayer.Color})

	// Print results
	fmt.Printf("Black player evaluation:\n")
	fmt.Printf("  CPU: %d\n", blackCPUScore)
	fmt.Printf("  GPU (normal): %d\n", gpuScoresBlack[0])
	fmt.Printf("  GPU (debug): %d\n", blackDebug.FinalScore)
	fmt.Printf("  Diff CPU-GPU: %d\n", blackCPUScore-gpuScoresBlack[0])

	fmt.Printf("White player evaluation:\n")
	fmt.Printf("  CPU: %d\n", whiteCPUScore)
	fmt.Printf("  GPU (normal): %d\n", gpuScoresWhite[0])
	fmt.Printf("  GPU (debug): %d\n", whiteDebug.FinalScore)
	fmt.Printf("  Diff CPU-GPU: %d\n", whiteCPUScore-gpuScoresWhite[0])

	// Test if scores match
	blackMatch := abs(blackCPUScore-gpuScoresBlack[0]) <= 5 &&
		abs(blackCPUScore-blackDebug.FinalScore) <= 5
	whiteMatch := abs(whiteCPUScore-gpuScoresWhite[0]) <= 5 &&
		abs(whiteCPUScore-whiteDebug.FinalScore) <= 5

	if blackMatch && whiteMatch {
		fmt.Println("SUCCESS: CPU, GPU normal, and GPU debug evaluations all match!")
	} else if abs(blackDebug.FinalScore-gpuScoresBlack[0]) <= 5 &&
		abs(whiteDebug.FinalScore-gpuScoresWhite[0]) <= 5 {
		fmt.Println("WARNING: GPU evaluations match each other but not CPU")
		fmt.Println("This may indicate that the GPU and CPU implementations calculate scores differently")
	} else {
		fmt.Println("ERROR: Evaluations don't match consistently")
	}
}

// Simple abs function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
