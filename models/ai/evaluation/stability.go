package evaluation

import (
	"github.com/Coloc3G/othello-engine/models/ai"
	"github.com/Coloc3G/othello-engine/models/game"
)

// StabilityEvaluation évalue la stabilité des pièces sur le plateau
type StabilityEvaluation struct{}

func NewStabilityEvaluation() *StabilityEvaluation {
	return &StabilityEvaluation{}
}

func (e *StabilityEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

// Evaluate évalue la stabilité des pièces et utilise une carte de poids prédéfinie
func (e *StabilityEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	var whiteScore, blackScore int16

	// Iterate through all positions using bit operations
	for pos := range 64 {
		mask := uint64(1) << pos
		if b.WhitePieces&mask != 0 {
			row := pos / 8
			col := pos % 8
			whiteScore += ai.StabilityMap[row][col]
		} else if b.BlackPieces&mask != 0 {
			row := pos / 8
			col := pos % 8
			blackScore += ai.StabilityMap[row][col]
		}
	}

	return whiteScore - blackScore
}
