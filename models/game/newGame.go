package game

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
	g.Players[0] = Player{Color: Black, Name: "Black"}
	g.Players[1] = Player{Color: White, Name: "White"}

	// Set Black as the first player
	g.CurrentPlayer = g.Players[0]
	g.NbMoves = 0

	return g
}
