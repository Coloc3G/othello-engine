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
)

// MoveRecord represents a single move made by a player
type MoveRecord struct {
	Position game.Position
	Pass     bool
}

// GameScreen manages the main game UI
type GameScreen struct {
	ui              *UI
	lastMove        time.Time
	lastMovePos     game.Position   // Track the last move position
	moveHistory     [][2]MoveRecord // Store move history as pairs [black, white]
	scrollOffset    int             // For scrolling through move history
	maxVisibleMoves int             // Maximum number of visible moves in the history panel
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
		lastMovePos:     game.Position{Row: -1, Col: -1}, // Initialize with invalid position
		moveHistory:     make([][2]MoveRecord, 0),
		scrollOffset:    0,
		maxVisibleMoves: 10, // Number of moves visible in the history panel
		face:            basicfont.Face7x13,
		evalHistory:     make([]int, 0),
		evaluator:       evaluation.NewMixedEvaluationWithCoefficients(evaluation.V2Coeff),
		evalChan:        make(chan int, 1),      // Buffered channel for evaluation results
		depthUpdateChan: make(chan int, 1),      // Buffered channel for depth updates
		evalCancelChan:  make(chan struct{}, 1), // Buffered channel for cancellation signal
		maxDepth:        5,                      // Maximum evaluation depth
	}
}

// Layout implements the Screen interface
func (s *GameScreen) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// AddMoveToHistory adds a move to the history table
func (s *GameScreen) AddMoveToHistory(pos game.Position, playerColor game.Piece, pass bool) {
	moveRecord := MoveRecord{
		Position: pos,
		Pass:     pass,
	}

	// If it's a black move, create a new turn
	if playerColor == game.Black {
		s.moveHistory = append(s.moveHistory, [2]MoveRecord{
			moveRecord, // Black move
			{Position: game.Position{Row: -1, Col: -1}}, // Placeholder for white move
		})
	} else {
		// If it's a white move, update the last turn's white move
		if len(s.moveHistory) > 0 {
			lastIdx := len(s.moveHistory) - 1
			s.moveHistory[lastIdx][1] = moveRecord
		} else {
			// This is an unusual case where white moves first
			// Create a new entry with placeholder for black
			s.moveHistory = append(s.moveHistory, [2]MoveRecord{
				{Position: game.Position{Row: -1, Col: -1}}, // Placeholder for black
				moveRecord, // White move
			})
		}
	}

	// Adjust scroll offset to show the latest moves only if we need to scroll
	if len(s.moveHistory) > s.maxVisibleMoves {
		s.scrollOffset = len(s.moveHistory) - s.maxVisibleMoves
	} else {
		s.scrollOffset = 0 // No need to scroll when fewer moves than visible area
	}
}

// Update updates the game state
func (s *GameScreen) Update() error {
	// Calculate board dimensions based on screen size
	screenWidth, screenHeight := ebiten.WindowSize()
	s.boardSize = min(screenWidth-300, screenHeight-100) // Reduce board size to make room for history
	s.cellSize = s.boardSize / 8
	s.boardOffsetX = (screenWidth - s.boardSize - 250) / 2 // Shift board left to make room for eval bar and history
	s.boardOffsetY = 80                                    // Leave space for header

	// Handle mouse wheel for scrolling move history
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		// Check if mouse is over history panel
		mouseX, _ := ebiten.CursorPosition()
		historyPanelX := s.boardOffsetX + s.boardSize + 80
		if mouseX >= historyPanelX {
			// Scroll history
			s.scrollOffset -= int(scrollY * 3) // Adjust scroll speed

			// Clamp scrolling
			if s.scrollOffset < 0 {
				s.scrollOffset = 0
			}
			maxScroll := max(0, len(s.moveHistory)-s.maxVisibleMoves)
			if s.scrollOffset > maxScroll {
				s.scrollOffset = maxScroll
			}
		}
	}

	// Check if game is over
	if game.IsGameFinished(s.ui.game.Board) {
		s.ui.EndGame()
		return nil
	}

	// Check if current player has any valid moves
	if !s.ui.game.HasAnyMovesInGame() {
		// No valid moves, add a "Pass" record to history
		s.AddMoveToHistory(game.Position{Row: -1, Col: -1}, s.ui.game.CurrentPlayer.Color, true)

		// Switch to the other player
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

	// Handle AI vs AI mode
	if s.ui.aivsAiMode {
		currentTime := time.Now()
		if currentTime.Sub(s.ui.aivsAiTimer) >= s.ui.aivsAiMoveDelay {
			// Time to make another AI move
			eval := s.evaluator
			pos := evaluation.Solve(*s.ui.game, s.ui.game.CurrentPlayer, 5, eval)

			// Apply move and update evaluation
			if s.ui.game.ApplyMove(pos) {
				s.lastMovePos = pos                                           // Update last move position
				s.AddMoveToHistory(pos, s.ui.game.CurrentPlayer.Color, false) // Add to history
				s.updateProgressiveEvaluation()                               // Update evaluation
				s.ui.aivsAiTimer = currentTime                                // Reset timer for next move
			}
		}
		return nil
	}

	// Handle human vs AI mode
	if s.ui.game.CurrentPlayer.Name == "Human" {
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
					s.lastMovePos = pos                                           // Update last move position
					s.AddMoveToHistory(pos, s.ui.game.CurrentPlayer.Color, false) // Add to history
					s.updateProgressiveEvaluation()                               // Update evaluation
					s.lastMove = time.Now()
				}
			}
		}
	} else if s.ui.game.CurrentPlayer.Name != "Human" {
		// Handle AI move
		eval := s.evaluator
		pos := evaluation.Solve(*s.ui.game, s.ui.game.CurrentPlayer, 5, eval)

		// Apply move and update evaluation
		if s.ui.game.ApplyMove(pos) {
			s.lastMovePos = pos                                           // Update last move position
			s.AddMoveToHistory(pos, s.ui.game.CurrentPlayer.Color, false) // Add to history
			s.updateProgressiveEvaluation()                               // Update evaluation
			s.lastMove = time.Now()
		}
	}

	return nil
}

