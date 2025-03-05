package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/Coloc3G/othello-engine/models/game"
)

// ResultScreen shows the game results
type ResultScreen struct {
	ui   *UI
	face font.Face
}

// NewResultScreen creates a new result screen
func NewResultScreen(ui *UI) *ResultScreen {
	return &ResultScreen{
		ui:   ui,
		face: basicfont.Face7x13,
	}
}

// Layout implements ebiten.Game
func (s *ResultScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input on the result screen
func (s *ResultScreen) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.ui.NewGame()
	}

	return nil
}

// Draw renders the result screen
func (s *ResultScreen) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(ColorBackground)

	// Calculate scores
	blackCount, whiteCount := game.CountPieces(s.ui.game.Board)

	// Determine winner
	var winner string
	if blackCount > whiteCount {
		winner = "Black wins!"
	} else if whiteCount > blackCount {
		winner = "White wins!"
	} else {
		winner = "It's a tie!"
	}

	// Draw title
	title := "Game Over"
	titleBounds := text.BoundString(s.face, title)
	titleX := (screen.Bounds().Dx() - titleBounds.Dx()) / 2
	text.Draw(screen, title, s.face, titleX, 100, color.White)

	// Draw score
	scoreText := fmt.Sprintf("Final Score - Black: %d  White: %d", blackCount, whiteCount)
	scoreBounds := text.BoundString(s.face, scoreText)
	scoreX := (screen.Bounds().Dx() - scoreBounds.Dx()) / 2
	text.Draw(screen, scoreText, s.face, scoreX, 130, color.White)

	// Draw winner
	winnerBounds := text.BoundString(s.face, winner)
	winnerX := (screen.Bounds().Dx() - winnerBounds.Dx()) / 2
	text.Draw(screen, winner, s.face, winnerX, 160, color.White)

	// Draw instructions
	instructions := "Click anywhere to play again"
	instBounds := text.BoundString(s.face, instructions)
	instX := (screen.Bounds().Dx() - instBounds.Dx()) / 2
	text.Draw(screen, instructions, s.face, instX, 200, color.White)
}
