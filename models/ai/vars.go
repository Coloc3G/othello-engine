package ai

// Game phase constants
const (
	EarlyGame = 0 // Moins de 20 pièces
	MidGame   = 1 // Entre 20 et 58 pièces
	LateGame  = 2 // Plus de 58 pièces

	BoardSize = 8
)

// Weights for stability map
var StabilityMap = [8][8]int16{
	{4, -3, 2, 2, 2, 2, -3, 4},
	{-3, -4, -1, -1, -1, -1, -4, -3},
	{2, -1, 1, 0, 0, 1, -1, 2},
	{2, -1, 0, 1, 1, 0, -1, 2},
	{2, -1, 0, 1, 1, 0, -1, 2},
	{2, -1, 1, 0, 0, 1, -1, 2},
	{-3, -4, -1, -1, -1, -1, -4, -3},
	{4, -3, 2, 2, 2, 2, -3, 4},
}
