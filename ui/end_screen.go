package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/Coloc3G/othello-engine/models/game"
)

// EndScreen represents the game over screen
type EndScreen struct {
	ui           *UI
	face         font.Face
	buttonHover  bool
	buttonBounds [4]int // x, y, width, height
}

// NewEndScreen creates a new end screen
func NewEndScreen(ui *UI) *EndScreen {
	return &EndScreen{
		ui:   ui,
		face: basicfont.Face7x13,
	}
}

// Layout implements ebiten.Game
func (s *EndScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input on the end screen
func (s *EndScreen) Update() error {
	// Update button bounds
	screenWidth, screenHeight := ebiten.WindowSize()
	buttonWidth := 200
	buttonHeight := 40
	s.buttonBounds = [4]int{
		(screenWidth - buttonWidth) / 2,
		screenHeight - 150,
		buttonWidth,
		buttonHeight,
	}

	// Check if mouse is over button
	mouseX, mouseY := ebiten.CursorPosition()
	s.buttonHover = mouseX >= s.buttonBounds[0] &&
		mouseX < s.buttonBounds[0]+s.buttonBounds[2] &&
		mouseY >= s.buttonBounds[1] &&
		mouseY < s.buttonBounds[1]+s.buttonBounds[3]

	// Handle button click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && s.buttonHover {
		s.ui.SwitchToHomeScreen()
	}

	return nil
}

// Draw renders the end screen
func (s *EndScreen) Draw(screen *ebiten.Image) {
	screenWidth, _ := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Fill background
	screen.Fill(ColorBackground)

	// Get game results
	blackCount, whiteCount := game.CountPieces(s.ui.game.Board)
	var resultText string
	var winnerName string

	if blackCount > whiteCount {
		resultText = "Black Wins!"
		for _, player := range s.ui.game.Players {
			if player.Color == game.Black {
				winnerName = player.Name
				break
			}
		}
	} else if whiteCount > blackCount {
		resultText = "White Wins!"
		for _, player := range s.ui.game.Players {
			if player.Color == game.White {
				winnerName = player.Name
				break
			}
		}
	} else {
		resultText = "It's a Tie!"
		winnerName = "Nobody"
	}

	// Draw title
	title := "Game Over"
	titleBounds := text.BoundString(s.face, title)
	titleX := (screenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, title, s.face, titleX, 100, color.White)

	// Draw result
	resBounds := text.BoundString(s.face, resultText)
	resX := (screenWidth - resBounds.Dx()) / 2
	text.Draw(screen, resultText, s.face, resX, 140, color.White)

	// Draw winner
	winnerText := fmt.Sprintf("%s wins!", winnerName)
	winBounds := text.BoundString(s.face, winnerText)
	winX := (screenWidth - winBounds.Dx()) / 2
	text.Draw(screen, winnerText, s.face, winX, 170, color.White)

	// Draw score
	scoreText := fmt.Sprintf("Final Score: Black %d - %d White", blackCount, whiteCount)
	scoreBounds := text.BoundString(s.face, scoreText)
	scoreX := (screenWidth - scoreBounds.Dx()) / 2
	text.Draw(screen, scoreText, s.face, scoreX, 200, color.White)

	// Draw button
	buttonColor := color.RGBA{0, 100, 0, 255}
	if s.buttonHover {
		buttonColor = color.RGBA{0, 150, 0, 255}
	}

	ebitenutil.DrawRect(screen,
		float64(s.buttonBounds[0]),
		float64(s.buttonBounds[1]),
		float64(s.buttonBounds[2]),
		float64(s.buttonBounds[3]),
		buttonColor)

	// Draw button text
	buttonText := "Main Menu"
	btnBounds := text.BoundString(s.face, buttonText)
	btnTextX := s.buttonBounds[0] + (s.buttonBounds[2]-btnBounds.Dx())/2
	btnTextY := s.buttonBounds[1] + (s.buttonBounds[3]+btnBounds.Dy())/2
	text.Draw(screen, buttonText, s.face, btnTextX, btnTextY, color.White)
}
