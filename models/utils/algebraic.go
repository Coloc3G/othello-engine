package utils

import "github.com/Coloc3G/othello-engine/models/game"

func PositionToAlgebraic(positions game.Position) string {
	return string(rune('a'+positions.Row)) + string(rune('1'+positions.Col))
}

func AlgebraicToPosition(algebraic string) game.Position {
	return game.Position{Row: int(algebraic[0] - 'a'), Col: int(algebraic[1] - '1')}
}

func PositionsToAlgebraic(positions []game.Position) []string {
	algebraic := make([]string, len(positions))
	for i, position := range positions {
		algebraic[i] = PositionToAlgebraic(position)
	}
	return algebraic
}

func AlgebraicToPositions(algebraic []string) []game.Position {
	positions := make([]game.Position, len(algebraic))
	for i, alg := range algebraic {
		positions[i] = AlgebraicToPosition(alg)
	}
	return positions
}
