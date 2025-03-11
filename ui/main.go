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
	screenHeight = 600
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
	buttonY1     = 250 // Position du bouton "Jouer contre l'IA"
	buttonY2     = 350 // Position du bouton "IA contre IA"
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

		// Jouer contre l'IA
		if x >= buttonX && x <= buttonX+buttonWidth && y >= buttonY1 && y <= buttonY1+buttonHeight {
			g.state = GameStatePlaying
		}

		// IA contre IA
		if x >= buttonX && x <= buttonX+buttonWidth && y >= buttonY2 && y <= buttonY2+buttonHeight {
			g.state = GameStateIAvsIA
			g.lastMove = time.Now() // Initialisation du timer
		}
	}
}

// Mode Joueur vs IA
func (g *Game) handlePlayerVsAIMode() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		row := y / cellSize
		col := x / cellSize

		// Tenter de jouer le coup du joueur humain
		moveSuccess := g.othello.ApplyMove(game.Position{Row: row, Col: col})

		if moveSuccess {
			// IA joue le meilleur coup après le joueur
			bestMove := evaluation.Solve(g.othello, g.othello.CurrentPlayer, 3, &g.eval) // Profondeur 3
			g.othello.ApplyMove(bestMove)
		}

		// Vérifier si la partie est finie
		g.checkGameEnd()
	}
}

// Mode IA contre IA (avec délai)
func (g *Game) handleIAvsIAMode() {
	if time.Since(g.lastMove) < time.Second { // 1 seconde entre chaque coup
		return
	}

	// IA joue
	bestMove := evaluation.Solve(g.othello, g.othello.CurrentPlayer, 3, &g.eval)
	g.othello.ApplyMove(bestMove)
	g.lastMove = time.Now() // Met à jour le timer

	// Vérifier si la partie est finie
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
	screen.Fill(color.RGBA{255, 0, 0, 255}) // Fond rouge

	switch g.state {
	case GameStateMenu:
		g.drawMenu(screen)
	case GameStatePlaying, GameStateIAvsIA:
		g.drawBoard(screen)
		g.drawPieces(screen)
	case GameStateEnded:
		ebitenutil.DebugPrint(screen, g.winnerText+"\nAppuyez sur R pour recommencer")
	}
}

// Affiche la page d'accueil
func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "OTHELLO GAME\n\nCliquez sur un mode")

	// Dessiner le bouton "Jouer contre l'IA"
	ebitenutil.DrawRect(screen, float64(buttonX), float64(buttonY1), float64(buttonWidth), float64(buttonHeight), color.White)
	ebitenutil.DebugPrintAt(screen, "Jouer contre l'IA", buttonX+40, buttonY1+20)

	// Dessiner le bouton "IA contre IA"
	ebitenutil.DrawRect(screen, float64(buttonX), float64(buttonY2), float64(buttonWidth), float64(buttonHeight), color.White)
	ebitenutil.DebugPrintAt(screen, "IA contre IA", buttonX+60, buttonY2+20)
}

// Dessine la grille
func (g *Game) drawBoard(screen *ebiten.Image) {
	gridColor := color.RGBA{255, 255, 255, 255}

	for i := 0; i <= gridSize; i++ {
		ebitenutil.DrawLine(screen, 0, float64(i*cellSize), screenWidth, float64(i*cellSize), gridColor)
		ebitenutil.DrawLine(screen, float64(i*cellSize), 0, float64(i*cellSize), screenHeight, gridColor)
	}
}

// Dessine les pions Othello
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

	game := &Game{state: GameStateMenu, othello: game.NewGame(), eval: *evaluation.NewMixedEvaluation()}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
