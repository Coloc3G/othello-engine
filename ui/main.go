package main

import (
	"image/color"
	"log"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 600
	screenHeight = 700 // Augmenté pour le titre
	gridSize     = 8
	cellSize     = screenWidth / gridSize // Chaque case fait 75x75 px
)

// États du jeu
const (
	GameStateMenu    = 0 // Page d'accueil
	GameStatePlaying = 1 // Partie Joueur vs IA
	GameStateIAvsIA  = 2 // Mode IA contre IA
	GameStateEnded   = 3 // Fin de partie
)

// Coordonnées des boutons
var (
	buttonX      = 200
	buttonY1     = 400 // Position du bouton "Jouer contre l'IA"
	buttonY2     = 500 // Position du bouton "IA contre IA"
	buttonWidth  = 200
	buttonHeight = 50
)

// Game représente l'état du jeu
type Game struct {
	state      int                        // État du jeu
	othello    game.Game                  // Instance du moteur Othello
	winnerText string                     // Texte du gagnant
	eval       evaluation.MixedEvaluation // IA utilisant plusieurs heuristiques
	lastMove   time.Time                  // Pour gérer le délai en mode IA vs IA
}

// Layout définit la taille de l'écran
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Update gère la logique du jeu
func (g *Game) Update() error {
	switch g.state {
	case GameStateMenu:
		g.handleMenuClick()
	case GameStatePlaying:
		g.handlePlayerVsAIMode()
	case GameStateIAvsIA:
		g.handleIAvsIAMode()
	case GameStateEnded:
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.resetGame()
		}
	}
	return nil
}

// Gère le clic sur les boutons de la page d'accueil
func (g *Game) handleMenuClick() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x >= buttonX && x <= buttonX+buttonWidth {
			if y >= buttonY1 && y <= buttonY1+buttonHeight {
				g.state = GameStatePlaying
			}
			if y >= buttonY2 && y <= buttonY2+buttonHeight {
				g.state = GameStateIAvsIA
				g.lastMove = time.Now() // Initialisation du timer
			}
		}
	}
}

// Mode Joueur vs IA
func (g *Game) handlePlayerVsAIMode() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row := (y - 100) / cellSize
		col := x / cellSize
		if row >= 0 && row < gridSize && col >= 0 && col < gridSize {
			moveSuccess := g.othello.ApplyMove(game.Position{Row: row, Col: col})
			if moveSuccess {
				bestMove := evaluation.Solve(g.othello, g.othello.CurrentPlayer, 5, &g.eval)
				g.othello.ApplyMove(bestMove)
			}
		}
	}
	g.checkGameEnd()
}

// Mode IA contre IA (avec délai)
func (g *Game) handleIAvsIAMode() {
	if time.Since(g.lastMove) < time.Second {
		return
	}
	bestMove := evaluation.Solve(g.othello, g.othello.CurrentPlayer, 3, &g.eval)
	g.othello.ApplyMove(bestMove)
	g.lastMove = time.Now()
	g.checkGameEnd()
}

// Vérifie si la partie est finie
func (g *Game) checkGameEnd() {
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

// Réinitialise le jeu
func (g *Game) resetGame() {
	g.state = GameStateMenu
	g.othello = game.NewGame()
	g.winnerText = ""
}

// Draw affiche l'écran selon l'état du jeu
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 30, 255}) // Fond gris foncé

	switch g.state {
	case GameStateMenu:
		g.drawMenu(screen)
	case GameStatePlaying, GameStateIAvsIA:
		g.drawTitle(screen)
		g.drawBoard(screen)
		g.drawPieces(screen)
	case GameStateEnded:
		ebitenutil.DebugPrint(screen, g.winnerText+"\nAppuyez sur R pour recommencer")
	}
}

// Affiche la page d'accueil
func (g *Game) drawMenu(screen *ebiten.Image) {
	g.drawTitle(screen)

	// Dessiner les boutons
	g.drawButton(screen, buttonX, buttonY1, "Jouer contre l'IA")
	g.drawButton(screen, buttonX, buttonY2, "IA contre IA")
}

// Dessine le titre centré
func (g *Game) drawTitle(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "OTHELLO GAME", screenWidth/2-60, 30)
}

// Dessine un bouton stylisé
func (g *Game) drawButton(screen *ebiten.Image, x, y int, text string) {
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(buttonWidth), float64(buttonHeight), color.RGBA{0, 150, 0, 255})
	ebitenutil.DebugPrintAt(screen, text, x+50, y+20)
}

// Dessine la grille
func (g *Game) drawBoard(screen *ebiten.Image) {
	boardColor := color.RGBA{50, 205, 50, 255} // Vert plus agréable
	ebitenutil.DrawRect(screen, 0, 100, float64(screenWidth), float64(screenWidth), boardColor)

	gridColor := color.RGBA{0, 0, 0, 255}
	for i := 0; i <= gridSize; i++ {
		ebitenutil.DrawLine(screen, 0, float64(i*cellSize+100), screenWidth, float64(i*cellSize+100), gridColor)
		ebitenutil.DrawLine(screen, float64(i*cellSize), 100, float64(i*cellSize), screenWidth+100, gridColor)
	}
}

// Dessine les pions sous forme de cercles
func (g *Game) drawPieces(screen *ebiten.Image) {
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			if g.othello.Board[row][col] != game.Empty {
				c := color.Black
				if g.othello.Board[row][col] == game.White {
					c = color.White
				}
				ebitenutil.DrawCircle(screen, float64(col*cellSize+cellSize/2), float64(row*cellSize+100+cellSize/2), 30, c)
			}
		}
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Othello - Ebiten")

	game := &Game{state: GameStateMenu, othello: game.NewGame(), eval: *evaluation.NewMixedEvaluation()}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
