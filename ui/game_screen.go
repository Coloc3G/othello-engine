package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

// GameScreen manages the main game UI
type GameScreen struct {
	ui              *UI
	lastMove        time.Time
	boardSize       int
	cellSize        int
	boardOffsetX    int
	boardOffsetY    int
	face            font.Face
	evaluationValue int                         // Current evaluation value
	evalHistory     []int                       // History of evaluations for visualization
	evaluator       *evaluation.MixedEvaluation // Evaluation function
	evalChan        chan int                    // Channel for receiving evaluation results
	evaluating      bool                        // Flag to track if evaluation is in progress
	currentDepth    int                         // Current evaluation depth
	resultDepth     int                         // Depth of the current evaluation result
	maxDepth        int                         // Maximum evaluation depth
	depthUpdateChan chan int                    // Channel for receiving depth updates
	evalCancelChan  chan struct{}               // Channel for cancelling ongoing evaluations
}

// NewGameScreen creates a new game screen
func NewGameScreen(ui *UI) *GameScreen {
	return &GameScreen{
		ui:              ui,
		lastMove:        time.Now(),
		face:            basicfont.Face7x13,
		evalHistory:     make([]int, 0),
		evaluator:       evaluation.NewMixedEvaluationWithCoefficients(evaluation.V2Coeff),
		evalChan:        make(chan int, 1),      // Buffered channel for evaluation results
		depthUpdateChan: make(chan int, 1),      // Buffered channel for depth updates
		evalCancelChan:  make(chan struct{}, 1), // Buffered channel for cancellation signal
		maxDepth:        5,                      // Maximum evaluation depth
	}
}

// Layout implements ebiten.Game
func (s *GameScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// Update updates the game state
func (s *GameScreen) Update() error {
	// Calculate board dimensions based on screen size
	screenWidth, screenHeight := ebiten.WindowSize()
	s.boardSize = min(screenWidth-200, screenHeight-100) // Leave space for evaluation bar
	s.cellSize = s.boardSize / 8
	s.boardOffsetX = (screenWidth - s.boardSize - 150) / 2 // Shift board left to make room for eval bar
	s.boardOffsetY = 80                                    // Leave space for header

	// Check if game is over
	if game.IsGameFinished(s.ui.game.Board) {
		s.ui.EndGame()
		return nil
	}

	// Check if current player has any valid moves
	if !s.ui.game.HasAnyMovesInGame() {
		// No valid moves, switch to the other player
		s.ui.game.CurrentPlayer = s.ui.game.GetOtherPlayerMethod()
		return nil
	}

	// Check for depth updates
	select {
	case newDepth := <-s.depthUpdateChan:
		s.currentDepth = newDepth
	default:
		// No depth update
	}

	// Check for finished evaluations
	select {
	case evalResult := <-s.evalChan:
		s.evaluationValue = evalResult
		s.resultDepth = s.currentDepth // Store the depth of this evaluation result
		s.evalHistory = append(s.evalHistory, evalResult)

		// Cap history size to prevent memory issues
		if len(s.evalHistory) > 100 {
			s.evalHistory = s.evalHistory[len(s.evalHistory)-100:]
		}
	default:
		// No evaluation result ready yet
	}

	if s.ui.game.CurrentPlayer.Name == "AI" {
		eval := evaluation.NewMixedEvaluationWithCoefficients(evaluation.V2Coeff)
		pos := evaluation.Solve(*s.ui.game, s.ui.game.CurrentPlayer, 5, eval)
		// Apply move and update evaluation
		if s.ui.game.ApplyMove(pos) {
			fmt.Println(utils.PositionToAlgebraic(pos))
			s.updateEvaluation()
			s.lastMove = time.Now()
		}
	} else {
		// Handle mouse input
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()

			// Determine if click was within board bounds
			if x >= s.boardOffsetX && x < s.boardOffsetX+s.boardSize &&
				y >= s.boardOffsetY && y < s.boardOffsetY+s.boardSize {

				// Calculate board position
				boardX := (x - s.boardOffsetX) / s.cellSize
				boardY := (y - s.boardOffsetY) / s.cellSize

				pos := game.Position{Row: boardY, Col: boardX}

				// Try to make the move
				if s.ui.game.ApplyMove(pos) {
					s.updateEvaluation()
					s.lastMove = time.Now()
				}
			}
		}
	}

	return nil
}

