package utils

import "github.com/Coloc3G/othello-engine/models/game"

func PositionToAlgebraic(positions game.Position) string {
	return string(rune('a'+positions.Col)) + string(rune('1'+positions.Row))
}

func AlgebraicToPosition(algebraic string) game.Position {
	return game.Position{Col: int(algebraic[0] - 'a'), Row: int(algebraic[1] - '1')}
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
