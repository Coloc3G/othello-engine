package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

func main() {
	fmt.Println("=== Testing Board and Bitboard Function Matching ===")

	// Test cases: various board states including random ones
	testCases := []struct {
		name  string
		board game.Board
	}{
		// {
		// 	name:  "Initial Game State",
		// 	board: getInitialBoard(),
		// },
		// {
		// 	name:  "Mid-game State 1",
		// 	board: getMidGameBoard1(),
		// },
		// {
		// 	name:  "Mid-game State 2",
		// 	board: getMidGameBoard2(),
		// },
		// {
		// 	name:  "Near End Game",
		// 	board: getNearEndGameBoard(),
		// },
		// {
		// 	name:  "Empty Board",
		// 	board: getEmptyBoard(),
		// },
	}

	// Add random board test cases
	// numRandomBoards := 100
	// for i := 0; i < numRandomBoards; i++ {
	// 	testCases = append(testCases, struct {
	// 		name  string
	// 		board game.Board
	// 	}{
	// 		name:  fmt.Sprintf("Random Board %d", i+1),
	// 		board: generateRandomBoard(),
	// 	})
	// }

	g := game.NewGame("Black", "White")
	applyPosition(g, utils.AlgebraicToPositions("c4c3d3c5f6e2c6d6b5c7b4e3b7e6f4b6a6f5f3g4g5a8"))
	testCases = append(testCases, struct {
		name  string
		board game.Board
	}{
		name:  "Perf",
		board: g.Board,
	})

	// Track results for summary
	var results []TestResult

	// Test each case
	for _, tc := range testCases {
		result := testBoardBitboardMatch(tc.board)
		result.TestCase = tc.name
		results = append(results, result)
	}

	// Print summary
	printSummary(results)
}

func applyPosition(g *game.Game, pos []game.Position) (err error) {
	for _, move := range pos {
		if !game.IsValidMove(g.Board, g.CurrentPlayer.Color, move) {
			return fmt.Errorf("invalid move %s for player %s", utils.PositionToAlgebraic(move), g.CurrentPlayer.Name)
		}
		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer.Color)
		g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		}
	}
	return
}

type TestResult struct {
	TestCase                string
	ValidMovesMatch         bool
	ApplyMoveMatch          bool
	IsGameFinishedMatch     bool
	CountPiecesMatch        bool
	BitboardConversionMatch bool
	EvaluationMatch         bool
}

func testBoardBitboardMatch(board game.Board) TestResult {
	// Convert board to bitboard for comparison
	bitboard := utils.BoardToBits(board)

	result := TestResult{}

	// Test bitboard conversion match
	convertedBack := utils.BitsToBoard(bitboard)
	result.BitboardConversionMatch = reflect.DeepEqual(board, convertedBack)

	// Test ValidMoves match
	result.ValidMovesMatch = testValidMovesMatch(board, bitboard)

	// Test ApplyMove match
	result.ApplyMoveMatch = testApplyMoveMatch(board, bitboard)

	// Test game state functions match
	result.IsGameFinishedMatch = testIsGameFinishedMatch(board, bitboard)
	result.CountPiecesMatch = testCountPiecesMatch(board, bitboard)

	// Test evaluation functions match
	result.EvaluationMatch = testEvaluationMatch(board, bitboard)

	return result
}

func testValidMovesMatch(board game.Board, bitboard game.BitBoard) bool {
	colors := []game.Piece{game.Black, game.White}

	for _, color := range colors {
		moves := game.ValidMoves(board, color)
		bitboardMoves := game.ValidMovesBitBoard(bitboard, color)
		fmt.Printf("Valid moves for color %d:\nBoard: %v\nBitboard: %v\n", color, utils.PositionsToAlgebraic(moves), utils.PositionsToAlgebraic(bitboardMoves))
		if len(moves)+len(bitboardMoves) == 0 {
			continue
		}
		// Sort moves for consistent comparison
		sort.Slice(moves, func(i, j int) bool {
			if moves[i].Row == moves[j].Row {
				return moves[i].Col < moves[j].Col
			}
			return moves[i].Row < moves[j].Row
		})

		sort.Slice(bitboardMoves, func(i, j int) bool {
			if bitboardMoves[i].Row == bitboardMoves[j].Row {
				return bitboardMoves[i].Col < bitboardMoves[j].Col
			}
			return bitboardMoves[i].Row < bitboardMoves[j].Row
		})

		if !reflect.DeepEqual(moves, bitboardMoves) {
			utils.PrintBoard(board)
			fmt.Printf("Valid moves mismatch for color %d:\nBoard: %v\nBitboard: %v\n", color, moves, bitboardMoves)
			return false
		}
	}
	return true
}