// Draw renders the game screen
func (s *GameScreen) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(ColorBackground)

	// Draw header info
	s.drawHeader(screen)

	// Draw the game board
	s.drawGameBoard(screen)

	// Draw evaluation bar
	s.drawEvaluationBar(screen)
}

// drawHeader renders the game status information
func (s *GameScreen) drawHeader(screen *ebiten.Image) {
	currentPlayer := s.ui.game.CurrentPlayer
	blackCount, whiteCount := game.CountPieces(s.ui.game.Board)

	// Draw title
	title := "Othello"
	titleBounds := text.BoundString(s.face, title)
	titleX := (screen.Bounds().Dx() - titleBounds.Dx()) / 2
	text.Draw(screen, title, s.face, titleX, 20, color.White)

	// Draw player info
	playerColorTxt := "Black"
	if currentPlayer.Color == game.White {
		playerColorTxt = "White"
	}
	playerInfo := fmt.Sprintf("Current Player: %s (%s)", currentPlayer.Name, playerColorTxt)
	playerBounds := text.BoundString(s.face, playerInfo)
	playerX := (screen.Bounds().Dx() - playerBounds.Dx()) / 2
	text.Draw(screen, playerInfo, s.face, playerX, 40, color.White)

	// Draw score
	scoreInfo := fmt.Sprintf("Black: %d | White: %d", blackCount, whiteCount)
	scoreBounds := text.BoundString(s.face, scoreInfo)
	scoreX := (screen.Bounds().Dx() - scoreBounds.Dx()) / 2
	text.Draw(screen, scoreInfo, s.face, scoreX, 60, color.White)
}

// drawGameBoard renders the game board
func (s *GameScreen) drawGameBoard(screen *ebiten.Image) {
	// Draw board background
	ebitenutil.DrawRect(screen, float64(s.boardOffsetX), float64(s.boardOffsetY),
		float64(s.boardSize), float64(s.boardSize),
		color.RGBA{34, 100, 34, 255})

	// Get valid moves for current player
	validMoves := s.ui.game.GetValidMovesForCurrentPlayer()

	// Draw grid and pieces
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			x := s.boardOffsetX + col*s.cellSize
			y := s.boardOffsetY + row*s.cellSize

			// Draw cell border
			ebitenutil.DrawRect(screen, float64(x), float64(y),
				float64(s.cellSize), float64(s.cellSize),
				ColorGrid)

			// Draw cell interior
			ebitenutil.DrawRect(screen, float64(x+1), float64(y+1),
				float64(s.cellSize-2), float64(s.cellSize-2),
				color.RGBA{50, 150, 50, 255})

			// Check if this is a valid move
			isValidMove := false
			for _, pos := range validMoves {
				if pos.Row == row && pos.Col == col {
					isValidMove = true
					break
				}
			}

			// Draw valid move indicator
			if isValidMove {
				ebitenutil.DrawRect(screen, float64(x+3), float64(y+3),
					float64(s.cellSize-6), float64(s.cellSize-6),
					ColorValid)
			}

			// Draw piece if present
			piece := s.ui.game.Board[row][col]
			if piece != game.Empty {
				pieceColor := ColorWhite
				if piece == game.Black {
					pieceColor = ColorBlack
				}

				// Draw circle for piece
				centerX := float64(x + s.cellSize/2)
				centerY := float64(y + s.cellSize/2)
				radius := float64(s.cellSize/2 - 4)
				s.drawCircle(screen, centerX, centerY, radius, pieceColor)
			}
		}
	}
}

// drawCircle draws a filled circle
func (s *GameScreen) drawCircle(screen *ebiten.Image, x, y, radius float64, col color.Color) {
	// Draw a circle using the midpoint circle algorithm
	for yOff := -radius; yOff <= radius; yOff++ {
		for xOff := -radius; xOff <= radius; xOff++ {
			if xOff*xOff+yOff*yOff <= radius*radius {
				screen.Set(int(x+xOff), int(y+yOff), col)
			}
		}
	}
}