// Draw renders the game screen
func (s *GameScreen) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(ColorBackground)

	// Draw header info
	s.drawHeaderInfo(screen)

	// Draw the game board
	s.drawGameBoard(screen)

	// Draw move history table
	s.drawMoveHistory(screen)

	// Draw evaluation bar
	s.drawEvaluationBar(screen)

	// Draw AI vs AI indicator if in that mode
	if s.ui.aivsAiMode {
		screenWidth, _ := screen.Bounds().Dx(), screen.Bounds().Dy()
		aivsaiText := "AI vs AI Mode"
		text.Draw(screen, aivsaiText, s.face, screenWidth-120, 20, color.RGBA{255, 215, 0, 255})
	}
}

// drawHeaderInfo renders the game status information
func (s *GameScreen) drawHeaderInfo(screen *ebiten.Image) {
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

// drawMoveHistory draws the move history table
func (s *GameScreen) drawMoveHistory(screen *ebiten.Image) {
	// Calculate position for history panel
	historyX := s.boardOffsetX + s.boardSize + 80
	historyY := s.boardOffsetY
	historyWidth := 170
	historyHeight := s.boardSize
	cellHeight := 25

	// Dynamically calculate the maximum visible moves based on available height
	s.maxVisibleMoves = (historyHeight - 24) / cellHeight // Subtract header height (24px)

	// Draw history panel background
	ebitenutil.DrawRect(screen, float64(historyX), float64(historyY),
		float64(historyWidth), float64(historyHeight),
		color.RGBA{40, 40, 40, 255})

	// Draw history panel title
	titleText := "Move History"
	titleBounds := text.BoundString(s.face, titleText)
	titleX := historyX + (historyWidth-titleBounds.Dx())/2
	text.Draw(screen, titleText, s.face, titleX, historyY-10, color.White)

	// Draw column headers
	blackCol := "Black"
	whiteCol := "White"
	turnCol := "Turn"

	colWidth := historyWidth / 3

	// Draw header background
	ebitenutil.DrawRect(screen, float64(historyX), float64(historyY),
		float64(historyWidth), float64(24),
		color.RGBA{60, 60, 60, 255})

	// Draw table header
	text.Draw(screen, turnCol, s.face, historyX+10, historyY+16, color.White)
	text.Draw(screen, blackCol, s.face, historyX+colWidth+10, historyY+16, color.White)
	text.Draw(screen, whiteCol, s.face, historyX+2*colWidth+10, historyY+16, color.White)

	// Draw horizontal line under header
	ebitenutil.DrawLine(screen,
		float64(historyX), float64(historyY+24),
		float64(historyX+historyWidth), float64(historyY+24),
		color.RGBA{100, 100, 100, 255})

	// Draw vertical lines between columns
	ebitenutil.DrawLine(screen,
		float64(historyX+colWidth), float64(historyY),
		float64(historyX+colWidth), float64(historyY+historyHeight),
		color.RGBA{100, 100, 100, 255})

	ebitenutil.DrawLine(screen,
		float64(historyX+2*colWidth), float64(historyY),
		float64(historyX+2*colWidth), float64(historyY+historyHeight),
		color.RGBA{100, 100, 100, 255})

	// Determine visible range of moves
	startIdx := 0
	if len(s.moveHistory) > s.maxVisibleMoves {
		startIdx = s.scrollOffset
	}
	endIdx := min(len(s.moveHistory), startIdx+s.maxVisibleMoves)

	// Draw visible moves
	for i := startIdx; i < endIdx; i++ {
		rowY := historyY + 24 + (i-startIdx)*cellHeight

		// Draw row background (alternating colors)
		rowColor := color.RGBA{50, 50, 50, 255}
		if i%2 == 1 {
			rowColor = color.RGBA{45, 45, 45, 255}
		}

		ebitenutil.DrawRect(screen, float64(historyX), float64(rowY),
			float64(historyWidth), float64(cellHeight),
			rowColor)

		// Draw turn number
		turnText := fmt.Sprintf("%d", i+1)
		text.Draw(screen, turnText, s.face, historyX+10, rowY+16, color.White)

		// Draw black move
		blackMove := s.moveHistory[i][0]
		blackText := "Pass"
		if !blackMove.Pass && blackMove.Position.Row >= 0 {
			colLetter := string('A' + blackMove.Position.Col)
			rowNumber := blackMove.Position.Row + 1
			blackText = fmt.Sprintf("%s%d", colLetter, rowNumber)
		}
		text.Draw(screen, blackText, s.face, historyX+colWidth+10, rowY+16, color.White)

		// Draw white move
		whiteMove := s.moveHistory[i][1]
		whiteText := "Pass"
		if !whiteMove.Pass && whiteMove.Position.Row >= 0 {
			colLetter := string('A' + whiteMove.Position.Col)
			rowNumber := whiteMove.Position.Row + 1
			whiteText = fmt.Sprintf("%s%d", colLetter, rowNumber)
		}
		text.Draw(screen, whiteText, s.face, historyX+2*colWidth+10, rowY+16, color.White)

		// Draw horizontal line under each row
		ebitenutil.DrawLine(screen,
			float64(historyX), float64(rowY+cellHeight),
			float64(historyX+historyWidth), float64(rowY+cellHeight),
			color.RGBA{70, 70, 70, 255})
	}

	// Only show scroll indicators and instructions if there are more moves than can be displayed
	if len(s.moveHistory) > s.maxVisibleMoves {
		// Draw scroll indicators if needed
		if s.scrollOffset > 0 {
			// More moves above
			upArrow := "▲"
			arrowBounds := text.BoundString(s.face, upArrow)
			arrowX := historyX + (historyWidth-arrowBounds.Dx())/2
			text.Draw(screen, upArrow, s.face, arrowX, historyY+40, color.RGBA{200, 200, 200, 255})
		}

		if s.scrollOffset+s.maxVisibleMoves < len(s.moveHistory) {
			// More moves below
			downArrow := "▼"
			arrowBounds := text.BoundString(s.face, downArrow)
			arrowX := historyX + (historyWidth-arrowBounds.Dx())/2
			text.Draw(screen, downArrow, s.face, arrowX, historyY+historyHeight-10, color.RGBA{200, 200, 200, 255})
		}

		// Draw scroll instructions
		scrollText := "Mouse wheel to scroll"
		textBounds := text.BoundString(s.face, scrollText)
		textX := historyX + (historyWidth-textBounds.Dx())/2
		text.Draw(screen, scrollText, s.face, textX, historyY+historyHeight+15, color.RGBA{180, 180, 180, 255})
	}
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

			// Determine cell color - check if this is the last move position
			cellColor := color.RGBA{50, 150, 50, 255} // Default cell color

			if s.lastMovePos.Row == row && s.lastMovePos.Col == col {
				// Highlight the last move with a different color
				cellColor = ColorLastMove
			}

			// Draw cell interior
			ebitenutil.DrawRect(screen, float64(x+1), float64(y+1),
				float64(s.cellSize-2), float64(s.cellSize-2),
				cellColor)

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

	// Draw coordinate labels around the board
	s.drawBoardCoordinates(screen)

	// Draw last move indicator text
	if s.lastMovePos.Row >= 0 && s.lastMovePos.Row < 8 &&
		s.lastMovePos.Col >= 0 && s.lastMovePos.Col < 8 {
		colLetter := string('A' + s.lastMovePos.Col)
		rowNumber := s.lastMovePos.Row + 1
		lastMoveText := fmt.Sprintf("Last move: %s%d", colLetter, rowNumber)

		textX := s.boardOffsetX + s.boardSize + 80
		textY := s.boardOffsetY + s.boardSize - 20

		// Draw with a more visible color
		text.Draw(screen, lastMoveText, s.face, textX, textY, ColorLastMove)
	}
}

// drawBoardCoordinates draws the row and column coordinate labels
func (s *GameScreen) drawBoardCoordinates(screen *ebiten.Image) {
	// Column labels (A-H)
	for col := 0; col < 8; col++ {
		colLabel := string('A' + col)
		labelBounds := text.BoundString(s.face, colLabel)
		labelX := s.boardOffsetX + col*s.cellSize + (s.cellSize-labelBounds.Dx())/2
		labelY := s.boardOffsetY - 5 // Above the board
		text.Draw(screen, colLabel, s.face, labelX, labelY, ColorLabelText)
	}

	// Row labels (1-8) - only on the left
	for row := 0; row < 8; row++ {
		rowLabel := fmt.Sprintf("%d", row+1)
		labelBounds := text.BoundString(s.face, rowLabel)
		labelX := s.boardOffsetX - labelBounds.Dx() - 5 // Left of the board
		labelY := s.boardOffsetY + row*s.cellSize + (s.cellSize+labelBounds.Dy())/2
		text.Draw(screen, rowLabel, s.face, labelX, labelY, ColorLabelText)
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

// updateProgressiveEvaluation starts an asynchronous progressive depth evaluation
func (s *GameScreen) updateProgressiveEvaluation() {
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
				s.evaluator,
				nil) // Pass nil for performance stats since we don't track them in the UI

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

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
