package game

import "fmt"

// DisplayBoard prints a representation of the board to the console
// DisplayBoard prints the current state of the Othello board to the console.
// The function displays column letters (A-H) across the top and row numbers (1-8) along the left side,
// using chess-style notation. Empty cells are shown as "·", black pieces as "●", and white pieces as "○".
//
// Parameters:
//   - board: The Board to display
func (g *Game) DisplayBoard(board Board) {
	// Display column letters (A-H)
	fmt.Print("   ")
	for i := 0; i < 8; i++ {
		fmt.Printf(" %c", 'A'+i)
	}
	fmt.Println()

	// Display board with row numbers (1-8)
	for i := range board {
		fmt.Printf("%d |", i+1) // Row numbers start from 1
		for j := range board[i] {
			switch board[i][j] {
			case Empty:
				fmt.Print(" ·")
			case Black:
				fmt.Print(" ○")
			case White:
				fmt.Print(" ●")
			}
		}
		fmt.Println()
	}
}

// GetNewBoardAfterMove returns a new game state after applying a move
func GetNewBoardAfterMove(board Board, pos Position, player Piece) (Board, bool) {
	return ApplyMoveToBoard(board, player, pos)
}

// GetNewBitBoardAfterMove returns a new bitboard state after applying a move
func GetNewBitBoardAfterMove(bb BitBoard, pos Position, player Piece) (BitBoard, bool) {
	return ApplyMoveToBitBoard(bb, player, pos)
}

// GetNewBoardAfterMoveMethod is a method wrapper for GetNewBoardAfterMove
func (g *Game) GetNewBoardAfterMoveMethod(pos Position) (Board, bool) {
	return GetNewBoardAfterMove(g.Board, pos, g.CurrentPlayer.Color)
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

func CountPiecesBitBoard(bb BitBoard) (int, int) {
	// Use optimized popcount for maximum performance
	return popcount(bb.BlackPieces), popcount(bb.WhitePieces)
}

// popcount returns the number of set bits using Brian Kernighan's algorithm
// This is much faster than the bit-by-bit approach
func popcount(x uint64) int {
	count := 0
	for x != 0 {
		count++
		x &= x - 1 // Clear the lowest set bit
	}
	return count
}

// CountPiecesMethod is a method wrapper for CountPieces
func (g *Game) CountPiecesMethod() (int, int) {
	return CountPieces(g.Board)
}
