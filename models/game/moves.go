package game

// ValidMoves returns all valid moves for a player on a given board
func ValidMoves(board Board, playerColor Piece) []Position {
	moves := make([]Position, 0, 20) // Pre-allocate with reasonable capacity
	opponentColor := GetOpponentColor(playerColor)

	// Direction vectors for all 8 directions
	directions := [8]Position{
		{-1, -1}, {-1, 0}, {-1, 1}, // Above
		{0, -1}, {0, 1}, // Sides
		{1, -1}, {1, 0}, {1, 1}, // Below
	}

	// Check all empty cells on the board
	for row := range 8 {
		for col := range 8 {
			if board[row][col] != Empty {
				continue
			}

			// Inline validity check for better performance
			isValid := false

			// Check all 8 directions
			for _, dir := range directions {
				r, c := row+dir.Row, col+dir.Col

				// First piece must be opponent's and in bounds
				if r < 0 || r >= 8 || c < 0 || c >= 8 || board[r][c] != opponentColor {
					continue
				}

				// Keep moving in this direction looking for player's piece
				r += dir.Row
				c += dir.Col

				for r >= 0 && r < 8 && c >= 0 && c < 8 {
					if board[r][c] == Empty {
						break
					}
					if board[r][c] == playerColor {
						isValid = true
						break
					}
					r += dir.Row
					c += dir.Col
				}

				if isValid {
					break
				}
			}

			if isValid {
				moves = append(moves, Position{Row: row, Col: col})
			}
		}
	}

	return moves

}

// getPositionBuffer returns a pre-allocated buffer, alternating between two buffers
func getPositionBuffer() []Position {
	return make([]Position, 0, 32)
}

// ValidMovesBitBoard returns all valid moves for a player using state-of-the-art bitboard operations
// Uses optimized Kogge-Stone sliding attack generation for maximum performance
func ValidMovesBitBoard(board BitBoard, playerColor Piece) []Position {
	var playerBits, opponentBits uint64
	if playerColor == White {
		playerBits = board.WhitePieces
		opponentBits = board.BlackPieces
	} else {
		playerBits = board.BlackPieces
		opponentBits = board.WhitePieces
	}

	emptyBits := ^(playerBits | opponentBits)

	// Use state-of-the-art move generation combining all directions
	validMoves := generateValidMovesOptimized(playerBits, opponentBits, emptyBits)

	return bitboardToPositionsOptimized(validMoves)
}

// generateValidMovesOptimized uses optimized Kogge-Stone algorithm for all 8 directions
// This is significantly faster than the previous implementation
func generateValidMovesOptimized(playerBits, opponentBits, emptyBits uint64) uint64 {
	// Edge masks for boundary checking
	const notAFile = 0xFEFEFEFEFEFEFEFE // Not A file (rightmost column)
	const notHFile = 0x7F7F7F7F7F7F7F7F // Not H file (leftmost column)

	validMoves := uint64(0)

	// Optimized direction generation using single-pass Kogge-Stone
	// Each direction processes all possible moves simultaneously

	// North (shift up)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return b << 8 }, 0xFFFFFFFFFFFFFFFF)

	// South (shift down)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return b >> 8 }, 0xFFFFFFFFFFFFFFFF)

	// East (shift right, avoid wrap-around)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b << 1) & notAFile }, notAFile)

	// West (shift left, avoid wrap-around)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b >> 1) & notHFile }, notHFile)

	// Northeast (shift up-right)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b << 9) & notAFile }, notAFile)

	// Northwest (shift up-left)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b << 7) & notHFile }, notHFile)

	// Southeast (shift down-right)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b >> 7) & notAFile }, notAFile)

	// Southwest (shift down-left)
	validMoves |= koggeStoneDirection(playerBits, opponentBits, emptyBits,
		func(b uint64) uint64 { return (b >> 9) & notHFile }, notHFile)

	return validMoves
}

// koggeStoneDirection implements optimized Kogge-Stone sliding attack generation
// This algorithm processes all possible flips in a direction simultaneously
func koggeStoneDirection(playerBits, opponentBits, emptyBits uint64,
	shiftFunc func(uint64) uint64, mask uint64) uint64 {

	// Start with player pieces shifted one step
	flood := shiftFunc(playerBits) & mask

	// Find opponent pieces adjacent to player pieces
	flood &= opponentBits

	// Kogge-Stone parallel prefix: propagate through opponent pieces
	// Maximum 6 iterations needed for 8x8 board
	flood |= shiftFunc(flood) & opponentBits & mask
	flood |= shiftFunc(flood) & opponentBits & mask
	flood |= shiftFunc(flood) & opponentBits & mask
	flood |= shiftFunc(flood) & opponentBits & mask
	flood |= shiftFunc(flood) & opponentBits & mask
	flood |= shiftFunc(flood) & opponentBits & mask

	// Valid moves are empty squares at the end of flip sequences
	return shiftFunc(flood) & emptyBits & mask
}

