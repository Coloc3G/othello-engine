package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

func NewMaterialEvaluation() *MaterialEvaluation {
	return &MaterialEvaluation{}
}

// Add a raw evaluation function that doesn't normalize the score
func (e *MaterialEvaluation) rawEvaluate(b game.Board, player game.Player) int {
	playerPieces := 0
	opponentPieces := 0
	opponent := game.GetOpponentColor(player.Color)

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if b[i][j] == player.Color {
				playerPieces++
			} else if b[i][j] == opponent {
				opponentPieces++
			}
		}
	}

	// Simple difference, no normalization
	return playerPieces - opponentPieces
}

// Evaluate the material advantage (number of pieces)
func (e *MaterialEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	return e.rawEvaluate(b, player)
}
