package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// MobilityEvaluation is an evaluation function that scores a board based on the number of possible moves for each player
type MobilityEvaluation struct {
}

func NewMobilityEvaluation() *MobilityEvaluation {
	return &MobilityEvaluation{}
}

func (e *MobilityEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

func (e *MobilityEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	return int16(len(pec.WhiteValidMoves) - len(pec.BlackValidMoves))
}