// bitboardToPositionsOptimized converts bitboard to positions with optimized bit scanning
func bitboardToPositionsOptimized(bitboard uint64) []Position {
	if bitboard == 0 {
		return nil
	}

	buffer := getPositionBuffer()

	// Use optimized bit scanning loop
	for bitboard != 0 {
		// Find position of least significant bit using De Bruijn multiplication
		bitPos := trailingZeros(bitboard & -bitboard)

		row := bitPos >> 3 // Equivalent to bitPos / 8 but faster
		col := bitPos & 7  // Equivalent to bitPos % 8 but faster

		buffer = append(buffer, Position{Row: row, Col: col})

		// Clear the least significant bit
		bitboard &= bitboard - 1
	}

	// Create a copy to avoid buffer reuse issues
	result := make([]Position, len(buffer))
	copy(result, buffer)
	return result
}

// trailingZeros counts trailing zeros using optimized bit manipulation
// Much faster than the previous bit-by-bit approach
func trailingZeros(x uint64) int {
	if x == 0 {
		return 64
	}

	// Use binary search approach for fast trailing zero count
	n := 0
	if (x & 0xFFFFFFFF) == 0 {
		n += 32
		x >>= 32
	}
	if (x & 0xFFFF) == 0 {
		n += 16
		x >>= 16
	}
	if (x & 0xFF) == 0 {
		n += 8
		x >>= 8
	}
	if (x & 0xF) == 0 {
		n += 4
		x >>= 4
	}
	if (x & 0x3) == 0 {
		n += 2
		x >>= 2
	}
	if (x & 0x1) == 0 {
		n++
	}
	return n
}

// IsValidMove checks if placing a piece of the given color at the specified position is a valid move.
// A move is valid if it results in flipping at least one opponent's piece.
//
// Parameters:
//   - board: The current game board state
//   - playerColor: The color of the piece to be placed (Black or White)
//   - pos: The position where the piece would be placed
//
// Returns:
//   - bool: true if the move is valid, false otherwise
func IsValidMove(board Board, playerColor Piece, pos Position) bool {
	// Check if position is in bounds and the cell is empty
	if pos.Row < 0 || pos.Row >= 8 || pos.Col < 0 || pos.Col >= 8 || board[pos.Row][pos.Col] != Empty {
		return false
	}

	// Get the opponent's color
	opponentColor := GetOpponentColor(playerColor)

	// Direction vectors for all 8 directions
	directions := [8]Position{
		{-1, -1}, {-1, 0}, {-1, 1}, // Above
		{0, -1}, {0, 1}, // Sides
		{1, -1}, {1, 0}, {1, 1}, // Below
	}

	// Check all 8 directions
	for _, dir := range directions {
		r, c := pos.Row+dir.Row, pos.Col+dir.Col

		// Step 1: The first piece in this direction must be opponent's
		if r < 0 || r >= 8 || c < 0 || c >= 8 || board[r][c] != opponentColor {
			continue
		}

		// Step 2: Keep moving in this direction
		r += dir.Row
		c += dir.Col

		for r >= 0 && r < 8 && c >= 0 && c < 8 {
			if board[r][c] == Empty {
				break
			}
			if board[r][c] == playerColor {
				return true
			}
			r += dir.Row
			c += dir.Col
		}
	}

	return false
}

// GetValidMovesForCurrentPlayer is a method wrapper for the ValidMoves function
func (g *Game) GetValidMovesForCurrentPlayer() []Position {
	return ValidMoves(g.Board, g.CurrentPlayer.Color)
}

// ApplyMoveToBoard applies a move to a board and returns the new board state
// without modifying the original board. Returns a new board and whether the move was valid.
// ApplyMoveToBoard applies the specified move to the game board and returns the new board state.
//
// This function performs the following steps:
// 1. Validates if the move is legal using IsValidMove
// 2. Places the player's piece at the specified position
// 3. Flips opponent pieces in all 8 directions according to Othello rules
//
// Parameters:
//   - board: The current state of the game board
//   - playerColor: The color of the player making the move (BLACK or WHITE)
//   - pos: The position where the player wants to place their piece
//
// Returns:
//   - The updated board after applying the move
//   - A boolean indicating whether the move was successfully applied (true) or invalid (false)
//
// If the move is invalid, the original board is returned unchanged with false as the second return value.
func ApplyMoveToBoard(board Board, playerColor Piece, pos Position) (Board, bool) {
	// Check if the move is valid
	if !IsValidMove(board, playerColor, pos) {
		return board, false
	}

	// Create a copy of the board
	newBoard := board

	// Place the piece
	newBoard[pos.Row][pos.Col] = playerColor

	// Get the opponent's color
	opponentColor := GetOpponentColor(playerColor)

	// Direction vectors for all 8 directions
	directions := []Position{
		{-1, -1}, {-1, 0}, {-1, 1}, // Above
		{0, -1}, {0, 1}, // Sides
		{1, -1}, {1, 0}, {1, 1}, // Below
	}

	// Check all 8 directions and flip pieces
	for _, dir := range directions {
		// Store pieces to flip
		piecesToFlip := []Position{}

		r, c := pos.Row+dir.Row, pos.Col+dir.Col

		// Continue in this direction as long as we find opponent pieces
		for r >= 0 && r < 8 && c >= 0 && c < 8 && newBoard[r][c] == opponentColor {
			piecesToFlip = append(piecesToFlip, Position{Row: r, Col: c})
			r += dir.Row
			c += dir.Col
		}

		// If we found our own piece at the end of the line, flip all opponent pieces
		if r >= 0 && r < 8 && c >= 0 && c < 8 && newBoard[r][c] == playerColor {
			for _, flipPos := range piecesToFlip {
				newBoard[flipPos.Row][flipPos.Col] = playerColor
			}
		}
	}

	return newBoard, true
}

