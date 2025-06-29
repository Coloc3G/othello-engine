package game

// GetOtherPlayer returns the Player object with the opposite color
// GetOtherPlayer returns the opponent player given the current player's color.
// It takes an array of two players and the current player's color as arguments,
// then returns the player whose color differs from the current player's color.
// This function is useful for alternating turns in the game.
func GetOtherPlayer(currentColor Piece) Player {
	if currentColor == White {
		return Player{Color: Black}
	}
	return Player{Color: White}
}

// GetOtherPlayerMethod is a method wrapper for GetOtherPlayer
func (g *Game) GetOtherPlayerMethod() Player {
	return GetOtherPlayer(g.CurrentPlayer.Color)
}
