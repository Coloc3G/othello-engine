package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

// CornersEvaluation is an evaluation function that scores a board based on the position of the pieces
type CornersEvaluation struct {
}

func NewCornersEvaluation() *CornersEvaluation {
	return &CornersEvaluation{}
}

func (e *CornersEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

func (e *CornersEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	var whiteCorners, blackCorners int16

	// Define corner positions as bit masks
	const (
		topLeft     = uint64(1) << 63
		topRight    = uint64(1) << 56
		bottomLeft  = uint64(1) << 7
		bottomRight = uint64(1) << 0
	)

	corners := []uint64{topLeft, topRight, bottomLeft, bottomRight}

	for _, corner := range corners {
		if b.WhitePieces&corner != 0 {
			whiteCorners++
		} else if b.BlackPieces&corner != 0 {
			blackCorners++
		}
	}

	return whiteCorners - blackCorners
}
