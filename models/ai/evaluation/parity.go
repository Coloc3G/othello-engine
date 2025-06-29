package evaluation

import (
	"github.com/Coloc3G/othello-engine/models/ai"
	"github.com/Coloc3G/othello-engine/models/game"
)

type ParityEvaluation struct {
}

func NewParityEvaluation() *ParityEvaluation {
	return &ParityEvaluation{}
}

func (e *ParityEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

func (e *ParityEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	// Count empty squares
	emptyCount := ai.BoardSize*ai.BoardSize - pec.WhitePieces - pec.BlackPieces
	return -((emptyCount%2)*2 - 1)
}
