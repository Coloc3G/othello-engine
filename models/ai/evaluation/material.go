package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

func NewMaterialEvaluation() *MaterialEvaluation {
	return &MaterialEvaluation{}
}

func (e *MaterialEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	playerPieces := 0
	opponentPieces := 0
	opponent := game.GetOpponentColor(player.Color)

	for i := range 8 {
		for j := range 8 {
			if b[i][j] == player.Color {
				playerPieces++
			} else if b[i][j] == opponent {
				opponentPieces++
			}
		}
	}

	return playerPieces - opponentPieces
}
