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

// Add a raw evaluation function that doesn't normalize the score
func (e *FrontierEvaluation) rawEvaluate(b game.Board, player game.Player) int {
	opponent := game.GetOpponentColor(player.Color)
	playerFrontier := 0
	opponentFrontier := 0

	// Direction vectors for checking neighboring cells
	dx := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy := []int{-1, 0, 1, -1, 1, -1, 0, 1}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if b[i][j] == player.Color {
				// Check if this piece is adjacent to any empty square
				for k := 0; k < 8; k++ {
					nx, ny := i+dx[k], j+dy[k]
					if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 && b[nx][ny] == game.Empty {
						playerFrontier++
						break // Count each piece only once
					}
				}
			} else if b[i][j] == opponent {
				// Same check for opponent pieces
				for k := 0; k < 8; k++ {
					nx, ny := i+dx[k], j+dy[k]
					if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 && b[nx][ny] == game.Empty {
						opponentFrontier++
						break // Count each piece only once
					}
				}
			}
		}
	}

	// Simple difference, but inverted (fewer frontier discs is better)
	return opponentFrontier - playerFrontier
}

// Evaluate computes the frontier score
func (e *FrontierEvaluation) Evaluate(g game.Game, b game.Board, player game.Player) int {
	return e.rawEvaluate(b, player)
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
