package game

// ValidMoves returns all valid moves for a player on a given board
func ValidMoves(board Board, player Player) []Position {
	moves := []Position{}

	// Check all empty cells on the board
	for row := range board {
		for col := range board[row] {
			if board[row][col] != Empty {
				continue
			}

			pos := Position{Row: row, Col: col}
			if IsValidMove(board, player.Color, pos) {
				moves = append(moves, pos)
			}
		}
	}

	return moves
}

// IsValidMove checks if a move is valid for a player on a given board
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
//
// The function checks:
//  1. If the position is within bounds and the cell is empty
//  2. For each of the 8 directions:
//     - Checks if there's an opponent's piece adjacent
//     - Follows that direction to find a player's piece
//     - If found, the move is valid as it would flip opponent's pieces
func IsValidMove(board Board, playerColor Piece, pos Position) bool {
	// Check if position is in bounds and the cell is empty
	if pos.Row < 0 || pos.Row >= 8 || pos.Col < 0 || pos.Col >= 8 || board[pos.Row][pos.Col] != Empty {
		return false
	}

	// Get the opponent's color
	opponentColor := White
	if playerColor == White {
		opponentColor = Black
	}

	// Direction vectors for all 8 directions
	directions := []Position{
		{-1, -1}, {-1, 0}, {-1, 1}, // Above
		{0, -1}, {0, 1}, // Sides
		{1, -1}, {1, 0}, {1, 1}, // Below
	}

	validMove := false

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
		foundPlayerPiece := false

		for r >= 0 && r < 8 && c >= 0 && c < 8 {
			if board[r][c] == Empty {
				break
			}
			if board[r][c] == playerColor {
				foundPlayerPiece = true
				break
			}
			r += dir.Row
			c += dir.Col
		}

		if foundPlayerPiece {
			validMove = true
			break
		}
	}

	return validMove
}

// GetValidMovesForCurrentPlayer returns all valid moves for the current player
func (g *Game) GetValidMovesForCurrentPlayer() []Position {
	return ValidMoves(g.Board, g.CurrentPlayer)
}
