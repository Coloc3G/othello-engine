package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

func precomputeEvaluation(g game.Game, b game.Board, p game.Player) (pec PreEvaluationComputation) {
	pec.Player = p
	pec.Opponent = game.GetOtherPlayer(g.Players, p.Color)
	black, white := game.CountPieces(b)
	switch p.Color {
	case game.Black:
		pec.PlayerPieces = black
		pec.OpponentPieces = white
	case game.White:
		pec.PlayerPieces = white
		pec.OpponentPieces = black
	}

	pec.PlayerValidMoves = game.ValidMoves(b, p.Color)
	pec.OpponentValidMoves = game.ValidMoves(b, pec.Opponent.Color)

	if black+white == 64 || game.IsGameFinished(b) {
		pec.IsGameOver = true
	}

	return
}
