package game

// IsGameFinished checks if the game is over on a given board by determining if any valid moves remain
// The game is finished when neither player has any valid moves
func IsGameFinished(board Board) bool {
	// Check if black player has valid moves
	blackMoves := ValidMoves(board, Black)

	// Check if white player has valid moves
	whiteMoves := ValidMoves(board, White)

	// Game is finished if neither player has valid moves
	return len(blackMoves) == 0 && len(whiteMoves) == 0
}

func IsGameFinishedBitBoard(bb BitBoard) bool {
	// Check if black player has valid moves
	blackMoves := ValidMovesBitBoard(bb, Black)

	// Check if white player has valid moves
	whiteMoves := ValidMovesBitBoard(bb, White)

	// Game is finished if neither player has valid moves
	return len(blackMoves) == 0 && len(whiteMoves) == 0
}

// IsGameFinishedMethod is a method wrapper for IsGameFinished
func (g *Game) IsGameFinishedMethod() bool {
	return IsGameFinished(g.Board)
}

// GetWinner returns the winner of the game (color with more pieces)
// If it's a tie, returns Empty
// GetWinner determines the winner of an Othello game based on the current board.
// It counts the number of black and white pieces on the board and returns:
// - Black if there are more black pieces
// - White if there are more white pieces
// - Empty if there's a tie (equal number of black and white pieces)
//
// Parameters:
//   - board: The game board to evaluate
//
// Returns:
//   - Piece: The winning piece color (Black, White) or Empty in case of a tie
func GetWinner(board Board) Piece {
	blackCount := 0
	whiteCount := 0

	// Count pieces
	for row := range board {
		for col := range board[row] {
			switch board[row][col] {
			case Black:
				blackCount++
			case White:
				whiteCount++
			}
		}
	}

	// Determine winner
	if blackCount > whiteCount {
		return Black
	} else if whiteCount > blackCount {
		return White
	}

	// Tie
	return Empty
}

// GetWinnerMethod is a method wrapper for GetWinner
func (g *Game) GetWinnerMethod() Piece {
	return GetWinner(g.Board)
}