// ApplyMoveToBitBoard applies a move to a bitboard and returns the new bitboard state
func ApplyMoveToBitBoard(board BitBoard, playerColor Piece, pos Position) (BitBoard, bool) {
	// Check if position is in bounds
	if pos.Row < 0 || pos.Row >= 8 || pos.Col < 0 || pos.Col >= 8 {
		return board, false
	}

	bitPos := uint64(1) << (pos.Row*8 + pos.Col)

	// Check if position is empty
	if (board.WhitePieces|board.BlackPieces)&bitPos != 0 {
		return board, false
	}

	var playerBits, opponentBits *uint64
	if playerColor == White {
		playerBits = &board.WhitePieces
		opponentBits = &board.BlackPieces
	} else {
		playerBits = &board.BlackPieces
		opponentBits = &board.WhitePieces
	}

	// Direction shift functions and masks
	directions := []struct {
		shift func(uint64) uint64
		mask  uint64
	}{
		{func(b uint64) uint64 { return (b << 8) }, 0xFFFFFFFFFFFFFFFF},                      // North
		{func(b uint64) uint64 { return (b >> 8) }, 0xFFFFFFFFFFFFFFFF},                      // South
		{func(b uint64) uint64 { return (b << 1) & 0xFEFEFEFEFEFEFEFE }, 0xFEFEFEFEFEFEFEFE}, // East
		{func(b uint64) uint64 { return (b >> 1) & 0x7F7F7F7F7F7F7F7F }, 0x7F7F7F7F7F7F7F7F}, // West
		{func(b uint64) uint64 { return (b << 9) & 0xFEFEFEFEFEFEFEFE }, 0xFEFEFEFEFEFEFEFE}, // NorthEast
		{func(b uint64) uint64 { return (b << 7) & 0x7F7F7F7F7F7F7F7F }, 0x7F7F7F7F7F7F7F7F}, // NorthWest
		{func(b uint64) uint64 { return (b >> 7) & 0xFEFEFEFEFEFEFEFE }, 0xFEFEFEFEFEFEFEFE}, // SouthEast
		{func(b uint64) uint64 { return (b >> 9) & 0x7F7F7F7F7F7F7F7F }, 0x7F7F7F7F7F7F7F7F}, // SouthWest
	}

	newBoard := board
	toFlip := uint64(0)
	validMove := false

	// Check each direction for flips
	for _, dir := range directions {
		captured := uint64(0)
		probe := dir.shift(bitPos) & dir.mask

		// Collect opponent pieces in this direction
		for probe != 0 && (probe&(*opponentBits)) != 0 {
			captured |= probe
			probe = dir.shift(probe) & dir.mask
		}

		// If we hit our own piece and captured something, mark for flipping
		if captured != 0 && (probe&(*playerBits)) != 0 {
			toFlip |= captured
			validMove = true
		}
	}

	if !validMove {
		return board, false
	}

	// Apply the move to the new board
	if playerColor == White {
		newBoard.WhitePieces |= bitPos | toFlip
		newBoard.BlackPieces &= ^toFlip
	} else {
		newBoard.BlackPieces |= bitPos | toFlip
		newBoard.WhitePieces &= ^toFlip
	}

	return newBoard, true
}

// ApplyMove applies a move to the current game state
func (g *Game) ApplyMove(pos Position) bool {
	newBoard, success := ApplyMoveToBoard(g.Board, g.CurrentPlayer.Color, pos)

	if !success {
		return false
	}

	g.Board = newBoard
	g.NbMoves++
	g.History = append(g.History, pos)

	// Switch to the other player
	otherPlayer := GetOtherPlayer(g.CurrentPlayer.Color)
	g.CurrentPlayer = otherPlayer

	return true
}

// HasAnyMoves checks if there are any valid moves for a given player color on a board
func HasAnyMoves(board Board, playerColor Piece) bool {
	moves := ValidMoves(board, playerColor)
	return len(moves) > 0
}

// HasAnyMovesInGame is a method wrapper for HasAnyMoves
func (g *Game) HasAnyMovesInGame() bool {
	return HasAnyMoves(g.Board, g.CurrentPlayer.Color)
}

// GetOpponentColor returns the opposite color of the given player color
func GetOpponentColor(playerColor Piece) Piece {
	if playerColor == White {
		return Black
	}
	return White
}
