package ui

import (
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/hajimehoshi/ebiten/v2"
)

// AppState represents the current state/screen of the application
type AppState int

const (
	StateStart AppState = iota
	StateGame
	StateEnd
)

// UI manages the game UI
type UI struct {
	game          *game.Game
	gameScreen    *GameScreen
	resultScreen  *ResultScreen
	currentScreen Screen
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
		game: g,
	}
	ui.gameScreen = NewGameScreen(ui)
	ui.resultScreen = NewResultScreen(ui)
	ui.currentScreen = ui.gameScreen

	return ui
}

// EndGame switches to the result screen
func (ui *UI) EndGame() {
	ui.currentScreen = ui.resultScreen
}

// NewGame starts a new game
func (ui *UI) NewGame() {
	ui.game = game.NewGame(ui.game.Players[0].Name, ui.game.Players[1].Name)
	ui.currentScreen = ui.gameScreen
}

// RunUI runs the UI
func RunUI() {
	// Create initial game
	g := game.NewGame("Human", "AI")

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
