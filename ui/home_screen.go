package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// HomeScreen represents the home/entry screen of the application
type HomeScreen struct {
	ui            *UI
	face          font.Face
	buttonBounds  [2][4]int // Two buttons: [0] for Player vs AI, [1] for AI vs AI
	buttonHovered int       // -1: none, 0: Player vs AI, 1: AI vs AI
}

// NewHomeScreen creates a new home screen
func NewHomeScreen(ui *UI) *HomeScreen {
	return &HomeScreen{
		ui:            ui,
		face:          basicfont.Face7x13,
		buttonHovered: -1,
	}
}

// Layout implements the Screen interface
func (s *HomeScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input on the home screen
func (s *HomeScreen) Update() error {
	screenWidth, screenHeight := ebiten.WindowSize()

	// Define button dimensions
	buttonWidth := 250
	buttonHeight := 50
	buttonSpacing := 30

	// Calculate button positions
	firstButtonY := screenHeight/2 + 20
	secondButtonY := firstButtonY + buttonHeight + buttonSpacing

	// Update button bounds
	s.buttonBounds[0] = [4]int{
		(screenWidth - buttonWidth) / 2,
		firstButtonY,
		buttonWidth,
		buttonHeight,
	}

	s.buttonBounds[1] = [4]int{
		(screenWidth - buttonWidth) / 2,
		secondButtonY,
		buttonWidth,
		buttonHeight,
	}

	// Check if mouse is over any button
	mouseX, mouseY := ebiten.CursorPosition()
	s.buttonHovered = -1

	for i := 0; i < 2; i++ {
		bounds := s.buttonBounds[i]
		if mouseX >= bounds[0] && mouseX < bounds[0]+bounds[2] &&
			mouseY >= bounds[1] && mouseY < bounds[1]+bounds[3] {
			s.buttonHovered = i
			break
		}
	}

	// Handle button clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch s.buttonHovered {
		case 0:
			// Player vs AI button clicked - go to AI selection screen
			s.ui.SwitchToAISelectionScreen()
		case 1:
			// AI vs AI button clicked - go to dual AI selection screen
			s.ui.SwitchToDualAISelectionScreen()
		}
	}

	return nil
}

// Draw renders the home screen
func (s *HomeScreen) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Fill background
	screen.Fill(ColorBackground)

	// Draw title
	title := "Othello Game"
	titleFace := s.face
	titleBounds, _ := font.BoundString(titleFace, title)
	titleX := (screenWidth - (titleBounds.Max.X - titleBounds.Min.X).Ceil()) / 2
	text.Draw(screen, title, titleFace, titleX, screenHeight/4, color.White)

	// Draw buttons
	buttonTexts := []string{"Player vs AI", "AI vs AI"}

	for i, buttonText := range buttonTexts {
		bounds := s.buttonBounds[i]

		// Draw button background
		buttonColor := color.RGBA{0, 100, 0, 255}
		if s.buttonHovered == i {
			buttonColor = color.RGBA{0, 150, 0, 255}
		}

		ebitenutil.DrawRect(screen,
			float64(bounds[0]),
			float64(bounds[1]),
			float64(bounds[2]),
			float64(bounds[3]),
			buttonColor)

		// Draw button text
		btnBounds := text.BoundString(s.face, buttonText)
		btnTextX := bounds[0] + (bounds[2]-btnBounds.Dx())/2
		btnTextY := bounds[1] + (bounds[3]+btnBounds.Dy())/2
		text.Draw(screen, buttonText, s.face, btnTextX, btnTextY, color.White)
	}
}
