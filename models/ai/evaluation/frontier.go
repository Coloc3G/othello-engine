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
	// Get all pieces for both players
	whitePieces := b.WhitePieces
	blackPieces := b.BlackPieces
	emptySquares := ^(whitePieces | blackPieces)

	// Precomputed masks for boundary checks (more efficient than runtime computation)
	const (
		notLeftEdge   = 0xFEFEFEFEFEFEFEFE
		notRightEdge  = 0x7F7F7F7F7F7F7F7F
		notTopEdge    = 0x00FFFFFFFFFFFFFF
		notBottomEdge = 0xFFFFFFFFFFFFFF00
	)

	// Calculate adjacent squares using optimized bit operations
	adjacent := emptySquares>>8 | emptySquares<<8 | // North & South
		(emptySquares&notLeftEdge)>>1 | (emptySquares&notRightEdge)<<1 | // East & West
		(emptySquares&notLeftEdge&notTopEdge)>>9 | (emptySquares&notRightEdge&notTopEdge)>>7 | // NE & NW
		(emptySquares&notLeftEdge&notBottomEdge)<<7 | (emptySquares&notRightEdge&notBottomEdge)<<9 // SE & SW

	// Find frontier pieces: pieces that are adjacent to empty squares
	whiteFrontierMask := whitePieces & adjacent
	blackFrontierMask := blackPieces & adjacent

	// Count bits using native popcount
	whiteFrontier := int16(bits.OnesCount64(whiteFrontierMask))
	blackFrontier := int16(bits.OnesCount64(blackFrontierMask))

	return blackFrontier - whiteFrontier
}
