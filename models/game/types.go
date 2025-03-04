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

type Player struct {
	Color Piece
	Name  string
}

type Game struct {
	Board         Board
	CurrentPlayer Player
	NbMoves       int
}
