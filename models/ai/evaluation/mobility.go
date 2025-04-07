package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MobilityEvaluation is an evaluation function that scores a board based on the number of possible moves for each player
type MobilityEvaluation struct {
}

func NewMobilityEvaluation() *MobilityEvaluation {
	return &MobilityEvaluation{}
}

// Evaluate the given board state and return a score
func (e *MobilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {

	opponent := game.GetOtherPlayer(g.Players, player.Color)

	myMoveCount := len(game.ValidMoves(b, player.Color))
	opMoveCount := len(game.ValidMoves(b, opponent.Color))

	return 100 * (myMoveCount - opMoveCount) / (myMoveCount + opMoveCount + 1)
}
