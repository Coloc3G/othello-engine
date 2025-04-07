package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

func NewMaterialEvaluation() *MaterialEvaluation {
	return &MaterialEvaluation{}
}

// Evaluate the given board state and return a score
func (e *MaterialEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	sum := 0
	for _, row := range b {
		for _, piece := range row {
			sum += int(piece)
		}
	}
	return sum
}
