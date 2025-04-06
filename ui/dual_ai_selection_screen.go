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
)

// DualAISelectionScreen represents the screen for selecting two AI players
type DualAISelectionScreen struct {
	ui                 *UI
	face               font.Face
	selectedAIs        [2]int      // Selected AI for each player: -1 for none, 0 for V1, 1 for V2
	aiButtonBounds     [2][][4]int // Bounds for each AI button [player][button]
	playButtonBounds   [4]int      // Bounds for play button
	backButtonBounds   [4]int      // Bounds for back button
	buttonHovered      int         // -1: none, positive: specific button
	currentHoverPlayer int         // Which player's buttons are being hovered (-1 for none, 0 for first, 1 for second)
	currentHoverButton int         // Which button in that player's row (-1 for none, 0+ for button index)
	initialized        bool        // Whether the screen has been initialized
}

// NewDualAISelectionScreen creates a new dual AI selection screen
func NewDualAISelectionScreen(ui *UI) *DualAISelectionScreen {
	// Initialize with 2 AI options per player
	aiButtonBounds := [2][][4]int{
		make([][4]int, 2),
		make([][4]int, 2),
	}

	return &DualAISelectionScreen{
		ui:                 ui,
		face:               basicfont.Face7x13,
		selectedAIs:        [2]int{-1, -1},
		aiButtonBounds:     aiButtonBounds,
		buttonHovered:      -1,
		currentHoverPlayer: -1,
		currentHoverButton: -1,
		initialized:        false,
	}
}

