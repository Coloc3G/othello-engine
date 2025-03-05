package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

type ParityEvaluation struct {
}

func NewParityEvaluation() *ParityEvaluation {
	return &ParityEvaluation{}
}

// Evaluate the given board state and return a score
func (e *ParityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	piecesCount := 0
	for _, row := range b {
		for _, piece := range row {
			if piece != game.Empty {
				piecesCount++
			}
		}
	}
	return (piecesCount%2)*2 - 1
}