func testApplyMoveMatch(board game.Board, bitboard game.BitBoard) bool {
	colors := []game.Piece{game.Black, game.White}

	for _, color := range colors {
		validMoves := game.ValidMoves(board, color)
		if len(validMoves) == 0 {
			continue
		}

		// Test first valid move
		move := validMoves[0]

		newBoard, success1 := game.ApplyMoveToBoard(board, color, move)
		newBitboard, success2 := game.ApplyMoveToBitBoard(bitboard, color, move)

		if success1 != success2 {
			return false
		}

		if success1 && success2 {
			convertedBoard := utils.BoardToBits(newBoard)
			if !reflect.DeepEqual(convertedBoard, newBitboard) {
				return false
			}
		}
	}
	return true
}

func testIsGameFinishedMatch(board game.Board, bitboard game.BitBoard) bool {
	result1 := game.IsGameFinished(board)
	result2 := game.IsGameFinishedBitBoard(bitboard)
	return result1 == result2
}

func testCountPiecesMatch(board game.Board, bitboard game.BitBoard) bool {
	black1, white1 := game.CountPieces(board)
	black2, white2 := game.CountPiecesBitBoard(bitboard)
	return black1 == black2 && white1 == white2
}

func testEvaluationMatch(board game.Board, bitboard game.BitBoard) bool {
	pec := evaluation.PrecomputeEvaluation(board)
	pecBit := evaluation.PrecomputeEvaluationBitBoard(bitboard)

	if pec.IsGameOver != pecBit.IsGameOver ||
		pec.BlackPieces != pecBit.BlackPieces ||
		pec.WhitePieces != pecBit.WhitePieces ||
		len(pec.BlackValidMoves) != len(pecBit.BlackValidMoves) ||
		len(pec.WhiteValidMoves) != len(pecBit.WhiteValidMoves) {

		fmt.Println("Evaluation mismatch:")
		fmt.Printf("IsGameOver: %v vs %v\n", pec.IsGameOver, pecBit.IsGameOver)
		fmt.Printf("BlackPieces: %d vs %d\n", pec.BlackPieces, pecBit.BlackPieces)
		fmt.Printf("WhitePieces: %d vs %d\n", pec.WhitePieces, pecBit.WhitePieces)
		fmt.Printf("BlackValidMoves: %v vs %v\n", pec.BlackValidMoves, pecBit.BlackValidMoves)
		fmt.Printf("WhiteValidMoves: %v vs %v\n", pec.WhiteValidMoves, pecBit.WhiteValidMoves)
		return false
	}

	// Sort valid moves for consistent comparison
	sort.Slice(pec.BlackValidMoves, func(i, j int) bool {
		if pec.BlackValidMoves[i].Row == pec.BlackValidMoves[j].Row {
			return pec.BlackValidMoves[i].Col < pec.BlackValidMoves[j].Col
		}
		return pec.BlackValidMoves[i].Row < pec.BlackValidMoves[j].Row
	})

	sort.Slice(pecBit.BlackValidMoves, func(i, j int) bool {
		if pecBit.BlackValidMoves[i].Row == pecBit.BlackValidMoves[j].Row {
			return pecBit.BlackValidMoves[i].Col < pecBit.BlackValidMoves[j].Col
		}
		return pecBit.BlackValidMoves[i].Row < pecBit.BlackValidMoves[j].Row
	})

	return (reflect.DeepEqual(pec.BlackValidMoves, pecBit.BlackValidMoves) || len(pec.BlackValidMoves)+len(pecBit.BlackValidMoves) == 0) &&
		(reflect.DeepEqual(pec.WhiteValidMoves, pecBit.WhiteValidMoves) || len(pec.WhiteValidMoves)+len(pecBit.WhiteValidMoves) == 0)

}

