package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// CornersEvaluation is an evaluation function that scores a board based on the position of the pieces
type CornersEvaluation struct {
}

func NewCornersEvaluation() *CornersEvaluation {
	return &CornersEvaluation{}
}

func (e *CornersEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

func (e *CornersEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	playerCorners := 0
	opponentCorners := 0

	// Check each corner
	switch b[0][0] {
	case pec.Player.Color:
		playerCorners++
	case pec.Opponent.Color:
		opponentCorners++
	}

	switch b[0][7] {
	case pec.Player.Color:
		playerCorners++
	case pec.Opponent.Color:
		opponentCorners++
	}

	switch b[7][0] {
	case pec.Player.Color:
		playerCorners++
	case pec.Opponent.Color:
		opponentCorners++
	}

	switch b[7][7] {
	case pec.Player.Color:
		playerCorners++
	case pec.Opponent.Color:
		opponentCorners++
	}

	return playerCorners - opponentCorners
}
