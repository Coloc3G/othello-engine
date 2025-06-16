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

// Evaluate évalue la stabilité des pièces et utilise une carte de poids prédéfinie
func (e *StabilityEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	opponent := game.GetOtherPlayer(g.Players, player.Color)
	playerScore := 0
	opponentScore := 0

	// Utiliser la carte de stabilité pour évaluer chaque position
	for i := range ai.BoardSize {
		for j := range ai.BoardSize {
			if b[i][j] == player.Color {
				playerScore += ai.StabilityMap[i][j]
			} else if b[i][j] == opponent.Color {
				opponentScore += ai.StabilityMap[i][j]
			}
		}
	}

	return playerScore - opponentScore
}
