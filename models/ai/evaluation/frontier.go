package evaluation

import (
	"github.com/Coloc3G/othello-engine/models/ai"
	"github.com/Coloc3G/othello-engine/models/game"
)

// FrontierEvaluation évalue le nombre de pièces frontalières (adjacentes à des cases vides)
// Ces pièces sont généralement vulnérables et peuvent être retournées
type FrontierEvaluation struct{}

func NewFrontierEvaluation() *FrontierEvaluation {
	return &FrontierEvaluation{}
}

func (e *FrontierEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	opponent := game.GetOtherPlayer(g.Players, player.Color)

	playerFrontier := countFrontierDiscs(b, player.Color)
	opponentFrontier := countFrontierDiscs(b, opponent.Color)

	// On veut minimiser le nombre de pièces frontalières, donc un score négatif
	if playerFrontier+opponentFrontier == 0 {
		return 0
	}

	return -100 * (playerFrontier - opponentFrontier) / (playerFrontier + opponentFrontier)
}

// countFrontierDiscs compte le nombre de pièces frontalières pour un joueur donné
func countFrontierDiscs(b game.Board, color game.Piece) int {
	frontierCount := 0
	directions := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}

	for i := 0; i < ai.BoardSize; i++ {
		for j := 0; j < ai.BoardSize; j++ {
			if b[i][j] == color {
				// Vérifier si la pièce est adjacente à une case vide
				for _, dir := range directions {
					dx, dy := dir[0], dir[1]
					r, c := i+dx, j+dy

					if r >= 0 && r < ai.BoardSize && c >= 0 && c < ai.BoardSize && b[r][c] == game.Empty {
						frontierCount++
						break // Une pièce est comptée une seule fois même si elle est adjacente à plusieurs cases vides
					}
				}
			}
		}
	}

	return frontierCount
}
