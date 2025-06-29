package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

type PreEvaluationComputation struct {
	WhitePieces     int16
	BlackPieces     int16
	WhiteValidMoves []game.Position
	BlackValidMoves []game.Position
	IsGameOver      bool
	Debug           bool // For debugging purposes, can be set to true to print debug information
}

type Evaluation interface {
	// Evaluate the given board state and return a score
	Evaluate(bb game.BitBoard) int16
	PECEvaluate(bb game.BitBoard, pec PreEvaluationComputation) int16
}
