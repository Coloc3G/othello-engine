package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/stats"
	"github.com/Coloc3G/othello-engine/models/game"
	"github.com/Coloc3G/othello-engine/models/utils"
)

func applyPosition(g *game.Game, pos []game.Position) (err error) {
	for _, move := range pos {
		if !game.IsValidMove(g.Board, g.CurrentPlayer.Color, move) {
			return fmt.Errorf("invalid move %s for player %s", utils.PositionToAlgebraic(move), g.CurrentPlayer.Name)
		}
		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, move, g.CurrentPlayer.Color)
		g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
		}
	}
	return
}

func generateRandomBoard(numMoves int) (*game.Game, error) {
	g := game.NewGame("random", "v4")

	for i := 0; i < numMoves; i++ {
		validMoves := game.ValidMoves(g.Board, g.CurrentPlayer.Color)
		if len(validMoves) == 0 {
			// No valid moves, switch player
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
			validMoves = game.ValidMoves(g.Board, g.CurrentPlayer.Color)
			if len(validMoves) == 0 {
				// Game is over
				break
			}
		}

		// Choose a random valid move
		randomMove := validMoves[rand.Intn(len(validMoves))]

		// Apply the move
		g.Board, _ = game.GetNewBoardAfterMove(g.Board, randomMove, g.CurrentPlayer.Color)
		g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
	}

	return g, nil
}

func runBenchmarkWithRandomBoards(depth int8, eval evaluation.Evaluation, numBoards int, numMoves int, showStats bool) {

	totalStats := make(map[string]*stats.OperationStats)
	totalTime := time.Duration(0)

	fmt.Printf("Running benchmark with %d random boards (%d moves each)...\n", numBoards, numMoves)

	for i := 0; i < numBoards; i++ {
		g, err := generateRandomBoard(numMoves)
		if err != nil {
			fmt.Printf("Error generating random board %d: %v\n", i+1, err)
			continue
		}

		// Check if current player has valid moves
		if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
			g.CurrentPlayer = game.GetOtherPlayer(g.CurrentPlayer.Color)
			if !game.HasAnyMoves(g.Board, g.CurrentPlayer.Color) {
				fmt.Printf("Board %d: Game is over, skipping\n", i+1)
				continue
			}
		}

		var boardStats *stats.PerformanceStats
		if showStats {
			boardStats = stats.NewPerformanceStats()
		} else {
			boardStats = nil
		}

		// Get memory stats before
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		start := time.Now()
		bestMoves, score := evaluation.SolveWithStats(g.Board, g.CurrentPlayer.Color, depth, eval, boardStats)
		elapsed := time.Since(start)

		fmt.Printf("Board %d: Best move: %s, Score: %d, Time: %v\n",
			i+1, utils.PositionsToAlgebraic(bestMoves), score, elapsed)

		// Get memory stats after
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		totalTime += elapsed

		// Calculate memory usage
		allocDiff := memAfter.Alloc - memBefore.Alloc
		totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc

		fmt.Printf("Board %d: %v, Mem: %d KB allocated, %d KB total\n",
			i+1, elapsed, allocDiff/1024, totalAllocDiff/1024)

		fmt.Printf("Board %d: %v\n", i+1, elapsed)

		// Accumulate stats
		if showStats {
			for opName, opStats := range boardStats.Operations {
				if totalStats[opName] == nil {
					totalStats[opName] = &stats.OperationStats{
						Count: 0,
						Time:  0,
						Cache: make(map[string]int64),
					}
				}
				totalStats[opName].Count += opStats.Count
				totalStats[opName].Time += opStats.Time

				for hash, hits := range opStats.Cache {
					totalStats[opName].Cache[hash] += hits
				}
				boardStats.Reset()
			}
		}

	}

	fmt.Printf("\n=== AVERAGE RESULTS OVER %d BOARDS ===\n", numBoards)
	fmt.Printf("Average time: %v\n", totalTime/time.Duration(numBoards))
	fmt.Printf("Total time: %v\n", totalTime)
	if showStats {
		for opName, opStats := range totalStats {
			fmt.Printf("\nOperation: %s\n", opName)
			fmt.Printf("  Average count: %.1f\n", float64(opStats.Count)/float64(numBoards))
			fmt.Printf("  Average time: %v\n", opStats.Time/time.Duration(numBoards))

			// Sort cache hits
			type cacheStat struct {
				Hash string
				Hits int64
			}
			var cacheStatsSlice []cacheStat
			for hash, hits := range opStats.Cache {
				cacheStatsSlice = append(cacheStatsSlice, cacheStat{Hash: hash, Hits: hits})
			}

			sort.Slice(cacheStatsSlice, func(i, j int) bool {
				return cacheStatsSlice[i].Hits > cacheStatsSlice[j].Hits
			})

			fmt.Printf("  Top cache hits (total across all boards):\n")
			for _, cs := range cacheStatsSlice[:min(5, len(cacheStatsSlice))] {
				avgHits := float64(cs.Hits) / float64(numBoards)
				fmt.Printf("    Hash: %s, Total hits: %d, Avg hits: %.1f\n", cs.Hash, cs.Hits, avgHits)
			}
		}
	}
}

