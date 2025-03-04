package evaluation

// MixedEvaluation is a struct that contains the evaluation of a board state using a mix of different evaluation functions
type MixedEvaluation struct {
	// The evaluation of the board state using the material evaluation function
	MaterialEvaluation MaterialEvaluation
	// The evaluation of the board state using the mobility evaluation function
	MobilityEvaluation MobilityEvaluation
	// The evaluation of the board state using the positional evaluation function
	PositionalEvaluation PositionalEvaluation
}

// Evaluate the given board state and return a score
func (e *MixedEvaluation) Evaluate(board [8][8]int) int {
	materialCoeff, mobilityCoeff, positionalCoeff := e.ComputeGamePhaseCoefficients(board)
	materialScore := e.MaterialEvaluation.Evaluate(board)
	mobilityScore := e.MobilityEvaluation.Evaluate(board)
	positionalScore := e.PositionalEvaluation.Evaluate(board)
	return materialCoeff*materialScore + mobilityCoeff*mobilityScore + positionalCoeff*positionalScore
}

func (e *MixedEvaluation) ComputeGamePhaseCoefficients(board [8][8]int) (int, int, int) {
	return 1, 1, 1
}
