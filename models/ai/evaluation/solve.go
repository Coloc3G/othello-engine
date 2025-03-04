package evaluation

import "github.com/Coloc3G/othello-engine/models/game"

func Solve(g game.Game, player game.Player, depth int, eval Evaluation) game.Position {
	bestScore := -1 << 31
	var bestMove game.Position
	for _, move := range game.GetAllPossibleMoves(g.Board, player) {
		newNode := game.GetNewBoardAfterMove(g.Board, move, player)
		childScore := MMAB(g, newNode, player, depth-1, false, -1<<31, 1<<31-1, eval)
		if childScore > bestScore {
			bestScore = childScore
			bestMove = move
		}
	}
	return bestMove
}

func MMAB(g game.Game, node game.Board, player game.Player, depth int, max bool, alpha, beta int, eval Evaluation) int {
	if depth == 0 || game.IsGameFinished(node) {
		return eval.Evaluate(node, player)
	}
	oplayer := game.GetOtherPlayer(player)
	if (max && !game.HasAnyMoves(node, player)) || (!max && !game.HasAnyMoves(node, oplayer)) {
		return MMAB(g, node, player, depth-1, !max, alpha, beta, eval)
	}
	var score int
	if max {
		score = -1 << 31
		for _, move := range game.GetAllPossibleMoves(node, player) {
			newNode := game.GetNewBoardAfterMove(node, move, player)
			childScore := MMAB(g, newNode, player, depth-1, false, alpha, beta, eval)
			if childScore > score {
				score = childScore
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
	} else {
		score = 1<<31 - 1
		for _, move := range game.GetAllPossibleMoves(node, oplayer) {
			newNode := game.GetNewBoardAfterMove(node, move, oplayer)
			childScore := MMAB(g, newNode, player, depth-1, true, alpha, beta, eval)
			if childScore < score {
				score = childScore
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
	}
	return score
}
