package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

func NewMaterialEvaluation() *MaterialEvaluation {
	return &MaterialEvaluation{}
}

func (e *MaterialEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

func (e *MaterialEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	return pec.PlayerPieces - pec.OpponentPieces
}
