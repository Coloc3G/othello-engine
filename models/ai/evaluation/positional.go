package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// PositionalEvaluation is an evaluation function that scores a board based on the position of the pieces
type PositionalEvaluation struct {
}

func NewPositionalEvaluation() *PositionalEvaluation {
	return &PositionalEvaluation{}
}

// Evaluate the given board state and return a score
func (e *PositionalEvaluation) Evaluate(board game.Board, player game.Player) int {
	return 0
}
