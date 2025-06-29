package utils

import (
	"fmt"

	"github.com/Coloc3G/othello-engine/models/game"
)

func HashBoard(b game.Board) string {
	return HashBitBoard(BoardToBits(b))
}

func HashBitBoard(bb game.BitBoard) string {
	var buf [32]byte
	const hexChars = "0123456789abcdef"

	// Convert BlackPieces to hex
	black := bb.BlackPieces
	for i := 15; i >= 0; i-- {
		buf[i] = hexChars[black&0xf]
		black >>= 4
	}

	// Convert WhitePieces to hex
	white := bb.WhitePieces
	for i := 31; i >= 16; i-- {
		buf[i] = hexChars[white&0xf]
		white >>= 4
	}

	return string(buf[:])
}

func BoardToBits(b game.Board) game.BitBoard {
	var black, white uint64
	for i := range 8 {
		for j := range 8 {
			switch b[i][j] {
			case game.Black:
				black |= 1 << (i*8 + j)
			case game.White:
				white |= 1 << (i*8 + j)
			}
		}
	}
	return game.BitBoard{
		BlackPieces: black,
		WhitePieces: white,
	}
}

func BitsToBoard(bb game.BitBoard) game.Board {
	board := game.Board{}
	for i := range 8 {
		for j := range 8 {
			pos := i*8 + j
			if bb.BlackPieces&(1<<pos) != 0 {
				board[i][j] = game.Black
			} else if bb.WhitePieces&(1<<pos) != 0 {
				board[i][j] = game.White
			} else {
				board[i][j] = game.Empty
			}
		}
	}
	return board
}

func PrintBoard(b game.Board) {
	for i := range b {
		for j := range b[i] {
			switch b[i][j] {
			case game.Empty:
				fmt.Print(" ·")
			case game.Black:
				fmt.Print(" ○")
			case game.White:
				fmt.Print(" ●")
			}
		}
		fmt.Println()
	}
}

func PrintBitBoard(bb game.BitBoard) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			pos := i*8 + j
			if bb.BlackPieces&(1<<pos) != 0 {
				fmt.Print(" ○")
			} else if bb.WhitePieces&(1<<pos) != 0 {
				fmt.Print(" ●")
			} else {
				fmt.Print(" ·")
			}
		}
		fmt.Println()
	}
}
