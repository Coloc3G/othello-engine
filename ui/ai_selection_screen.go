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

// AISelectionScreen represents the screen for selecting an AI opponent
type AISelectionScreen struct {
	ui               *UI
	face             font.Face
	selectedAI       int      // -1: none, 0: V1, 1: V2
	aiButtonBounds   [][4]int // Bounds for each AI button
	playButtonBounds [4]int   // Bounds for play button
	backButtonBounds [4]int   // Bounds for back button
	buttonHovered    int      // -1: none, 0-n: AI buttons, n+1: play, n+2: back
	initialized      bool     // Whether the screen has been initialized
}

// NewAISelectionScreen creates a new AI selection screen
func NewAISelectionScreen(ui *UI) *AISelectionScreen {
	// Initialize with 2 AI options
	aiButtonBounds := make([][4]int, 2)

	return &AISelectionScreen{
		ui:             ui,
		face:           basicfont.Face7x13,
		selectedAI:     -1,
		buttonHovered:  -1,
		aiButtonBounds: aiButtonBounds,
		initialized:    false,
	}
}

// Layout implements the Screen interface
func (s *AISelectionScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input on the AI selection screen
func (s *AISelectionScreen) Update() error {
	screenWidth, screenHeight := ebiten.WindowSize()

	// Define button dimensions
	aiButtonWidth := 100
	aiButtonHeight := 40
	aiButtonSpacing := 20
	playButtonWidth := 150
	playButtonHeight := 50
	backButtonWidth := 100
	backButtonHeight := 40

	// Calculate positions
	aiButtonY := screenHeight / 2
	playButtonY := screenHeight - 120
	backButtonY := screenHeight - 120

	// Update AI button bounds - we have 2 AIs (V1, V2)
	numAIOptions := 2
	aiStartX := (screenWidth - ((aiButtonWidth * numAIOptions) + (aiButtonSpacing * (numAIOptions - 1)))) / 2

	s.aiButtonBounds = make([][4]int, numAIOptions)
	for i := 0; i < numAIOptions; i++ {
		s.aiButtonBounds[i] = [4]int{
			aiStartX + (i * (aiButtonWidth + aiButtonSpacing)),
			aiButtonY,
			aiButtonWidth,
			aiButtonHeight,
		}
	}

	// Mark as initialized
	s.initialized = true

	// Play button bounds
	s.playButtonBounds = [4]int{
		(screenWidth + aiButtonSpacing) / 2,
		playButtonY,
		playButtonWidth,
		playButtonHeight,
	}

	// Back button bounds
	s.backButtonBounds = [4]int{
		(screenWidth-playButtonWidth-aiButtonSpacing)/2 - backButtonWidth,
		backButtonY,
		backButtonWidth,
		backButtonHeight,
	}

	// Check mouse position
	mouseX, mouseY := ebiten.CursorPosition()
	s.buttonHovered = -1

	// Check AI buttons
	for i := 0; i < numAIOptions; i++ {
		bounds := s.aiButtonBounds[i]
		if mouseX >= bounds[0] && mouseX < bounds[0]+bounds[2] &&
			mouseY >= bounds[1] && mouseY < bounds[1]+bounds[3] {
			s.buttonHovered = i
			break
		}
	}

	// Check play button
	if mouseX >= s.playButtonBounds[0] && mouseX < s.playButtonBounds[0]+s.playButtonBounds[2] &&
		mouseY >= s.playButtonBounds[1] && mouseY < s.playButtonBounds[1]+s.playButtonBounds[3] {
		s.buttonHovered = numAIOptions
	}

	// Check back button
	if mouseX >= s.backButtonBounds[0] && mouseX < s.backButtonBounds[0]+s.backButtonBounds[2] &&
		mouseY >= s.backButtonBounds[1] && mouseY < s.backButtonBounds[1]+s.backButtonBounds[3] {
		s.buttonHovered = numAIOptions + 1
	}

	// Handle clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch s.buttonHovered {
		case 0, 1: // AI selection buttons
			s.selectedAI = s.buttonHovered
		case 2: // Play button
			if s.selectedAI >= 0 {
				// Start game with selected AI
				s.ui.StartPlayerVsAIGame(s.selectedAI)
			}
		case 3: // Back button
			s.ui.SwitchToHomeScreen()
		}
	}

	return nil
}

