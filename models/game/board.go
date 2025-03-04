package game

// DisplayBoard prints a representation of the board to the console
// DisplayBoard prints the current state of the Othello board to the console.
// The function displays column numbers (0-7) across the top and row numbers (0-7) along the left side.
// Empty cells are shown as "·", black pieces as "●", and white pieces as "○".
//
// Parameters:
//   - board: The Board to display

// GetNewBoardAfterMove returns a new game state after applying a move
func GetNewBoardAfterMove(board Board, pos Position, player Player) (Board, bool) {
	return ApplyMoveToBoard(board, player.Color, pos)
}

// GetNewBoardAfterMoveMethod is a method wrapper for GetNewBoardAfterMove
func (g *Game) GetNewBoardAfterMoveMethod(pos Position) (Board, bool) {
	return GetNewBoardAfterMove(g.Board, pos, g.CurrentPlayer)
}

// CountPieces counts the number of pieces of each color on the board
// Returns the count of black pieces and white pieces
func CountPieces(board Board) (int, int) {
	blackCount := 0
	whiteCount := 0

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

	return blackCount, whiteCount
}

// CountPiecesMethod is a method wrapper for CountPieces
func (g *Game) CountPiecesMethod() (int, int) {
	return CountPieces(g.Board)
}
