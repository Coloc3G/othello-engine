package main

import (
	"image/color"
	"log"

	"github.com/yourusername/othello/game" // Import de ton moteur Othello

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 600
	screenHeight = 600
	gridSize     = 8
	cellSize     = screenWidth / gridSize // Chaque case fait 75x75 px
)

// États du jeu
const (
	GameStatePlaying = 0 // En cours
	GameStateEnded   = 1 // Fin de partie
)

// Game représente l'état du jeu
type Game struct {
	state      int       // État du jeu
	othello    game.Game // Instance du moteur Othello
	winnerText string    // Texte du gagnant
}

// Update gère la logique du jeu
func (g *Game) Update() error {
	if g.state == GameStateEnded {
		return nil
	}

	// Gestion du clic pour jouer un coup
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row := y / cellSize
		col := x / cellSize

		// Tenter de jouer le coup
		moveSuccess := g.othello.ApplyMove(game.Position{Row: row, Col: col})

		if moveSuccess {
			// Si le coup est valide, l'IA joue après le joueur
			g.othello.ApplyMove(g.othello.GetValidMovesForCurrentPlayer()[0])
		}

		// Vérifier si la partie est finie
		if g.othello.IsGameFinishedMethod() {
			g.state = GameStateEnded
			winner := g.othello.GetWinnerMethod()
			if winner == game.Black {
				g.winnerText = "Le joueur noir gagne !"
			} else if winner == game.White {
				g.winnerText = "Le joueur blanc gagne !"
			} else {
				g.winnerText = "Match nul !"
			}
		}
	}
	return nil
}

// Draw dessine la grille et les pions
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBoard(screen)
	g.drawPieces(screen)

	// Affichage du gagnant si la partie est terminée
	if g.state == GameStateEnded {
		ebitenutil.DebugPrint(screen, g.winnerText+"\nAppuyez sur R pour recommencer")
	}
}

// drawBoard dessine la grille
func (g *Game) drawBoard(screen *ebiten.Image) {
	gridColor := color.RGBA{0, 255, 0, 255} // Vert

	for i := 0; i <= gridSize; i++ {
		ebitenutil.DrawLine(screen, 0, float64(i*cellSize), screenWidth, float64(i*cellSize), gridColor)
		ebitenutil.DrawLine(screen, float64(i*cellSize), 0, float64(i*cellSize), screenHeight, gridColor)
	}
}

// drawPieces dessine les pions Othello
func (g *Game) drawPieces(screen *ebiten.Image) {
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			switch g.othello.Board[row][col] {
			case game.Black:
				ebitenutil.DrawRect(screen, float64(col*cellSize+10), float64(row*cellSize+10), 55, 55, color.Black)
			case game.White:
				ebitenutil.DrawRect(screen, float64(col*cellSize+10), float64(row*cellSize+10), 55, 55, color.White)
			}
		}
	}
}

// Layout définit la taille de l'écran
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Othello - Ebiten")

	game := &Game{state: GameStatePlaying, othello: game.NewGame()} // Démarre une nouvelle partie
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
