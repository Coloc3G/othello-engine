package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MobilityEvaluation is an evaluation function that scores a board based on the number of possible moves for each player
type MobilityEvaluation struct {
}

func NewMobilityEvaluation() *MobilityEvaluation {
	return &MobilityEvaluation{}
}

func (e *MobilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

func (e *MobilityEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	return len(pec.PlayerValidMoves) - len(pec.OpponentValidMoves)
}
