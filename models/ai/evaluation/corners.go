package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// CornersEvaluation is an evaluation function that scores a board based on the position of the pieces
type CornersEvaluation struct {
}

func NewCornersEvaluation() *CornersEvaluation {
	return &CornersEvaluation{}
}

func (e *CornersEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	playerCorners := 0
	opponentCorners := 0
	opponent := game.GetOpponentColor(player.Color)

	// Check each corner
	if b[0][0] == player.Color {
		playerCorners++
	} else if b[0][0] == opponent {
		opponentCorners++
	}

	if b[0][7] == player.Color {
		playerCorners++
	} else if b[0][7] == opponent {
		opponentCorners++
	}

	if b[7][0] == player.Color {
		playerCorners++
	} else if b[7][0] == opponent {
		opponentCorners++
	}

	if b[7][7] == player.Color {
		playerCorners++
	} else if b[7][7] == opponent {
		opponentCorners++
	}

	return playerCorners - opponentCorners
}
