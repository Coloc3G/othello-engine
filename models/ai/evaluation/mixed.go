package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MixedEvaluation is a struct that contains the evaluation of a board state using a mix of different evaluation functions
type MixedEvaluation struct {
	// The evaluation of the board state using the material evaluation function
	MaterialEvaluation *MaterialEvaluation
	// The evaluation of the board state using the mobility evaluation function
	MobilityEvaluation *MobilityEvaluation
	// The evaluation of the board state using the positional evaluation function
	PositionalEvaluation *PositionalEvaluation
}

func NewMixedEvaluation() *MixedEvaluation {
	return &MixedEvaluation{
		MaterialEvaluation:   NewMaterialEvaluation(),
		MobilityEvaluation:   NewMobilityEvaluation(),
		PositionalEvaluation: NewPositionalEvaluation(),
	}
}

// Evaluate the given board state and return a score
func (e *MixedEvaluation) Evaluate(board game.Board, player game.Player) int {
	materialCoeff, mobilityCoeff, positionalCoeff := e.ComputeGamePhaseCoefficients(board)
	materialScore := e.MaterialEvaluation.Evaluate(board, player)
	mobilityScore := e.MobilityEvaluation.Evaluate(board, player)
	positionalScore := e.PositionalEvaluation.Evaluate(board, player)
	return materialCoeff*materialScore + mobilityCoeff*mobilityScore + positionalCoeff*positionalScore
}

func (e *MixedEvaluation) ComputeGamePhaseCoefficients(board game.Board) (int, int, int) {
	return 1, 1, 1
}
