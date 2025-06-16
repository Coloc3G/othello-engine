package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MobilityEvaluation is an evaluation function that scores a board based on the number of possible moves for each player
type MobilityEvaluation struct {
}

func NewMobilityEvaluation() *MobilityEvaluation {
	return &MobilityEvaluation{}
}

func (e *MobilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	playerMoves := len(game.ValidMoves(b, player.Color))
	opponent := game.GetOpponentColor(player.Color)
	opponentMoves := len(game.ValidMoves(b, opponent))

	return playerMoves - opponentMoves
}