// Draw renders the AI selection screen
func (s *AISelectionScreen) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Fill background
	screen.Fill(ColorBackground)

	// Draw title
	title := "Select AI Level"
	titleBounds := text.BoundString(s.face, title)
	titleX := (screenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, title, s.face, titleX, screenHeight/4, color.White)

	// Check if initialized before drawing buttons
	if !s.initialized || len(s.aiButtonBounds) == 0 {
		// Draw loading message or just return
		text.Draw(screen, "Loading...", s.face, screenWidth/2-30, screenHeight/2, color.White)
		return
	}

	// Draw AI buttons
	aiOptions := []string{"V1", "V2"}
	for i, optionText := range aiOptions {
		if i >= len(s.aiButtonBounds) {
			continue // Skip if index is out of bounds
		}

		bounds := s.aiButtonBounds[i]

		// Button background color
		var buttonColor color.RGBA
		if s.selectedAI == i {
			buttonColor = color.RGBA{0, 150, 0, 255} // Selected
		} else if s.buttonHovered == i {
			buttonColor = color.RGBA{0, 120, 0, 255} // Hovered
		} else {
			buttonColor = color.RGBA{0, 80, 0, 255} // Normal
		}

		// Draw button
		ebitenutil.DrawRect(screen,
			float64(bounds[0]),
			float64(bounds[1]),
			float64(bounds[2]),
			float64(bounds[3]),
			buttonColor)

		// Draw button text
		btnBounds := text.BoundString(s.face, optionText)
		btnTextX := bounds[0] + (bounds[2]-btnBounds.Dx())/2
		btnTextY := bounds[1] + (bounds[3]+btnBounds.Dy())/2
		text.Draw(screen, optionText, s.face, btnTextX, btnTextY, color.White)
	}

	// Draw play button (only if an AI is selected)
	buttonColor := color.RGBA{100, 100, 100, 255} // Disabled
	if s.selectedAI >= 0 {
		buttonColor = color.RGBA{0, 100, 0, 255} // Enabled
		if s.buttonHovered == 2 {
			buttonColor = color.RGBA{0, 150, 0, 255} // Hovered
		}
	}

	ebitenutil.DrawRect(screen,
		float64(s.playButtonBounds[0]),
		float64(s.playButtonBounds[1]),
		float64(s.playButtonBounds[2]),
		float64(s.playButtonBounds[3]),
		buttonColor)

	playText := "Play"
	btnBounds := text.BoundString(s.face, playText)
	btnTextX := s.playButtonBounds[0] + (s.playButtonBounds[2]-btnBounds.Dx())/2
	btnTextY := s.playButtonBounds[1] + (s.playButtonBounds[3]+btnBounds.Dy())/2
	text.Draw(screen, playText, s.face, btnTextX, btnTextY, color.White)

	// Draw back button
	backButtonColor := color.RGBA{100, 70, 70, 255}
	if s.buttonHovered == 3 {
		backButtonColor = color.RGBA{150, 70, 70, 255}
	}

	ebitenutil.DrawRect(screen,
		float64(s.backButtonBounds[0]),
		float64(s.backButtonBounds[1]),
		float64(s.backButtonBounds[2]),
		float64(s.backButtonBounds[3]),
		backButtonColor)

	backText := "Back"
	backBounds := text.BoundString(s.face, backText)
	backTextX := s.backButtonBounds[0] + (s.backButtonBounds[2]-backBounds.Dx())/2
	backTextY := s.backButtonBounds[1] + (s.backButtonBounds[3]+backBounds.Dy())/2
	text.Draw(screen, backText, s.face, backTextX, backTextY, color.White)
}
