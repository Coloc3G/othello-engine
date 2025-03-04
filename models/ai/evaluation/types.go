package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

type Evaluation interface {
	// Evaluate the given board state and return a score
	Evaluate(board game.Board, player game.Player) int
}