// min returns the minimum of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updateEvaluation starts an asynchronous progressive depth evaluation
func (s *GameScreen) updateEvaluation() {
	// Cancel any ongoing evaluation
	if s.evaluating {
		select {
		case s.evalCancelChan <- struct{}{}:
			// Signal sent successfully
		default:
			// Channel already has a signal, no need to send another one
		}
	}

	// Start the progressive evaluation process
	s.evaluating = true
	s.currentDepth = 1 // Reset depth counter

	// Create a copy of the game for evaluation
	gameCopy := *s.ui.game

	// Always evaluate from black's perspective for consistency
	player := s.ui.game.Players[0]

	// Create a new evaluation cycle
	go func() {
		defer func() { s.evaluating = false }()

		// Start with shallow depth and progressively increase
		for depth := 2; depth <= s.maxDepth; depth += 1 {
			// Check if we should cancel this evaluation
			select {
			case <-s.evalCancelChan:
				return // Stop evaluating
			default:
				// Continue evaluation
			}

			// Update the current depth
			select {
			case s.depthUpdateChan <- depth:
			default:
				// Channel full, continue anyway
			}

			// Perform evaluation at current depth
			evalScore := evaluation.MMAB(
				gameCopy,
				gameCopy.Board,
				player,
				depth,   // current progressive depth
				true,    // maximizing
				-1<<31,  // alpha
				1<<31-1, // beta
				s.evaluator)

			// Check again if we should cancel before sending result
			select {
			case <-s.evalCancelChan:
				return // Stop evaluating and don't send result
			default:
				// Send the result
				select {
				case s.evalChan <- evalScore:
					// Successfully sent
				default:
					// Channel full, clear it and send new value
					select {
					case <-s.evalChan: // Discard old value
					default:
						// Channel was already empty
					}
					s.evalChan <- evalScore
				}
			}

			// Small sleep to prevent CPU hogging and allow UI updates
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

// drawEvaluationBar draws the evaluation bar on the right side of the board
func (s *GameScreen) drawEvaluationBar(screen *ebiten.Image) {
	// Bar position and dimensions
	barX := s.boardOffsetX + s.boardSize + 20
	barY := s.boardOffsetY
	barWidth := 30
	barHeight := s.boardSize

	// Draw bar background
	ebitenutil.DrawRect(screen, float64(barX), float64(barY),
		float64(barWidth), float64(barHeight), color.RGBA{40, 40, 40, 255})

	// Calculate bar fill based on evaluation
	// Normalize evaluation value to a percentage (-2000 to +2000 range)
	evalRange := 2000.0
	normalizedEval := float64(s.evaluationValue) / evalRange
	if normalizedEval > 1.0 {
		normalizedEval = 1.0
	}
	if normalizedEval < -1.0 {
		normalizedEval = -1.0
	}

	// Center represents neutral evaluation (0)
	centerY := barY + barHeight/2

	// Draw the neutral line
	ebitenutil.DrawLine(screen,
		float64(barX), float64(centerY),
		float64(barX+barWidth), float64(centerY),
		color.RGBA{100, 100, 100, 255})

	// Draw the evaluation fill
	fillHeight := int(float64(barHeight/2) * normalizedEval)

	var fillColor color.RGBA

	if normalizedEval > 0 {
		// Positive evaluation (good for black) - green bar going up from center
		fillColor = color.RGBA{0, 200, 0, 255}
		ebitenutil.DrawRect(screen,
			float64(barX), float64(centerY-fillHeight),
			float64(barWidth), float64(fillHeight),
			fillColor)
	} else {
		// Negative evaluation (good for white) - red bar going down from center
		fillColor = color.RGBA{200, 0, 0, 255}
		ebitenutil.DrawRect(screen,
			float64(barX), float64(centerY),
			float64(barWidth), float64(-fillHeight),
			fillColor)
	}

	// Draw evaluation text with depth information
	var evalText string
	if s.evaluating {
		evalText = fmt.Sprintf("%+d d:%d/%d", s.evaluationValue, s.resultDepth, s.currentDepth)
	} else {
		evalText = fmt.Sprintf("%+d d:%d", s.evaluationValue, s.resultDepth)
	}

	textBounds := text.BoundString(s.face, evalText)
	textX := barX + (barWidth-textBounds.Dx())/2
	textY := barY + barHeight + 20
	text.Draw(screen, evalText, s.face, textX, textY, color.White)

	// Add a "thinking" indicator if evaluation is in progress
	if s.evaluating {
		thinkingText := "thinking..."
		thinkX := barX - 10
		thinkY := barY - 20
		text.Draw(screen, thinkingText, s.face, thinkX, thinkY, color.RGBA{200, 200, 0, 255})
	}

	// Label for black (top)
	text.Draw(screen, "Black", s.face, barX, barY-5, color.White)

	// Label for white (bottom)
	whiteLabelY := barY + barHeight + 35
	text.Draw(screen, "White", s.face, barX, whiteLabelY, color.White)
}
