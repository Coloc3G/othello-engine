package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

type PreEvaluationComputation struct {
	Player             game.Player
	Opponent           game.Player
	PlayerPieces       int
	OpponentPieces     int
	PlayerValidMoves   []game.Position
	OpponentValidMoves []game.Position
	IsGameOver         bool
}

type Evaluation interface {
	// Evaluate the given board state and return a score
	Evaluate(g game.Game, b game.Board, player game.Player) int
	PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int
}