func main() {
	d := flag.Int("depth", 10, "Search depth for evaluation")
	showStats := flag.Bool("stats", false, "Show perf stats")
	randomBoards := flag.Int("random", 0, "Number of random boards to test (0 = use fixed board)")
	randomMoves := flag.Int("moves", 20, "Number of random moves for random board generation")
	flag.Parse()

	depth := int8(*d)
	eval := evaluation.NewMixedEvaluation(evaluation.V4Coeff)

	if *randomBoards > 0 {
		runBenchmarkWithRandomBoards(depth, eval, *randomBoards, *randomMoves, *showStats)
		return
	}

	// Original fixed board logic
	g, err := generateRandomBoard(*randomMoves)
	if err != nil {
		fmt.Println("Error generating random board:", err)
		return
	}

	start := time.Now()
	if *showStats {
		stats := stats.NewPerformanceStats()
		bestMoves, score := evaluation.SolveWithStats(g.Board, g.CurrentPlayer.Color, depth, eval, stats)
		if len(bestMoves) == 0 || (len(bestMoves) == 1 && bestMoves[0].Row == -1 && bestMoves[0].Col == -1) {
			fmt.Println("No valid moves found")
			return
		}
		fmt.Println("Evaluation with stats completed in:", time.Since(start))
		fmt.Printf("Best move: %s, Score: %d\n", utils.PositionsToAlgebraic(bestMoves), score)
		fmt.Printf("Performance stats: \n")
		for name, op := range stats.Operations {
			fmt.Printf("Operation: %s, Count: %d, Time: %s\n", name, op.Count, op.Time)
			// Sort descending by Hits
			cachesStats := op.Cache
			// Convert map to slice for sorting
			type cacheStat struct {
				Hash string
				Hits int64
			}
			var cacheStatsSlice []cacheStat
			for hash, stat := range cachesStats {
				cacheStatsSlice = append(cacheStatsSlice, cacheStat{Hash: hash, Hits: stat})
			}
			// Sort descending by Hits
			sort.Slice(cacheStatsSlice, func(i, j int) bool {
				return cacheStatsSlice[i].Hits > cacheStatsSlice[j].Hits
			})
			// Print sorted cache stats
			for _, cs := range cacheStatsSlice[:min(10, len(cacheStatsSlice))] {
				fmt.Printf("  Hash: %s, Hits: %d\n", cs.Hash, cs.Hits)
			}
		}
	} else {
		bestMoves, score := evaluation.Solve(g.Board, g.CurrentPlayer.Color, depth, eval)
		if len(bestMoves) == 0 || (len(bestMoves) == 1 && bestMoves[0].Row == -1 && bestMoves[0].Col == -1) {
			fmt.Println("No valid moves found")
			return
		}
		fmt.Println("Evaluation completed in:", time.Since(start))
		fmt.Printf("Best moves: %s, Score: %d\n", utils.PositionsToAlgebraic(bestMoves), score)
	}
}
