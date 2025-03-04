package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// CornersEvaluation is an evaluation function that scores a board based on the position of the pieces
type CornersEvaluation struct {
}

func NewCornersEvaluation() *CornersEvaluation {
	return &CornersEvaluation{}
}

// Evaluate the given board state and return a score
func (e *CornersEvaluation) Evaluate(board game.Board, player game.Player) int {
	myCorners := 0
	opCorners := 0

	if board[0][0] == player.Color {
		myCorners++
	}
	if board[7][0] == player.Color {
		myCorners++
	}
	if board[0][7] == player.Color {
		myCorners++
	}
	if board[7][7] == player.Color {
		myCorners++
	}

	if board[0][0] != player.Color && board[0][0] != game.Empty {
		opCorners++
	}
	if board[7][0] != player.Color && board[7][0] != game.Empty {
		opCorners++
	}
	if board[0][7] != player.Color && board[0][7] != game.Empty {
		opCorners++
	}
	if board[7][7] != player.Color && board[7][7] != game.Empty {
		opCorners++
	}

	return 100 * (myCorners - opCorners) / (myCorners + opCorners + 1)
}
