package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

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
}

func NewMixedEvaluation() *MixedEvaluation {
	return &MixedEvaluation{
		MaterialEvaluation: NewMaterialEvaluation(),
		MobilityEvaluation: NewMobilityEvaluation(),
		CornersEvaluation:  NewCornersEvaluation(),
		ParityEvaluation:   NewParityEvaluation(),
	}
}

// Evaluate the given board state and return a score
func (e *MixedEvaluation) Evaluate(board game.Board, player game.Player) int {
	materialCoeff, mobilityCoeff, cornersCoeff, parityCoeff := e.ComputeGamePhaseCoefficients(board)
	materialScore := e.MaterialEvaluation.Evaluate(board, player)
	mobilityScore := e.MobilityEvaluation.Evaluate(board, player)
	cornersScore := e.CornersEvaluation.Evaluate(board, player)
	parityScore := e.ParityEvaluation.Evaluate(board, player)
	return materialCoeff*materialScore + mobilityCoeff*mobilityScore + cornersCoeff*cornersScore + parityCoeff*parityScore
}

// ComputeGamePhaseCoefficients computes the coefficients for the evaluation functions based on the number of pieces on the board
func (e *MixedEvaluation) ComputeGamePhaseCoefficients(board game.Board) (int, int, int, int) {
	pieceCount := 0
	for _, row := range board {
		for _, piece := range row {
			if piece != game.Empty {
				pieceCount++
			}
		}
	}

	if pieceCount == 64 {
		// Game over
		return 1000, 0, 0, 0
	}

	if pieceCount < 20 {
		// Early game
		return 0, 50, 1000, 0
	} else if pieceCount <= 58 {
		// Mid game
		return 10, 20, 1000, 100
	} else {
		// Late game
		return 500, 100, 1000, 500
	}
}
