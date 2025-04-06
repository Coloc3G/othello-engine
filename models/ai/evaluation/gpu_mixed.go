package evaluation

import (
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
		batchSize:           1024, // Can be tuned
	}
}

// Evaluate the given board state using GPU if available, fallback to CPU if not
func (e *GPUMixedEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	if !IsGPUAvailable() {
		// Fallback to CPU evaluation if GPU is not available
		return e.evaluateCPU(g, b, player)
	}

	// For single evaluation, use the GPU but with just one board
	boards := []game.Board{b}
	players := []game.Piece{player.Color}

	// Evaluate on GPU
	scores := EvaluateStatesCUDA(boards, players)

	if len(scores) == 0 {
		// GPU evaluation failed, fallback to CPU
		return e.evaluateCPU(g, b, player)
	}

	return scores[0]
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
