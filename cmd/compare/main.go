package main

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/Coloc3G/othello-engine/models/utils"
)

type Model struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (m *Model) recvUntil(delim []byte) ([]byte, error) {
	var buffer []byte
	buf := make([]byte, 1)

	for {
		n, err := m.stdout.Read(buf)
		if err != nil {
			return buffer, err
		}

		if n > 0 {
			buffer = append(buffer, buf[0])

			// Check if we've found the delimiter
			if len(buffer) >= len(delim) {
				if string(buffer[len(buffer)-len(delim):]) == string(delim) {
					return buffer, nil
				}
			}
		}
	}
}

func (m *Model) sendLine(command string) error {
	_, err := m.stdin.Write([]byte(command + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func (m *Model) recvLine() (string, error) {
	line, err := m.recvUntil([]byte("\n"))
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (m *Model) getNextMove(board string) (string, error) {
	m.recvUntil([]byte(">")) // Wait for the model to be ready
	// Send command to get the next move
	if err := m.sendLine(board); err != nil {
		println("❌ Failed to send command to model:", err.Error())
		return "", err
	}

	// Receive the next move
	move, err := m.recvLine()
	if err != nil {
		println("❌ Failed to receive move from model:", err.Error())
		return "", err
	}

	return strings.TrimSpace(move), nil
}

// applyOpening applies a predefined opening to a game
func applyPosition(g *game.Game, pos []game.Position) (err error) {
	for _, move := range pos {
		if !game.IsValidMove(g.Board, g.CurrentPlayer.Color, move) {
			return fmt.Errorf("invalid move %s for player %s", utils.PositionToAlgebraic(move), g.CurrentPlayer.Name)
		}
		// Apply the move
		g.ApplyMove(move)
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.Players, g.CurrentPlayer.Color)
		}
	}
	return
}

func playMatch(model1, model2 *Model, open []game.Position) game.Piece {
	g := game.NewGame("Model 1", "Model 2")
	if err := applyPosition(g, open); err != nil {
		println("❌ Failed to apply opening:", err.Error())
		return 0
	}

	for !game.IsGameFinished(g.Board) {
		var currentModel *Model
		if g.CurrentPlayer.Color == game.Black {
			currentModel = model1
		} else {
			currentModel = model2
		}
		// Model player's turn
		if len(game.ValidMoves(g.Board, g.CurrentPlayer.Color)) > 0 {
			move, err := currentModel.getNextMove(utils.PositionsToAlgebraic(g.History))
			if err != nil {
				println("❌ Failed to get move from model :", err.Error(), utils.PositionsToAlgebraic(g.History))
				return g.GetOtherPlayerMethod().Color
			}
			pos := utils.AlgebraicToPosition(move)
			ok := g.ApplyMove(pos)
			if !ok {
				println("❌ Invalid move received from model:", move)
				return g.GetOtherPlayerMethod().Color
			}
		} else {
			g.CurrentPlayer = g.GetOtherPlayerMethod()
		}

	}

	// Determine winner
	winner := g.GetWinnerMethod()
	return winner
}

func createModels(model1Path, model2Path string) (*Model, *Model, error) {
	// Create model 1
	exec1 := exec.Command(model1Path)
	stdin1, err := exec1.StdinPipe()
	if err != nil {
		println("❌ Failed to get stdin for model 1:", err.Error())
		return nil, nil, err
	}
	stdout1, err := exec1.StdoutPipe()
	if err != nil {
		println("❌ Failed to get stdout for model 1:", err.Error())
		return nil, nil, err
	}
	stderr1, err := exec1.StderrPipe()
	if err != nil {
		println("❌ Failed to get stderr for model 1:", err.Error())
		return nil, nil, err
	}

	model1Instance := &Model{
		cmd:    exec1,
		stdin:  stdin1,
		stdout: stdout1,
		stderr: stderr1,
	}

	if err := exec1.Start(); err != nil {
		println("❌ Failed to start model 1:", err.Error())
		return nil, nil, err
	}

	// Create model 2
	exec2 := exec.Command(model2Path)
	stdin2, err := exec2.StdinPipe()
	if err != nil {
		println("❌ Failed to get stdin for model 2:", err.Error())
		return nil, nil, err
	}
	stdout2, err := exec2.StdoutPipe()
	if err != nil {
		println("❌ Failed to get stdout for model 2:", err.Error())
		return nil, nil, err
	}
	stderr2, err := exec2.StderrPipe()
	if err != nil {
		println("❌ Failed to get stderr for model 2:", err.Error())
		return nil, nil, err
	}

	model2Instance := &Model{
		cmd:    exec2,
		stdin:  stdin2,
		stdout: stdout2,
		stderr: stderr2,
	}

	if err := exec2.Start(); err != nil {
		println("❌ Failed to start model 2:", err.Error())
		return nil, nil, err
	}

	return model1Instance, model2Instance, nil
}

func main() {
	// Parse command-line flags
	model1 := flag.String("model1", "", "CLI Executable path to first model")
	model2 := flag.String("model2", "", "CLI Executable path to second model")
	numMatches := flag.Int("matches", 10, "Number of matches to play between models (2 games per match)")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	flag.Parse()

	*numMatches = min(*numMatches, len(opening.KNOWN_OPENINGS))

	// Set max parallelism
	runtime.GOMAXPROCS(*threads)

	println("Running with", *threads, "threads")

	test1, test2, err := createModels(*model1, *model2)
	if err != nil {
		println("❌ Failed to create models:", err.Error())
		return
	}

	test1.sendLine("exit")
	test2.sendLine("exit")

	test1.cmd.Process.Kill()
	test2.cmd.Process.Kill()

	println("Models initialized successfully")
	println("Starting game comparison...")
	var wg sync.WaitGroup
	results := make([]int, *numMatches*2) // 0: draw, 1: model1 wins, 2: model2 wins
	var lock sync.Mutex

	for i := 0; i < *numMatches; i++ {
		wg.Add(1)
		go func(gameNum int) {
			defer wg.Done()

			model1Instance, model2Instance, err := createModels(*model1, *model2)
			if err != nil {
				println("❌ Failed to create models for game", gameNum, ":", err.Error())
				return
			}

			open := utils.AlgebraicToPositions(opening.KNOWN_OPENINGS[gameNum].Transcript)

			tmp := playMatch(model1Instance, model2Instance, open)
			res2 := 0
			if tmp == game.White {
				res2 = 2
			} else if tmp == game.Black {
				res2 = 1
			}
			res := playMatch(model2Instance, model1Instance, open)

			model1Instance.sendLine("exit")
			model2Instance.sendLine("exit")

			err = model1Instance.cmd.Process.Kill()
			if err != nil {
				println("❌ Failed to kill model 1 process:", err.Error())
			}
			err = model2Instance.cmd.Process.Kill()
			if err != nil {
				println("❌ Failed to kill model 2 process:", err.Error())
			}

			lock.Lock()
			results[2*gameNum] = int(res)
			results[2*gameNum+1] = res2
			lock.Unlock()
		}(i)
	}

	wg.Wait()

	// Count results
	model1Wins := 0
	model2Wins := 0
	draws := 0
	for _, result := range results {
		switch result {
		case 0:
			draws++
		case 1:
			model1Wins++
		case 2:
			model2Wins++
		}
	}

	println("Results:")
	println("Model 1 wins:", model1Wins)
	println("Model 2 wins:", model2Wins)
	println("Draws:", draws)

}
