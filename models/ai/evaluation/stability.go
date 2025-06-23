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

func (e *StabilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

// Evaluate évalue la stabilité des pièces et utilise une carte de poids prédéfinie
func (e *StabilityEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	playerScore := 0
	opponentScore := 0

	// Utiliser la carte de stabilité pour évaluer chaque position
	for i := range ai.BoardSize {
		for j := range ai.BoardSize {
			switch b[i][j] {
			case pec.Player.Color:
				playerScore += ai.StabilityMap[i][j]
			case pec.Opponent.Color:
				opponentScore += ai.StabilityMap[i][j]
			}
		}
	}

	return playerScore - opponentScore
}
