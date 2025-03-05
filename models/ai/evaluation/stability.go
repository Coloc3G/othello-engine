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
	for i := 0; i < ai.BoardSize; i++ {
		for j := 0; j < ai.BoardSize; j++ {
			if b[i][j] == player.Color {
				playerScore += ai.StabilityMap[i][j]
			} else if b[i][j] == opponent.Color {
				opponentScore += ai.StabilityMap[i][j]
			}
		}
	}

	// Calcul des pièces stables (qui ne peuvent plus être retournées)
	playerStable := countStablePieces(b, player.Color)
	opponentStable := countStablePieces(b, opponent.Color)

	stabilityDiff := (playerScore - opponentScore) + 3*(playerStable-opponentStable)

	return stabilityDiff
}

// countStablePieces compte les pièces stables (coins et pièces connectées aux coins)
func countStablePieces(b game.Board, color game.Piece) int {
	stable := 0
	stableBoard := [8][8]bool{}

	// Les coins sont toujours stables
	corners := [][2]int{{0, 0}, {0, 7}, {7, 0}, {7, 7}}

	// Marquer les coins stables
	for _, corner := range corners {
		row, col := corner[0], corner[1]
		if b[row][col] == color {
			stableBoard[row][col] = true
			stable++
		}
	}

	// Propager la stabilité depuis les coins
	changed := true
	for changed {
		changed = false

		for i := 0; i < ai.BoardSize; i++ {
			for j := 0; j < ai.BoardSize; j++ {
				if b[i][j] == color && !stableBoard[i][j] {
					if isStable(b, stableBoard, i, j, color) {
						stableBoard[i][j] = true
						stable++
						changed = true
					}
				}
			}
		}
	}

	return stable
}

// isStable détermine si une pièce est stable (ne peut plus être retournée)
func isStable(b game.Board, stableBoard [8][8]bool, row, col int, color game.Piece) bool {
	// Une pièce est stable si elle est entourée de pièces stables ou de bords dans toutes les directions
	directions := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Horizontales et verticales
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1}, // Diagonales
	}

	for _, dir := range directions {
		dx, dy := dir[0], dir[1]
		r, c := row+dx, col+dy

		// Si on sort du plateau, cette direction est stable
		if r < 0 || r >= ai.BoardSize || c < 0 || c >= ai.BoardSize {
			continue
		}

		// Si la case adjacente n'est pas stable dans cette direction, la pièce n'est pas stable
		if b[r][c] != color || !stableBoard[r][c] {
			return false
		}
	}

	return true
}
