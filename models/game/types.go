package game

type Piece int

const (
	Empty Piece = 0
	White Piece = 1
	Black Piece = 2
)

type Position struct {
	Row int
	Col int
}

type Board [8][8]Piece

type BitBoard struct {
	BlackPieces uint64
	WhitePieces uint64
}

type Player struct {
	Color Piece
	Name  string
}

// Game represents the state of an Othello game.
// It contains the game board, the two players, the current player's turn,
// and the number of moves that have been made in the game.
// This struct is used to maintain the complete state of a game session.
type Game struct {
	Board         Board
	Players       [2]Player
	CurrentPlayer Player
	NbMoves       int
	History       []Position
}
