package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

type ParityEvaluation struct {
}

func NewParityEvaluation() *ParityEvaluation {
	return &ParityEvaluation{}
}

// Evaluate computes the parity score
func (e *ParityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	// Count empty squares
	emptyCount := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if b[i][j] == game.Empty {
				emptyCount++
			}
		}
	}

	// Determine parity advantage based on player color and empty square count
	// Match the CUDA implementation exactly
	if emptyCount%2 == 0 {
		if player.Color == game.Black {
			return -1
		} else {
			return 1
		}
	} else {
		if player.Color == game.Black {
			return 1
		} else {
			return -1
		}
	}
}
