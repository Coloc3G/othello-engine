package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MaterialEvaluation is an evaluation function that scores a board based on the number of pieces difference between the players
type MaterialEvaluation struct {
}

func NewMaterialEvaluation() *MaterialEvaluation {
	return &MaterialEvaluation{}
}

func (e *MaterialEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

func (e *MaterialEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	return pec.WhitePieces - pec.BlackPieces
}
