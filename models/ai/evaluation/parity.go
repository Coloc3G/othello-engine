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

func (e *ParityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

// Evaluate computes the parity score
func (e *ParityEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	// Count empty squares
	emptyCount := ai.BoardSize*ai.BoardSize - pec.PlayerPieces - pec.OpponentPieces

	// Determine parity advantage based on player color and empty square count
	switch emptyCount % 2 {
	case 0:
		switch pec.Player.Color {
		case game.Black:
			return -1
		default:
			return 1
		}
	default:
		switch pec.Player.Color {
		case game.Black:
			return 1
		default:
			return -1
		}
	}
}
