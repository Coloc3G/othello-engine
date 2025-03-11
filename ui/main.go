package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 600
	screenHeight = 600
	gridSize     = 8
	cellSize     = screenWidth / gridSize // 600 / 8 = 75 px
)

// États du jeu
const (
	GameStateMenu    = 0 // Page d'accueil
	GameStatePlaying = 1 // Jeu en cours
)

// Bouton "Jouer"
var buttonX, buttonY = 200, 350         // Position du bouton
var buttonWidth, buttonHeight = 200, 50 // Taille du bouton

// Game représente l'état du jeu
type Game struct {
	state int // État actuel du jeu (Menu ou Jeu)
}

// Update gère la logique du jeu
func (g *Game) Update() error {
	// Vérifier si on clique sur le bouton
	if g.state == GameStateMenu && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x >= buttonX && x <= buttonX+buttonWidth && y >= buttonY && y <= buttonY+buttonHeight {
			g.state = GameStatePlaying
		}
	}
	return nil
}

// Draw dessine l'écran selon l'état du jeu
func (g *Game) Draw(screen *ebiten.Image) {
	if g.state == GameStateMenu {
		g.drawMenu(screen) // Afficher le menu
	} else {
		g.drawBoard(screen) // Afficher la grille
	}
}

// drawMenu dessine la page d'accueil avec le bouton "Jouer"
func (g *Game) drawMenu(screen *ebiten.Image) {
	// Titre du jeu en grand
	titleText := "OTHELLO GAME"
	titleFont := basicfont.Face7x13 // Simple police bitmap intégrée
	text.Draw(screen, titleText, titleFont, 250, 250, color.White)

	// Dessiner le bouton
	ebitenutil.DrawRect(screen, float64(buttonX), float64(buttonY), float64(buttonWidth), float64(buttonHeight), color.RGBA{0, 255, 0, 255})

	// Texte du bouton "Jouer"
	text.Draw(screen, "JOUER", titleFont, buttonX+60, buttonY+30, color.Black)
}

// drawBoard dessine la grille du jeu
func (g *Game) drawBoard(screen *ebiten.Image) {
	gridColor := color.RGBA{0, 255, 0, 255} // Vert

	// Dessine la grille
	for i := 0; i <= gridSize; i++ {
		ebitenutil.DrawLine(screen, 0, float64(i*cellSize), screenWidth, float64(i*cellSize), gridColor)  // Lignes horizontales
		ebitenutil.DrawLine(screen, float64(i*cellSize), 0, float64(i*cellSize), screenHeight, gridColor) // Lignes verticales
	}
}

// Layout définit la taille de l'écran
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Othello - Ebiten")

	game := &Game{state: GameStateMenu} // Démarre sur le menu
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
