package game

// ValidMoves returns all valid moves for a player on a given board
func ValidMoves(board Board, playerColor Piece) []Position {
	moves := []Position{}

	// Check all empty cells on the board
	for row := range board {
		for col := range board[row] {
			if board[row][col] != Empty {
				continue
			}

			pos := Position{Row: row, Col: col}
			if IsValidMove(board, playerColor, pos) {
				moves = append(moves, pos)
			}
		}
	}

	return moves
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
	otherPlayer := GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
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
