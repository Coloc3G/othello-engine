package ui

import (
	"time"

	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/hajimehoshi/ebiten/v2"
)

// AppState represents the current state/screen of the application
type AppState int

const (
	StateHome AppState = iota
	StateAISelection
	StateDualAISelection
	StateGame
	StateEnd
)

// UI manages the game UI
type UI struct {
	game                  *game.Game
	homeScreen            *HomeScreen
	aiSelectionScreen     *AISelectionScreen
	dualAISelectionScreen *DualAISelectionScreen
	gameScreen            *GameScreen
	resultScreen          *ResultScreen
	endScreen             *EndScreen
	currentScreen         Screen
	aivsAiMode            bool
	aivsAiTimer           time.Time
	aivsAiMoveDelay       time.Duration
}

// Screen interface for different game screens
type Screen interface {
	Update() error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (int, int)
}

// Game implements ebiten.Game interface
type Game struct {
	ui *UI
}

func (g *Game) Update() error {
	return g.ui.currentScreen.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ui.currentScreen.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ui.currentScreen.Layout(outsideWidth, outsideHeight)
}

// NewUI creates a new UI
func NewUI(g *game.Game) *UI {
	ui := &UI{
		game:            g,
		aivsAiMoveDelay: time.Second, // 1 second delay between AI moves
		aivsAiMode:      false,
	}

	// Create all screens
	ui.homeScreen = NewHomeScreen(ui)
	ui.aiSelectionScreen = NewAISelectionScreen(ui)
	ui.dualAISelectionScreen = NewDualAISelectionScreen(ui)
	ui.gameScreen = NewGameScreen(ui)
	ui.resultScreen = NewResultScreen(ui)
	ui.endScreen = NewEndScreen(ui)

	// Set initial screen to home screen
	ui.currentScreen = ui.homeScreen

	return ui
}

// SwitchToHomeScreen switches to the home screen
func (s *UI) SwitchToHomeScreen() {
	s.currentScreen = s.homeScreen
}

// SwitchToAISelectionScreen switches to the AI selection screen
func (s *UI) SwitchToAISelectionScreen() {
	s.currentScreen = s.aiSelectionScreen
}

// SwitchToDualAISelectionScreen switches to the dual AI selection screen
func (s *UI) SwitchToDualAISelectionScreen() {
	s.currentScreen = s.dualAISelectionScreen
}

// StartPlayerVsAIGame starts a game with a human player against the selected AI
func (s *UI) StartPlayerVsAIGame(aiVersion int) {
	// Create game with human player vs AI
	s.game = game.NewGame("Human", getAIVersionName(aiVersion))
	s.aivsAiMode = false

	// Reset the game screen
	if s.gameScreen != nil {
		s.gameScreen.lastMovePos = game.Position{Row: -1, Col: -1}
		s.gameScreen.moveHistory = make([][2]MoveRecord, 0)
		s.gameScreen.scrollOffset = 0
	}

	s.currentScreen = s.gameScreen
}

// StartAIVsAIGame starts a game with two AI players
func (s *UI) StartAIVsAIGame(ai1Version, ai2Version int) {
	// Create game with AI vs AI
	s.game = game.NewGame(
		getAIVersionName(ai1Version),
		getAIVersionName(ai2Version),
	)
	s.aivsAiMode = true
	s.aivsAiTimer = time.Now()

	// Reset the game screen
	if s.gameScreen != nil {
		s.gameScreen.lastMovePos = game.Position{Row: -1, Col: -1}
		s.gameScreen.moveHistory = make([][2]MoveRecord, 0)
		s.gameScreen.scrollOffset = 0
	}

	s.currentScreen = s.gameScreen
}

// EndGame switches to the result screen
func (ui *UI) EndGame() {
	ui.currentScreen = ui.endScreen
}

// NewGame starts a new game
func (ui *UI) NewGame() {
	ui.SwitchToHomeScreen()
}

// getAIVersionName returns the name for an AI version
func getAIVersionName(version int) string {
	switch version {
	case 0:
		return "AI (V1)"
	case 1:
		return "AI (V2)"
	default:
		return "AI"
	}
}

// RunUI runs the UI
func RunUI() {
	// Create initial game (won't be used until player makes a selection)
	g := game.NewGame("Player", "AI")

	// Create UI
	ui := NewUI(g)

	// Initialize window
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Othello")

	// Run game
	game := &Game{ui: ui}
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
