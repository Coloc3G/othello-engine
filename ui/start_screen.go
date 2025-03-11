package ui

import (
	"image/color"

	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// StartScreen represents the game's start screen
type StartScreen struct {
	ui           *UI
	face         font.Face
	playerNames  [2]string
	activeInput  int // -1: none, 0: player1, 1: player2
	cursorPos    int
	buttonHover  bool
	buttonBounds [4]int // x, y, width, height
}

// NewStartScreen creates a new start screen
func NewStartScreen(ui *UI) *StartScreen {
	return &StartScreen{
		ui:          ui,
		face:        basicfont.Face7x13,
		playerNames: [2]string{"Player 1", "AI"},
		activeInput: -1,
	}
}

// Layout implements ebiten.Game
func (s *StartScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input for the start screen
func (s *StartScreen) Update() error {
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

	// Handle mouse clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check button click
		if s.buttonHover {
			// Start the game
			player1 := s.playerNames[0]
			if player1 == "" {
				player1 = "Player 1"
			}

			player2 := s.playerNames[1]
			if player2 == "" {
				player2 = "AI"
			}

			s.ui.StartGame(player1, player2)
			return nil
		}

		// Check input field clicks
		y1 := 200
		y2 := 280
		inputWidth := 300
		inputX := (screenWidth - inputWidth) / 2

		if mouseX >= inputX && mouseX < inputX+inputWidth {
			if mouseY >= y1 && mouseY < y1+30 {
				s.activeInput = 0
				s.cursorPos = len(s.playerNames[0])
			} else if mouseY >= y2 && mouseY < y2+30 {
				s.activeInput = 1
				s.cursorPos = len(s.playerNames[1])
			} else {
				s.activeInput = -1
			}
		} else {
			s.activeInput = -1
		}
	}

	// Handle keyboard input for active text field
	if s.activeInput >= 0 {
		// Get typed runes
		for _, r := range ebiten.InputChars() {
			name := s.playerNames[s.activeInput]
			if s.cursorPos > len(name) {
				s.cursorPos = len(name)
			}

			// Insert character at cursor position
			s.playerNames[s.activeInput] = name[:s.cursorPos] + string(r) + name[s.cursorPos:]
			s.cursorPos++
		}

		// Handle backspace
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			name := s.playerNames[s.activeInput]
			if s.cursorPos > 0 && len(name) > 0 {
				s.playerNames[s.activeInput] = name[:s.cursorPos-1] + name[s.cursorPos:]
				s.cursorPos--
			}
		}

		// Handle arrow keys
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && s.cursorPos > 0 {
			s.cursorPos--
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) && s.cursorPos < len(s.playerNames[s.activeInput]) {
			s.cursorPos++
		}

		// Handle enter/tab to move between fields
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
			s.activeInput = (s.activeInput + 1) % 2
			s.cursorPos = len(s.playerNames[s.activeInput])
		}
	}

	return nil
}

// Draw renders the start screen
func (s *StartScreen) Draw(screen *ebiten.Image) {
	screenWidth, _ := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Fill background
	screen.Fill(ColorBackground)

	// Draw title
	title := "Othello"
	titleBounds, _ := font.BoundString(s.face, title)
	titleX := (screenWidth - (titleBounds.Max.X - titleBounds.Min.X).Ceil()) / 2
	text.Draw(screen, title, s.face, titleX, 100, color.White)

	// Draw input fields
	inputWidth := 300
	inputX := (screenWidth - inputWidth) / 2

	// Player 1 field
	text.Draw(screen, "Player 1 (Black):", s.face, inputX, 180, color.White)
	vector.DrawFilledRect(screen, float32(inputX), 190, float32(inputWidth), 30, color.RGBA{60, 60, 60, 255}, false)
	text.Draw(screen, s.playerNames[0], s.face, inputX+5, 210, color.White)

	// Draw cursor for player 1 field
	if s.activeInput == 0 {
		cursorX := inputX + 5
		if s.cursorPos > 0 {
			cursorX += text.BoundString(s.face, s.playerNames[0][:s.cursorPos]).Dx()
		}
		ebitenutil.DrawLine(screen, float64(cursorX), 195, float64(cursorX), 215, color.White)
	}

	// Player 2 field
	text.Draw(screen, "Player 2 (White):", s.face, inputX, 260, color.White)
	ebitenutil.DrawRect(screen, float64(inputX), 270, float64(inputWidth), 30, color.RGBA{60, 60, 60, 255})
	text.Draw(screen, s.playerNames[1], s.face, inputX+5, 290, color.White)

	// Draw cursor for player 2 field
	if s.activeInput == 1 {
		cursorX := inputX + 5
		if s.cursorPos > 0 {
			cursorX += text.BoundString(s.face, s.playerNames[1][:s.cursorPos]).Dx()
		}
		ebitenutil.DrawLine(screen, float64(cursorX), 275, float64(cursorX), 295, color.White)
	}

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
	buttonText := "Start Game"
	btnBounds := text.BoundString(s.face, buttonText)
	btnTextX := s.buttonBounds[0] + (s.buttonBounds[2]-btnBounds.Dx())/2
	btnTextY := s.buttonBounds[1] + (s.buttonBounds[3]+btnBounds.Dy())/2
	text.Draw(screen, buttonText, s.face, btnTextX, btnTextY, color.White)
}

// StartGame initializes a new game with the specified player names
func (s *UI) StartGame(player1, player2 string) {
	// Create new game
	s.game = game.NewGame(player1, player2)

	// Switch to game screen
	s.currentScreen = s.gameScreen
}