// Layout implements the Screen interface
func (s *DualAISelectionScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update handles input on the dual AI selection screen
func (s *DualAISelectionScreen) Update() error {
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
	firstRowY := screenHeight/2 - 30
	secondRowY := screenHeight/2 + 50
	playButtonY := screenHeight - 120
	backButtonY := screenHeight - 120

	// Update AI button bounds - we have 2 AIs (V1, V2) per player
	numAIOptions := 2
	aiStartX := (screenWidth - ((aiButtonWidth * numAIOptions) + (aiButtonSpacing * (numAIOptions - 1)))) / 2

	// Initialize button bounds for both players
	// First player's AI buttons
	for i := 0; i < numAIOptions; i++ {
		s.aiButtonBounds[0][i] = [4]int{
			aiStartX + (i * (aiButtonWidth + aiButtonSpacing)),
			firstRowY,
			aiButtonWidth,
			aiButtonHeight,
		}
	}

	// Second player's AI buttons
	for i := 0; i < numAIOptions; i++ {
		s.aiButtonBounds[1][i] = [4]int{
			aiStartX + (i * (aiButtonWidth + aiButtonSpacing)),
			secondRowY,
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

	// Reset hover state
	s.currentHoverPlayer = -1
	s.currentHoverButton = -1
	s.buttonHovered = -1

	// Check mouse position
	mouseX, mouseY := ebiten.CursorPosition()

	// Check first player's AI buttons
	for i := 0; i < numAIOptions; i++ {
		bounds := s.aiButtonBounds[0][i]
		if mouseX >= bounds[0] && mouseX < bounds[0]+bounds[2] &&
			mouseY >= bounds[1] && mouseY < bounds[1]+bounds[3] {
			s.currentHoverPlayer = 0
			s.currentHoverButton = i
			s.buttonHovered = i
			break
		}
	}

	// Check second player's AI buttons
	if s.buttonHovered == -1 {
		for i := 0; i < numAIOptions; i++ {
			bounds := s.aiButtonBounds[1][i]
			if mouseX >= bounds[0] && mouseX < bounds[0]+bounds[2] &&
				mouseY >= bounds[1] && mouseY < bounds[1]+bounds[3] {
				s.currentHoverPlayer = 1
				s.currentHoverButton = i
				s.buttonHovered = numAIOptions + i
				break
			}
		}
	}

	// Check play button
	if mouseX >= s.playButtonBounds[0] && mouseX < s.playButtonBounds[0]+s.playButtonBounds[2] &&
		mouseY >= s.playButtonBounds[1] && mouseY < s.playButtonBounds[1]+s.playButtonBounds[3] {
		s.buttonHovered = 2 * numAIOptions
	}

	// Check back button
	if mouseX >= s.backButtonBounds[0] && mouseX < s.backButtonBounds[0]+s.backButtonBounds[2] &&
		mouseY >= s.backButtonBounds[1] && mouseY < s.backButtonBounds[1]+s.backButtonBounds[3] {
		s.buttonHovered = 2*numAIOptions + 1
	}

	// Handle clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if s.currentHoverPlayer >= 0 && s.currentHoverButton >= 0 {
			// AI selection button clicked
			s.selectedAIs[s.currentHoverPlayer] = s.currentHoverButton
		} else if s.buttonHovered == 2*numAIOptions {
			// Play button clicked
			if s.selectedAIs[0] >= 0 && s.selectedAIs[1] >= 0 {
				// Start AI vs AI game with selected AIs
				s.ui.StartAIVsAIGame(s.selectedAIs[0], s.selectedAIs[1])
			}
		} else if s.buttonHovered == 2*numAIOptions+1 {
			// Back button clicked
			s.ui.SwitchToHomeScreen()
		}
	}

	return nil
}

// Draw renders the dual AI selection screen
func (s *DualAISelectionScreen) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Fill background
	screen.Fill(ColorBackground)

	// Draw title
	title := "Select Two AI Players"
	titleBounds := text.BoundString(s.face, title)
	titleX := (screenWidth - titleBounds.Dx()) / 2
	text.Draw(screen, title, s.face, titleX, screenHeight/4, color.White)

	// Make sure we're initialized before trying to draw buttons
	if !s.initialized || len(s.aiButtonBounds) < 2 ||
		len(s.aiButtonBounds[0]) == 0 || len(s.aiButtonBounds[1]) == 0 {
		// Draw error message or just return
		text.Draw(screen, "Loading...", s.face, screenWidth/2-30, screenHeight/2, color.White)
		return
	}

	// Draw player labels
	player1Label := "Black Player (AI):"
	text.Draw(screen, player1Label, s.face, s.aiButtonBounds[0][0][0], s.aiButtonBounds[0][0][1]-20, color.White)

	player2Label := "White Player (AI):"
	text.Draw(screen, player2Label, s.face, s.aiButtonBounds[1][0][0], s.aiButtonBounds[1][0][1]-20, color.White)

	// Draw AI buttons for both players
	aiOptions := []string{"V1", "V2"}

	// Draw first player's AI buttons
	for i, optionText := range aiOptions {
		if i >= len(s.aiButtonBounds[0]) {
			continue
		}
		bounds := s.aiButtonBounds[0][i]

		var buttonColor color.RGBA
		if s.selectedAIs[0] == i {
			buttonColor = color.RGBA{0, 150, 0, 255} // Selected
		} else if s.currentHoverPlayer == 0 && s.currentHoverButton == i {
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

	// Draw second player's AI buttons
	for i, optionText := range aiOptions {
		if i >= len(s.aiButtonBounds[1]) {
			continue
		}
		bounds := s.aiButtonBounds[1][i]

		var buttonColor color.RGBA
		if s.selectedAIs[1] == i {
			buttonColor = color.RGBA{0, 150, 0, 255} // Selected
		} else if s.currentHoverPlayer == 1 && s.currentHoverButton == i {
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

	// Draw selection summary
	var selectionText string
	if s.selectedAIs[0] >= 0 && s.selectedAIs[1] >= 0 {
		aiNames := []string{"V1", "V2"}
		selectionText = fmt.Sprintf("%s vs %s",
			aiNames[s.selectedAIs[0]],
			aiNames[s.selectedAIs[1]])
	} else {
		selectionText = "Please select both AIs"
	}

	selectionBounds := text.BoundString(s.face, selectionText)
	selectionX := (screenWidth - selectionBounds.Dx()) / 2
	text.Draw(screen, selectionText, s.face, selectionX, s.aiButtonBounds[1][0][1]+80, color.White)

	// Draw play button (only if both AIs are selected)
	buttonColor := color.RGBA{100, 100, 100, 255} // Disabled
	if s.selectedAIs[0] >= 0 && s.selectedAIs[1] >= 0 {
		buttonColor = color.RGBA{0, 100, 0, 255} // Enabled
		if s.buttonHovered == 2*len(aiOptions) {
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
	if s.buttonHovered == 2*len(aiOptions)+1 {
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
