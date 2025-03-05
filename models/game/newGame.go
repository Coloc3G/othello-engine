package game

// NewGame creates and initializes a new Othello game.
// It sets up the board with the standard initial position where four pieces
// are placed in the center of the board (two black and two white in a diagonal pattern).
// The function also initializes both players, sets Black as the first player to move,
// and initializes the move counter to zero.
// Returns a ready-to-play Game instance.
func NewGame() Game {
	g := Game{}

	// Initialize the board
	for i := range g.Board {
		for j := range g.Board[i] {
			g.Board[i][j] = Empty
		}
	}

	// Pieces in the initial position
	g.Board[3][3] = White
	g.Board[3][4] = Black
	g.Board[4][3] = Black
	g.Board[4][4] = White

	// Initialize both players
	g.Players[0] = Player{Color: Black, Name: "AI"}
	g.Players[1] = Player{Color: White, Name: "AI"}

	// Set Black as the first player
	g.CurrentPlayer = g.Players[0]
	g.NbMoves = 0

	return g
}
