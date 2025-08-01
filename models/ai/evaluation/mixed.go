package evaluation

import (
	"github.com/Coloc3G/othello-engine/models/game"
)

// MixedEvaluation is a struct that contains the evaluation of a board state using a mix of different evaluation functions
type MixedEvaluation struct {
	// The evaluation of the board state using the material evaluation function
	MaterialEvaluation *MaterialEvaluation
	// The evaluation of the board state using the mobility evaluation function
	MobilityEvaluation *MobilityEvaluation
	// The evaluation of the board state using the corners evaluation function
	CornersEvaluation *CornersEvaluation
	// The evaluation of the board state using the parity evaluation function
	ParityEvaluation *ParityEvaluation
	// The evaluation of the board state using the stability evaluation function
	StabilityEvaluation *StabilityEvaluation
	// The evaluation of the board state using the frontier evaluation function
	FrontierEvaluation *FrontierEvaluation
	// Coefficients for different game phases
	MaterialCoeff  []int16
	MobilityCoeff  []int16
	CornersCoeff   []int16
	ParityCoeff    []int16
	StabilityCoeff []int16
	FrontierCoeff  []int16
}

// Coefficients structure for serialization
type EvaluationCoefficients struct {
	// Coefficients for different evaluation functions
	MaterialCoeffs  []int16 `json:"material_coeff"`
	MobilityCoeffs  []int16 `json:"mobility_coeff"`
	CornersCoeffs   []int16 `json:"corners_coeff"`
	ParityCoeffs    []int16 `json:"parity_coeff"`
	StabilityCoeffs []int16 `json:"stability_coeff"`
	FrontierCoeffs  []int16 `json:"frontier_coeff"`
	// Name of the coefficients set
	Name string `json:"name"`
}

func NewMixedEvaluation(coeffs EvaluationCoefficients) *MixedEvaluation {
	return &MixedEvaluation{
		MaterialEvaluation:  NewMaterialEvaluation(),
		MobilityEvaluation:  NewMobilityEvaluation(),
		CornersEvaluation:   NewCornersEvaluation(),
		StabilityEvaluation: NewStabilityEvaluation(),
		FrontierEvaluation:  NewFrontierEvaluation(),
		MaterialCoeff:       coeffs.MaterialCoeffs,
		MobilityCoeff:       coeffs.MobilityCoeffs,
		CornersCoeff:        coeffs.CornersCoeffs,
		ParityCoeff:         coeffs.ParityCoeffs,
		StabilityCoeff:      coeffs.StabilityCoeffs,
		FrontierCoeff:       coeffs.FrontierCoeffs,
	}
}

func (e *MixedEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

// Evaluate implements the Evaluation interface for MixedEvaluation
func (e *MixedEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	if pec.WhitePieces == 0 {
		return MIN_EVAL - 64
	}
	if pec.BlackPieces == 0 {
		return MAX_EVAL + 64
	}
	if pec.IsGameOver {
		if pec.WhitePieces > pec.BlackPieces {
			return MAX_EVAL + pec.WhitePieces - pec.BlackPieces
		} else if pec.WhitePieces < pec.BlackPieces {
			return MIN_EVAL - pec.BlackPieces + pec.WhitePieces
		}
		return 0
	}

	materialCoeff, mobilityCoeff, cornersCoeff, parityCoeff, stabilityCoeff, frontierCoeff := e.ComputeGamePhaseCoefficients(pec)

	// Get all raw evaluation scores without normalization to match CUDA implementation
	materialScore := e.MaterialEvaluation.PECEvaluate(b, pec)
	mobilityScore := e.MobilityEvaluation.PECEvaluate(b, pec)
	cornersScore := e.CornersEvaluation.PECEvaluate(b, pec)
	parityScore := e.ParityEvaluation.PECEvaluate(b, pec)
	stabilityScore := e.StabilityEvaluation.PECEvaluate(b, pec)
	frontierScore := e.FrontierEvaluation.PECEvaluate(b, pec)

	if pec.Debug {
		println("materialCoeff:", materialCoeff, "\tmaterialScore:", materialScore)
		println("mobilityCoeff:", mobilityCoeff, "\tmobilityScore:", mobilityScore)
		println("cornersCoeff:", cornersCoeff, "\tcornersScore:", cornersScore)
		println("parityCoeff:", parityCoeff, "\tparityScore:", parityScore)
		println("stabilityCoeff:", stabilityCoeff, "\tstabilityScore:", stabilityScore)
		println("frontierCoeff:", frontierCoeff, "\tfrontierScore:", frontierScore)
		println("Resulting score:", materialCoeff*materialScore+
			mobilityCoeff*mobilityScore+
			cornersCoeff*cornersScore+
			parityCoeff*parityScore+
			stabilityCoeff*stabilityScore+
			frontierCoeff*frontierScore)
	}

	return materialCoeff*materialScore +
		mobilityCoeff*mobilityScore +
		cornersCoeff*cornersScore +
		parityCoeff*parityScore +
		stabilityCoeff*stabilityScore +
		frontierCoeff*frontierScore
}

// ComputeGamePhaseCoefficients computes the coefficients for the evaluation functions based on the number of pieces on the board
func (e *MixedEvaluation) ComputeGamePhaseCoefficients(pec PreEvaluationComputation) (int16, int16, int16, int16, int16, int16) {
	piecesCount := pec.WhitePieces + pec.BlackPieces
	var phase int
	if piecesCount < 10 {
		phase = 0 // Early game
	} else if piecesCount <= 20 {
		phase = 1 // Mid game
	} else if piecesCount <= 35 {
		phase = 2 // Mid game
	} else if piecesCount <= 50 {
		phase = 3 // Mid game
	} else if piecesCount <= 55 {
		phase = 4 // Mid game
	} else {
		phase = 5 // Late game
	}

	return e.MaterialCoeff[phase],
		e.MobilityCoeff[phase],
		e.CornersCoeff[phase],
		e.ParityCoeff[phase],
		e.StabilityCoeff[phase],
		e.FrontierCoeff[phase]
}
