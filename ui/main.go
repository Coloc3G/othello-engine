package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 600
	screenHeight = 600
	gridSize     = 8
	cellSize     = screenWidth / gridSize // 600 / 8 = 75 px
)

// Game représente l'état du jeu
type Game struct{}

// Update est appelé avant chaque frame (pas utilisé ici)
func (g *Game) Update() error {
	return nil
}

// Draw dessine la grille 8x8
func (g *Game) Draw(screen *ebiten.Image) {
	// Couleur de la grille
	gridColor := color.RGBA{0, 255, 0, 255} // Vert

	// Dessine la grille
	for i := 0; i <= gridSize; i++ {
		// Lignes horizontales
		ebitenutil.DrawLine(screen, 0, float64(i*cellSize), screenWidth, float64(i*cellSize), gridColor)
		// Lignes verticales
		ebitenutil.DrawLine(screen, float64(i*cellSize), 0, float64(i*cellSize), screenHeight, gridColor)
	}
}

// Layout définit la taille de l'écran
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Othello - Ebiten Grid")

	game := &Game{}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
