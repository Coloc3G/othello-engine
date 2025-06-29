package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

func PrecomputeEvaluation(b game.Board) (pec PreEvaluationComputation) {
	black, white := game.CountPieces(b)
	pec.BlackPieces = int16(black)
	pec.WhitePieces = int16(white)

	pec.BlackValidMoves = game.ValidMoves(b, game.Black)
	pec.WhiteValidMoves = game.ValidMoves(b, game.White)

	if black+white == 64 || game.IsGameFinished(b) {
		pec.IsGameOver = true
	}

	return
}

func PrecomputeEvaluationBitBoard(b game.BitBoard) (pec PreEvaluationComputation) {
	// Use optimized piece counting
	black, white := game.CountPiecesBitBoard(b)
	pec.BlackPieces = int16(black)
	pec.WhitePieces = int16(white)

	// Fast path: if board is full, game is over
	if black+white == 64 {
		pec.IsGameOver = true
		// Initialize empty slices for consistency
		pec.BlackValidMoves = make([]game.Position, 0)
		pec.WhiteValidMoves = make([]game.Position, 0)
		return
	}

	// Use optimized move generation
	pec.BlackValidMoves = game.ValidMovesBitBoard(b, game.Black)
	pec.WhiteValidMoves = game.ValidMovesBitBoard(b, game.White)

	// Game is over if neither player has valid moves
	if len(pec.BlackValidMoves)+len(pec.WhiteValidMoves) == 0 {
		pec.IsGameOver = true
	}
	return
}
