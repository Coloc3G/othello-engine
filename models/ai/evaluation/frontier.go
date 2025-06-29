package evaluation

import (
	"math/bits"

	"github.com/Coloc3G/othello-engine/models/game"
)

// FrontierEvaluation évalue le nombre de pièces frontalières (adjacentes à des cases vides)
// Ces pièces sont généralement vulnérables et peuvent être retournées
type FrontierEvaluation struct{}

func NewFrontierEvaluation() *FrontierEvaluation {
	return &FrontierEvaluation{}
}

func (e *FrontierEvaluation) Evaluate(b game.BitBoard) int16 {
	pec := PrecomputeEvaluationBitBoard(b)
	return e.PECEvaluate(b, pec)
}

func (e *FrontierEvaluation) PECEvaluate(b game.BitBoard, pec PreEvaluationComputation) int16 {
	var whiteFrontier, blackFrontier int16

	// Get all pieces for both players
	whitePieces := b.WhitePieces
	blackPieces := b.BlackPieces
	emptySquares := ^(whitePieces | blackPieces)

	// Helper function to get adjacent squares with proper boundary checks
	getAdjacent := func(pieces uint64) uint64 {
		adjacent := uint64(0)

		// North (up 8)
		adjacent |= (pieces >> 8)
		// South (down 8)
		adjacent |= (pieces << 8)
		// East (right 1, avoid wrapping from rightmost to leftmost)
		adjacent |= ((pieces & 0xFEFEFEFEFEFEFEFE) >> 1)
		// West (left 1, avoid wrapping from leftmost to rightmost)
		adjacent |= ((pieces & 0x7F7F7F7F7F7F7F7F) << 1)
		// NorthEast (up 8, right 1)
		adjacent |= ((pieces & 0xFEFEFEFEFEFEFEFE) >> 9)
		// NorthWest (up 8, left 1)
		adjacent |= ((pieces & 0x7F7F7F7F7F7F7F7F) >> 7)
		// SouthEast (down 8, right 1)
		adjacent |= ((pieces & 0xFEFEFEFEFEFEFEFE) << 7)
		// SouthWest (down 8, left 1)
		adjacent |= ((pieces & 0x7F7F7F7F7F7F7F7F) << 9)

		return adjacent
	}

	// Find frontier pieces: pieces that are adjacent to empty squares
	emptyAdjacent := getAdjacent(emptySquares)
	whiteFrontierMask := whitePieces & emptyAdjacent
	blackFrontierMask := blackPieces & emptyAdjacent

	// Count bits
	whiteFrontier = int16(bits.OnesCount64(whiteFrontierMask))
	blackFrontier = int16(bits.OnesCount64(blackFrontierMask))
	return blackFrontier - whiteFrontier
}
