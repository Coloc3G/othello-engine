package main

import (
	"flag"
	"fmt"
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

func main() {

	d := flag.Int("depth", 10, "Search depth for evaluation")
	showStats := flag.Bool("stats", false, "Show perf stats")
	flag.Parse()

	board := "c4c3d3e3e2c5f3c2b6c6b4b5d2d1e1f1d6e6a4d7c1b1f2c7f5a3a2b3f4g5g4h4d8g1e8g3g2b8f7g8h5h6"

	// Initialize the game board
	g := game.NewGame("test", "v4")
	pos := utils.AlgebraicToPositions(board)
	err := applyPosition(g, pos)
	if err != nil {
		fmt.Println(err)
		return
	}

	depth := int8(*d)

	eval := evaluation.NewMixedEvaluation(evaluation.V4Coeff)
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

// Depth 6

// Evaluation completed in: 667.1792ms
// Best move: d7, Score: 129
// Performance stats:
// Operation: hashBoard, Count: 140317, Time: 7.6103ms
//   Hash: 633dd9246f636c7f, Hits: 11
//   Hash: -12e2d815610cd9f9, Hits: 10
//   Hash: -4c4eb6474438d241, Hits: 10
//   Hash: -a8193ddad7b559d, Hits: 10
//   Hash: -77399087cb0c339d, Hits: 10
//   Hash: -6c3addac13d54e7d, Hits: 10
//   Hash: -69e79b15278eb425, Hits: 10
//   Hash: -52e013fd411ef9d, Hits: 9
//   Hash: -55230e6f06ab7a0, Hits: 9
//   Hash: -474a64239cc8c445, Hits: 9
// Operation: pec, Count: 140317, Time: 502.2363ms
//   Hash: 633dd9246f636c7f, Hits: 11
//   Hash: -69e79b15278eb425, Hits: 10
//   Hash: -6c3addac13d54e7d, Hits: 10
//   Hash: -4c4eb6474438d241, Hits: 10
//   Hash: -77399087cb0c339d, Hits: 10
//   Hash: -a8193ddad7b559d, Hits: 10
//   Hash: -12e2d815610cd9f9, Hits: 10
//   Hash: -55230e6f06ab7a0, Hits: 9
//   Hash: -474a64239cc8c445, Hits: 9
//   Hash: -52e013fd411ef9d, Hits: 9
// Operation: leaf_eval, Count: 111546, Time: 54.0034ms
//   Hash: 633dd9246f636c7f, Hits: 11
//   Hash: -a8193ddad7b559d, Hits: 10
//   Hash: -69e79b15278eb425, Hits: 10
//   Hash: -4c4eb6474438d241, Hits: 10
//   Hash: -77399087cb0c339d, Hits: 10
//   Hash: -12e2d815610cd9f9, Hits: 10
//   Hash: -6c3addac13d54e7d, Hits: 10
//   Hash: ed29c766a78c2a0, Hits: 9
//   Hash: -52e013fd411ef9d, Hits: 9
//   Hash: -55230e6f06ab7a0, Hits: 9

// Evaluation completed in: 743.7956ms
// Best move: d7, Score: 129
// Performance stats:
// Operation: leaf_eval, Count: 90963, Time: 45.5847ms
//   Hash: 000023623c38381001261c1c42450000, Hits: 1
//   Hash: 00003b72ec0404240106040c12391008, Hits: 1
//   Hash: 00003b320e2c30200106844c70100814, Hits: 1
//   Hash: 00083762548804000106081c2a751200, Hits: 1
//   Hash: 00002f1628303008050e5068570c0000, Hits: 1
//   Hash: 000001506c702000011e7e2f120c1000, Hits: 1
//   Hash: 00083d324e8402000107c24c30381000, Hits: 1
//   Hash: 0804392008801008010206de767e0000, Hits: 1
//   Hash: 000023723c383810012e1c0c42050000, Hits: 1
//   Hash: 000837620c1020100106081c726d1008, Hits: 1
// Operation: prune, Count: 23113, Time: 0s
//   Hash: , Hits: 23113
// Operation: pec_cache_hit, Count: 24651, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083f726c8000200106000c123d1408, Hits: 8
//   Hash: 00083d70688002000106020f163d1400, Hits: 8
//   Hash: 00083d722c3030200107020c520d0400, Hits: 8
// Operation: leaf_eval_cache_hit, Count: 20583, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00083d722c3030200107020c520d0400, Hits: 8
//   Hash: 00001b62303830100146241c4e050404, Hits: 8
//   Hash: 00083f726c8000200106000c123d1408, Hits: 8
// Operation: hashBoard, Count: 140317, Time: 39.5891ms
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 11
//   Hash: 00083d726e8000200107020c103c1408, Hits: 10
//   Hash: 00003972383830100107060c46050404, Hits: 10
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 10
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 10
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 10
//   Hash: 00083f726c8000200106000c123d1408, Hits: 9
//   Hash: 00001b62303830100146241c4e050404, Hits: 9
//   Hash: 00083d70688002000106020f163d1400, Hits: 9
// Operation: pec, Count: 115666, Time: 478.4584ms
//   Hash: 202029722e3030080107160c500c0000, Hits: 1
//   Hash: 04023b00e0200008010404fe1f1c1010, Hits: 1
//   Hash: 0008bf4226200000010600bc581c3040, Hits: 1
//   Hash: 040627766eb0200001301808100c1000, Hits: 1
//   Hash: 00203920243820000146065eda061200, Hits: 1
//   Hash: 00083f0a3f204000010600f4401d1000, Hits: 1
//   Hash: 00003b320c0201200106844c723d1010, Hits: 1
//   Hash: 00000bd26e3030080126742c100c4000, Hits: 1
//   Hash: 000033626636000001060c1c1848b840, Hits: 1
//   Hash: 08043b763c3438000102040842090402, Hits: 1

// Evaluation completed in: 678.6387ms
// Best move: d7, Score: 129
// Performance stats:
// Operation: move, Count: 123429, Time: 25.5438ms
//   Hash: a1-00003b224ea42200010604dc30181800, Hits: 1
//   Hash: d7-00003970684412080106060e163a4000, Hits: 1
//   Hash: d2-00087f62200810200146001c5f340000, Hits: 1
//   Hash: b8-08040950181412100122362e662a0000, Hits: 1
//   Hash: a4-00001962e63030100147261c180c0000, Hits: 1
//   Hash: d7-00003b72fc1000100106040c022d1008, Hits: 1
//   Hash: e1-00003b42ec383800010604bc12050000, Hits: 1
//   Hash: a2-08043b226e820000010204dd103c1000, Hits: 1
//   Hash: a7-0000235f02141210010e1c207d280000, Hits: 1
//   Hash: e7-00000342361c1410014e3c3c48200000, Hits: 1
// Operation: leaf_eval, Count: 90963, Time: 45.8528ms
//   Hash: 000003323414041001067c4c4a291808, Hits: 1
//   Hash: 00002972e63620000127160c18081800, Hits: 1
//   Hash: 000023626e70200801161c1c100c5020, Hits: 1
//   Hash: 00000352cf201008012e3c2c301d0000, Hits: 1
//   Hash: 00003104ce21100801060efb305c0000, Hits: 1
//   Hash: 00101f662e302000094e2018504c1000, Hits: 1
//   Hash: 040438703c101010110b070f422c0000, Hits: 1
//   Hash: 00002742ee800000011e58bc103c1000, Hits: 1
//   Hash: 0000236a1e0e020001161c1460701404, Hits: 1
//   Hash: 0000351038001410090e4a6e463e0000, Hits: 1
// Operation: prune, Count: 23113, Time: 0s
//   Hash: , Hits: 23113
// Operation: pec_cache_hit, Count: 24651, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 00083d70688002000106020f163d1400, Hits: 8
//   Hash: 00001b62303830100146241c4e050404, Hits: 8
//   Hash: 00083d722c3030200107020c520d0400, Hits: 8
// Operation: leaf_eval_cache_hit, Count: 20583, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00083d70688002000106020f163d1400, Hits: 8
//   Hash: 00083f726c8000200106000c123d1408, Hits: 8
//   Hash: 00001b62303830100146241c4e050404, Hits: 8
// Operation: move_cache_hit, Count: 16875, Time: 0s
//   Hash: g1-00083d022e303010010702fc500c0000, Hits: 5
//   Hash: d2-000039723c3c34100107060c42010000, Hits: 5
//   Hash: h4-020e3f72688000000100000c163f1000, Hits: 5
//   Hash: h3-080c3f72684000000102000c163f1000, Hits: 5
//   Hash: h6-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: e7-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: a5-080c3f72e80000000102000c163f1000, Hits: 5
//   Hash: g1-000039723c3c34100107060c42010000, Hits: 5
//   Hash: c8-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: g7-080c3f72688000000102000c163f1000, Hits: 5
// Operation: hashBoard, Count: 140317, Time: 51.8696ms
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 11
//   Hash: 00083d726e8000200107020c103c1408, Hits: 10
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 10
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 10
//   Hash: 00003972383830100107060c46050404, Hits: 10
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 10
//   Hash: 00001b62303830100146241c4e050404, Hits: 9
//   Hash: 00083d722c3030200107020c520d0400, Hits: 9
//   Hash: 00083d70688002000106020f163d1400, Hits: 9
// Operation: pec, Count: 115666, Time: 293.0615ms
//   Hash: 00001b02263e3008014624fc58000000, Hits: 1
//   Hash: 040637664e40000001000818303cd080, Hits: 1
//   Hash: 00083978281808000106060757241404, Hits: 1
//   Hash: 08043974321100100103060a4d2c1008, Hits: 1
//   Hash: 00002fd608302000050e1028f70c1000, Hits: 1
//   Hash: 000023523c1c1810010e1c2c42210204, Hits: 1
//   Hash: 00083f6a5af800000106001424041404, Hits: 1
//   Hash: 000023022e34301001161cfc50080402, Hits: 1
//   Hash: 02063b324e8800200100844c30341008, Hits: 1
//   Hash: 0000276a64800000011e18141b3f1000, Hits: 1

// Evaluation completed in: 499.7475ms
// Best move: d7, Score: 129
// Performance stats:
// Operation: pec, Count: 115666, Time: 213.8673ms
//   Hash: 000037623e3c2400090e081c40401000, Hits: 1
//   Hash: 00000b120c1c3420012634ec72210000, Hits: 1
//   Hash: 00083d702c3030200126020f520c0400, Hits: 1
//   Hash: 00083f06341c0410010600f84b201008, Hits: 1
//   Hash: 000021023e1c1610010f1efc40200000, Hits: 1
//   Hash: 00003b760e90202001060408f02c1808, Hits: 1
//   Hash: 00003b72727a00000106040c0c051c04, Hits: 1
//   Hash: 00003b70201e00000106040e5f207000, Hits: 1
//   Hash: 0804310048a0100801020efe365e0000, Hits: 1
//   Hash: 00080702061010100106787c786c0201, Hits: 1
// Operation: move, Count: 123429, Time: 6.9999ms
//   Hash: e1-08043b340e0910200102844b70340000, Hits: 1
//   Hash: h5-00001b6a3c3838100146241442050000, Hits: 1
//   Hash: d7-00003b324e8800200106844c30341010, Hits: 1
//   Hash: b1-00001b662e3434240146241850080000, Hits: 1
//   Hash: b1-081423623634240001021c1c48081800, Hits: 1
//   Hash: g6-00083f6ac0201008010600143f5c0000, Hits: 1
//   Hash: b3-000837622e3020400106081c504c1010, Hits: 1
//   Hash: f1-00003b2266a01010010604dc181c2040, Hits: 1
//   Hash: b2-00003870fc1010100107070f022c0000, Hits: 1
//   Hash: f1-00103b621a1810100106041c64640404, Hits: 1
// Operation: leaf_eval, Count: 90963, Time: 17.006ms
//   Hash: 000029027e703010012716fc000c0000, Hits: 1
//   Hash: 0000034216181810012e3c3c68240201, Hits: 1
//   Hash: 00001921371c14100146665e48200000, Hits: 1
//   Hash: 00000b023f1c0c00012634fc40211000, Hits: 1
//   Hash: 0000136266704080014e2c1c180c3000, Hits: 1
//   Hash: 000039706c0030200106060f12fc4000, Hits: 1
//   Hash: 0000237266a4020001361c0c18183040, Hits: 1
//   Hash: 00003930cc0402010106864e323b1000, Hits: 1
//   Hash: 000003023e1c041c01067cfc40201000, Hits: 1
//   Hash: 000033623624123001060c1c48582040, Hits: 1
// Operation: prune, Count: 23113, Time: 0s
//   Hash: , Hits: 23113
// Operation: pec_cache_hit, Count: 24651, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 00083d722c3030200107020c520d0400, Hits: 8
//   Hash: 00083d70688002000106020f163d1400, Hits: 8
//   Hash: 00001b62303830100146241c4e050404, Hits: 8
// Operation: leaf_eval_cache_hit, Count: 20583, Time: 0s
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 9
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 9
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 9
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 9
//   Hash: 00003972383830100107060c46050404, Hits: 9
//   Hash: 00083d726e8000200107020c103c1408, Hits: 9
//   Hash: 00001b62303830100146241c4e050404, Hits: 8
//   Hash: 00083d70688002000106020f163d1400, Hits: 8
//   Hash: 00083f726c8000200106000c123d1408, Hits: 8
// Operation: move_cache_hit, Count: 16875, Time: 0s
//   Hash: g1-000039723c3c34100107060c42010000, Hits: 5
//   Hash: f7-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: a5-020e3f72e80000000100000c163f1000, Hits: 5
//   Hash: h6-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: h4-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: f1-00083d022e303010010702fc500c0000, Hits: 5
//   Hash: a7-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: c8-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: e7-080c3f72688000000102000c163f1000, Hits: 5
//   Hash: h3-080c3f72684000000102000c163f1000, Hits: 5
// Operation: hashBoard, Count: 140317, Time: 28.4159ms
//   Hash: 00083d7264a000000107020c1a1d3040, Hits: 11
//   Hash: 00083d72242010080107020c5a1d2040, Hits: 10
//   Hash: 00083d726e8000200107020c103c1408, Hits: 10
//   Hash: 00083f7260a000000106000c1e1f3040, Hits: 10
//   Hash: 080c3f022e000020010200fc503c1408, Hits: 10
//   Hash: 00083d72ee0000200107020c103c1408, Hits: 10
//   Hash: 00003972383830100107060c46050404, Hits: 10
//   Hash: 00083d70688002000106020f163d1400, Hits: 9
//   Hash: 00083d722c3030200107020c520d0400, Hits: 9
//   Hash: 00001b62303830100146241c4e050404, Hits: 9
