package learning

import (
	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
)

// CustomMixedEvaluation is a customizable version of MixedEvaluation
type CustomMixedEvaluation struct {
	// The evaluation components
	MaterialEvaluation  *evaluation.MaterialEvaluation
	MobilityEvaluation  *evaluation.MobilityEvaluation
	CornersEvaluation   *evaluation.CornersEvaluation
	ParityEvaluation    *evaluation.ParityEvaluation
	StabilityEvaluation *evaluation.StabilityEvaluation
	FrontierEvaluation  *evaluation.FrontierEvaluation

	// Custom coefficients for each game phase
	Coefficients evaluation.EvaluationCoefficients

	// GPU context for accelerated evaluation
	UseGPU bool
}

// NewCustomMixedEvaluation creates a new custom mixed evaluation
func NewCustomMixedEvaluation(coeffs evaluation.EvaluationCoefficients) *CustomMixedEvaluation {
	return &CustomMixedEvaluation{
		MaterialEvaluation:  evaluation.NewMaterialEvaluation(),
		MobilityEvaluation:  evaluation.NewMobilityEvaluation(),
		CornersEvaluation:   evaluation.NewCornersEvaluation(),
		ParityEvaluation:    evaluation.NewParityEvaluation(),
		StabilityEvaluation: evaluation.NewStabilityEvaluation(),
		FrontierEvaluation:  evaluation.NewFrontierEvaluation(),
		Coefficients:        coeffs,
		UseGPU:              false,
	}
}

// EnableGPU enables GPU acceleration for this evaluation
func (e *CustomMixedEvaluation) EnableGPU(enable bool) {
	e.UseGPU = enable
}

// Evaluate the given board state and return a score using custom coefficients
func (e *CustomMixedEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
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

// computeCoefficients returns the coefficients based on game phase and custom settings
func (e *CustomMixedEvaluation) computeCoefficients(board game.Board) (int, int, int, int, int, int) {
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

	return e.Coefficients.MaterialCoeffs[phase],
		e.Coefficients.MobilityCoeffs[phase],
		e.Coefficients.CornersCoeffs[phase],
		e.Coefficients.ParityCoeffs[phase],
		e.Coefficients.StabilityCoeffs[phase],
		e.Coefficients.FrontierCoeffs[phase]
}
