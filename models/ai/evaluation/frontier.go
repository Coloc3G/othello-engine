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
	pec := precomputeEvaluation(g, b, player)
	return e.PECEvaluate(g, b, pec)
}

func (e *FrontierEvaluation) PECEvaluate(g game.Game, b game.Board, pec PreEvaluationComputation) int {
	playerFrontier, opponentFrontier := countFrontierDiscs(b, pec.Player.Color, pec.Opponent.Color)
	return opponentFrontier - playerFrontier
}

// countFrontierDiscs compte le nombre de pièces frontalières pour les deux joueurs en une seule passe
func countFrontierDiscs(b game.Board, playerColor, opponentColor game.Piece) (int, int) {
	playerFrontier := 0
	opponentFrontier := 0
	directions := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}

	for i := range ai.BoardSize {
		for j := range ai.BoardSize {
			currentPiece := b[i][j]
			if currentPiece == playerColor || currentPiece == opponentColor {
				for _, dir := range directions {
					dx, dy := dir[0], dir[1]
					r, c := i+dx, j+dy

					if r >= 0 && r < ai.BoardSize && c >= 0 && c < ai.BoardSize && b[r][c] == game.Empty {
						if currentPiece == playerColor {
							playerFrontier++
						} else {
							opponentFrontier++
						}
						break
					}
				}
			}
		}
	}

	return playerFrontier, opponentFrontier
}
