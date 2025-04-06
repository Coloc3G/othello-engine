package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MobilityEvaluation is an evaluation function that scores a board based on the number of possible moves for each player
type MobilityEvaluation struct {
}

func NewMobilityEvaluation() *MobilityEvaluation {
	return &MobilityEvaluation{}
}

// Add a raw evaluation function that doesn't normalize the score
func (e *MobilityEvaluation) rawEvaluate(b game.Board, player game.Player) int {
	playerMoves := len(game.ValidMoves(b, player.Color))
	opponent := game.GetOpponentColor(player.Color)
	opponentMoves := len(game.ValidMoves(b, opponent))

	// Simple difference, no normalization
	return playerMoves - opponentMoves
}

// Evaluate computes the mobility score
func (e *MobilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	return e.rawEvaluate(b, player)
}