func printSummary(results []TestResult) {
	fmt.Println("=== SUMMARY ===")
	fmt.Printf("%-20s | %-12s | %-12s | %-15s | %-12s | %-20s | %-15s\n",
		"Test Case", "ValidMoves", "ApplyMove", "IsGameFinished", "CountPieces", "BitboardConversion", "Evaluation")
	fmt.Println(strings.Repeat("-", 115))

	totalTests := len(results)
	passCount := map[string]int{
		"ValidMoves":         0,
		"ApplyMove":          0,
		"IsGameFinished":     0,
		"CountPieces":        0,
		"BitboardConversion": 0,
		"Evaluation":         0,
	}

	for _, result := range results {
		validMovesStatus := "FAIL"
		if result.ValidMovesMatch {
			validMovesStatus = "PASS"
			passCount["ValidMoves"]++
		}

		applyMoveStatus := "FAIL"
		if result.ApplyMoveMatch {
			applyMoveStatus = "PASS"
			passCount["ApplyMove"]++
		}

		gameFinishedStatus := "FAIL"
		if result.IsGameFinishedMatch {
			gameFinishedStatus = "PASS"
			passCount["IsGameFinished"]++
		}

		countPiecesStatus := "FAIL"
		if result.CountPiecesMatch {
			countPiecesStatus = "PASS"
			passCount["CountPieces"]++
		}

		conversionStatus := "FAIL"
		if result.BitboardConversionMatch {
			conversionStatus = "PASS"
			passCount["BitboardConversion"]++
		}

		evaluationStatus := "FAIL"
		if result.EvaluationMatch {
			evaluationStatus = "PASS"
			passCount["Evaluation"]++
		}

		fmt.Printf("%-20s | %-12s | %-12s | %-15s | %-12s | %-20s | %-15s\n",
			result.TestCase, validMovesStatus, applyMoveStatus, gameFinishedStatus, countPiecesStatus, conversionStatus, evaluationStatus)
	}

	fmt.Println(strings.Repeat("-", 115))
	fmt.Printf("%-20s | %-12s | %-12s | %-15s | %-12s | %-20s | %-15s\n",
		"TOTALS",
		fmt.Sprintf("%d/%d", passCount["ValidMoves"], totalTests),
		fmt.Sprintf("%d/%d", passCount["ApplyMove"], totalTests),
		fmt.Sprintf("%d/%d", passCount["IsGameFinished"], totalTests),
		fmt.Sprintf("%d/%d", passCount["CountPieces"], totalTests),
		fmt.Sprintf("%d/%d", passCount["BitboardConversion"], totalTests),
		fmt.Sprintf("%d/%d", passCount["Evaluation"], totalTests))
}

// Helper functions to create test boards
func getInitialBoard() game.Board {
	var board game.Board
	// Standard Othello starting position
	board[3][3] = game.White
	board[3][4] = game.Black
	board[4][3] = game.Black
	board[4][4] = game.White
	return board
}

func getMidGameBoard1() game.Board {
	var board game.Board
	// Create a mid-game scenario
	board[2][3] = game.Black
	board[3][2] = game.Black
	board[3][3] = game.Black
	board[3][4] = game.Black
	board[3][5] = game.Black
	board[4][3] = game.White
	board[4][4] = game.White
	board[5][3] = game.White
	return board
}

func getMidGameBoard2() game.Board {
	var board game.Board
	// Another mid-game scenario
	board[1][1] = game.Black
	board[2][2] = game.Black
	board[3][3] = game.Black
	board[4][4] = game.White
	board[5][5] = game.White
	board[6][6] = game.White
	board[3][4] = game.White
	board[4][3] = game.Black
	return board
}

func getNearEndGameBoard() game.Board {
	var board game.Board
	// Fill most of the board
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if (i+j)%2 == 0 {
				board[i][j] = game.Black
			} else if i < 6 && j < 6 {
				board[i][j] = game.White
			}
		}
	}
	return board
}

func getEmptyBoard() game.Board {
	var board game.Board
	// All cells are Empty (default value)
	return board
}

// generateRandomBoard creates a random board state for testing
func generateRandomBoard() game.Board {
	var board game.Board

	// Random density of pieces (between 5% and 80% of the board)
	totalCells := 64
	minPieces := totalCells * 5 / 100  // 5%
	maxPieces := totalCells * 80 / 100 // 80%
	numPieces := rand.Intn(maxPieces-minPieces+1) + minPieces

	// Create a slice of all possible positions
	positions := make([]struct{ row, col int }, 0, totalCells)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			positions = append(positions, struct{ row, col int }{i, j})
		}
	}

	// Shuffle positions
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Place pieces randomly
	for i := 0; i < numPieces; i++ {
		pos := positions[i]
		// Randomly choose between Black and White (roughly equal distribution)
		if rand.Float32() < 0.5 {
			board[pos.row][pos.col] = game.Black
		} else {
			board[pos.row][pos.col] = game.White
		}
	}

	return board
}
