package evaluation

import (
	"encoding/json"
	"os"

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
		return MIN_EVAL
	}
	if pec.BlackPieces == 0 {
		return MAX_EVAL
	}
	if pec.IsGameOver {
		if pec.WhitePieces > pec.BlackPieces {
			return MAX_EVAL
		} else if pec.WhitePieces < pec.BlackPieces {
			return MIN_EVAL
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
	if piecesCount < 20 {
		phase = 0 // Early game
	} else if piecesCount <= 50 {
		phase = 1 // Mid game
	} else {
		phase = 2 // Late game
	}

	return e.MaterialCoeff[phase],
		e.MobilityCoeff[phase],
		e.CornersCoeff[phase],
		e.ParityCoeff[phase],
		e.StabilityCoeff[phase],
		e.FrontierCoeff[phase]
}

// SaveCoefficients saves the current coefficients to a file
func (e *MixedEvaluation) SaveCoefficients(filename string) error {
	coeffs := EvaluationCoefficients{
		MaterialCoeffs:  e.MaterialCoeff,
		MobilityCoeffs:  e.MobilityCoeff,
		CornersCoeffs:   e.CornersCoeff,
		ParityCoeffs:    e.ParityCoeff,
		StabilityCoeffs: e.StabilityCoeff,
		FrontierCoeffs:  e.FrontierCoeff,
	}

	data, err := json.MarshalIndent(coeffs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadCoefficients loads coefficients from a file
func LoadCoefficients(filename string) (EvaluationCoefficients, error) {
	var coeffs EvaluationCoefficients

	data, err := os.ReadFile(filename)
	if err != nil {
		return coeffs, err
	}

	err = json.Unmarshal(data, &coeffs)
	return coeffs, err
}
