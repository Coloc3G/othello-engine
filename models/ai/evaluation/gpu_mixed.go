package evaluation

import (
	"fmt"
	"math/rand"

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
	batchSize int
}

// NewGPUMixedEvaluation creates a new GPU-accelerated mixed evaluation
func NewGPUMixedEvaluation(coeffs EvaluationCoefficients) *GPUMixedEvaluation {
	// Set the coefficients immediately if we're creating a GPU evaluator
	if IsGPUAvailable() {
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
		batchSize:           256, // More moderate batch size
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

// Evaluate the given board state using GPU if available, fallback to CPU if not
func (e *GPUMixedEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	if !IsGPUAvailable() {
		// Fallback to CPU evaluation if GPU is not available
		return e.evaluateCPU(g, b, player)
	}

	// Direct GPU evaluation - no caching
	scores := EvaluateStatesCUDA([]game.Board{b}, []game.Piece{player.Color})
	if len(scores) > 0 {
		// Add verification - occasionally compare with CPU evaluation
		if rand.Float64() < 0.01 { // Increased verification rate from 0.001 to 0.01
			cpuScore := e.evaluateCPU(g, b, player)
			if abs(scores[0]-cpuScore) > 10 { // Reduced threshold for being more strict
				// Log the discrepancy and use CPU score instead
				fmt.Printf("ERROR: GPU/CPU eval difference: GPU=%d, CPU=%d (diff=%d)\n",
					scores[0], cpuScore, scores[0]-cpuScore)
				// Debug: print board
				fmt.Printf("Board causing evaluation difference:\n")
				printBoardDetail(b)
				return cpuScore // Use CPU score as it's more reliable
			}
		}
		return scores[0]
	}

	// If GPU evaluation failed, use CPU fallback
	return e.evaluateCPU(g, b, player)
}

// PrintBoardDetail prints a board in detailed format
func printBoardDetail(b game.Board) {
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

// VerifyWithCPU performs explicit verification of GPU vs CPU evaluations
func (e *GPUMixedEvaluation) VerifyWithCPU(g game.Game, b game.Board, player game.Player) bool {
	// Get GPU evaluation
	gpuScore := e.Evaluate(g, b, player)

	// Get CPU evaluation
	cpuScore := e.evaluateCPU(g, b, player)

	// Compare results
	diff := abs(gpuScore - cpuScore)
	if diff > 0 {
		fmt.Printf("GPU/CPU evaluation difference: GPU=%d, CPU=%d (diff=%d)\n",
			gpuScore, cpuScore, diff)
		return false
	}

	return true
}

// Simple absolute value function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// evaluateCPU is a fallback CPU implementation
func (e *GPUMixedEvaluation) evaluateCPU(g game.Game, b game.Board, player game.Player) int {
	materialCoeff, mobilityCoeff, cornersCoeff, parityCoeff, stabilityCoeff, frontierCoeff := e.computeCoefficients(b)

	// Use the raw evaluation functions that match the CUDA implementation
	materialScore := e.MaterialEvaluation.rawEvaluate(b, player)
	mobilityScore := e.MobilityEvaluation.rawEvaluate(b, player)
	cornersScore := e.CornersEvaluation.rawEvaluate(b, player)
	parityScore := e.ParityEvaluation.Evaluate(g, b, player)
	stabilityScore := e.StabilityEvaluation.Evaluate(g, b, player)
	frontierScore := e.FrontierEvaluation.rawEvaluate(b, player)

	// Calculate final score using same formula as CUDA
	score := materialCoeff*materialScore +
		mobilityCoeff*mobilityScore +
		cornersCoeff*cornersScore +
		parityCoeff*parityScore +
		stabilityCoeff*stabilityScore +
		frontierCoeff*frontierScore

	return score
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
