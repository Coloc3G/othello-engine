package evaluation

import (
	"sync"

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
	batchSize      int
	batchMutex     sync.Mutex
	boardBatch     []game.Board
	playerBatch    []game.Piece
	batchIsRunning bool

	// Stats
	hitCount  int
	missCount int
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
		batchSize:           8192, // Large batch size for modern GPUs
		batchIsRunning:      false,
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
		return scores[0]
	}

	// If GPU evaluation failed, use CPU fallback
	return e.evaluateCPU(g, b, player)
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
