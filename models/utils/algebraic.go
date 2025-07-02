package utils

import "github.com/Coloc3G/othello-engine/models/game"

// AlgebraicToPosition converts an algebraic position (like "c4") to a Position
func AlgebraicToPosition(algebraic string) game.Position {
	if len(algebraic) < 2 {
		return game.Position{Row: -1, Col: -1} // Invalid position
	}

	col := int8(algebraic[0] - 'a')
	row := int8(algebraic[1] - '1')

	// Check boundaries
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return game.Position{Row: -1, Col: -1} // Invalid position
	}

	return game.Position{Row: row, Col: col}
}

// PositionToAlgebraic converts a Position to algebraic notation (like "c4")
func PositionToAlgebraic(pos game.Position) string {
	if pos.Row < 0 || pos.Row > 7 || pos.Col < 0 || pos.Col > 7 {
		return "invalid" // Invalid position
	}

	col := 'a' + byte(pos.Col)
	row := '1' + byte(pos.Row)

	return string([]byte{col, row})
}

func PositionsToAlgebraic(positions []game.Position) string {
	algebraic := ""
	for _, position := range positions {
		algebraic += PositionToAlgebraic(position)
	}
	return algebraic
}

func AlgebraicToPositions(algebraic string) []game.Position {
	positions := make([]game.Position, len(algebraic)/2)
	for i := 0; i < len(algebraic); i += 2 {
		positions[i/2] = AlgebraicToPosition(algebraic[i : i+2])
	}
	return positions
}
