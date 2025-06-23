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
	MaterialCoeff  []int
	MobilityCoeff  []int
	CornersCoeff   []int
	ParityCoeff    []int
	StabilityCoeff []int
	FrontierCoeff  []int
}

// Coefficients structure for serialization
type EvaluationCoefficients struct {
	// Coefficients for different evaluation functions
	MaterialCoeffs  []int `json:"material_coeff"`
	MobilityCoeffs  []int `json:"mobility_coeff"`
	CornersCoeffs   []int `json:"corners_coeff"`
	ParityCoeffs    []int `json:"parity_coeff"`
	StabilityCoeffs []int `json:"stability_coeff"`
	FrontierCoeffs  []int `json:"frontier_coeff"`
	// Name of the coefficients set
	Name string `json:"name"`
}

func NewMixedEvaluation(coeffs EvaluationCoefficients) *MixedEvaluation {
	return &MixedEvaluation{
		MaterialEvaluation:  NewMaterialEvaluation(),
		MobilityEvaluation:  NewMobilityEvaluation(),
		CornersEvaluation:   NewCornersEvaluation(),
		ParityEvaluation:    NewParityEvaluation(),
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

func (e *MixedEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

// Evaluate implements the Evaluation interface for MixedEvaluation
func (e *MixedEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	if pec.PlayerPieces == 0 {
		return -1000000
	}
	if pec.OpponentPieces == 0 {
		return 1000000
	}
	if pec.IsGameOver {
		if pec.PlayerPieces > pec.OpponentPieces {
			return 1000000
		} else if pec.PlayerPieces < pec.OpponentPieces {
			return -1000000
		}
		return 0
	}

	materialCoeff, mobilityCoeff, cornersCoeff, parityCoeff, stabilityCoeff, frontierCoeff := e.ComputeGamePhaseCoefficients(b)

	// Get all raw evaluation scores without normalization to match CUDA implementation
	materialScore := e.MaterialEvaluation.PECEvaluate(g, b, pec)
	mobilityScore := e.MobilityEvaluation.PECEvaluate(g, b, pec)
	cornersScore := e.CornersEvaluation.PECEvaluate(g, b, pec)
	parityScore := e.ParityEvaluation.PECEvaluate(g, b, pec)
	stabilityScore := e.StabilityEvaluation.PECEvaluate(g, b, pec)
	frontierScore := e.FrontierEvaluation.PECEvaluate(g, b, pec)

	return materialCoeff*materialScore +
		mobilityCoeff*mobilityScore +
		cornersCoeff*cornersScore +
		parityCoeff*parityScore +
		stabilityCoeff*stabilityScore +
		frontierCoeff*frontierScore
}

// ComputeGamePhaseCoefficients computes the coefficients for the evaluation functions based on the number of pieces on the board
func (e *MixedEvaluation) ComputeGamePhaseCoefficients(board game.Board) (int, int, int, int, int, int) {
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
	} else if pieceCount <= 50 {
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
