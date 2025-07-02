package main

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/learning"
	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/schollz/progressbar/v3"
)

// CompareCoefficients compares two sets of evaluation coefficients concurrently
func CompareCoefficients(coeff1, coeff2 evaluation.EvaluationCoefficients, numGames int, searchDepth int8) PerformanceResult {

	selectedOpenings := opening.SelectRandomOpenings(numGames)
	numGames = len(selectedOpenings)

	// Create stats object
	stats := PerformanceResult{
		Version1Name: coeff1.Name,
		Version2Name: coeff2.Name,
		TotalGames:   numGames * 2,
	}

	// Create two evaluation functions with different coefficients
	eval1 := evaluation.NewMixedEvaluation(coeff1)
	eval2 := evaluation.NewMixedEvaluation(coeff2)

	// Create progress bar
	bar := progressbar.NewOptions(numGames*2,
		progressbar.OptionSetDescription("Playing games"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()

	// Set up job and result channels and a worker pool
	jobsCh := make(chan int, numGames)
	resultsCh := make(chan int, numGames*2) // Buffer for all results

	for i := range numGames {
		jobsCh <- i
	}
	close(jobsCh)

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobsCh {
				for index := range 2 {
					win1, win2, draw := learning.PlayMatchWithOpening(eval1, eval2, selectedOpenings[i], index, searchDepth)
					bar.Add(1)
					if win1 {
						resultsCh <- 1
					} else if win2 {
						resultsCh <- 2
					} else if draw {
						resultsCh <- 0
					}
				}
			}
		}()
	}

	// Close results channel after all workers are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results without additional progress bar updates
	for outcome := range resultsCh {
		switch outcome {
		case 0:
			stats.Draws++
		case 1:
			stats.Version1Wins++
		default:
			stats.Version2Wins++
		}
	}

	// Calculate percentages
	stats.Version1WinPct = float64(stats.Version1Wins) * 100.0 / float64(numGames*2)
	stats.Version2WinPct = float64(stats.Version2Wins) * 100.0 / float64(numGames*2)
	stats.DrawPct = float64(stats.Draws) * 100.0 / float64(numGames*2)

	return stats
}

// PrintComparison prints comparison statistics
func PrintComparison(stats PerformanceResult) {
	fmt.Println("\n==== Comparison Results ====")
	fmt.Printf("%s vs %s (Total games: %d)\n", stats.Version1Name, stats.Version2Name, stats.TotalGames)
	fmt.Printf("%s wins: %d (%.1f%%)\n", stats.Version1Name, stats.Version1Wins, stats.Version1WinPct)
	fmt.Printf("%s wins: %d (%.1f%%)\n", stats.Version2Name, stats.Version2Wins, stats.Version2WinPct)
	fmt.Printf("Draws: %d (%.1f%%)\n", stats.Draws, stats.DrawPct)

	// Print summary judgment
	fmt.Print("Conclusion: ")
	if stats.Version1Wins > stats.Version2Wins {
		diff := stats.Version1Wins - stats.Version2Wins
		fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
			stats.Version1Name, diff, float64(diff)*100.0/float64(stats.TotalGames))
	} else if stats.Version2Wins > stats.Version1Wins {
		diff := stats.Version2Wins - stats.Version1Wins
		fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
			stats.Version2Name, diff, float64(diff)*100.0/float64(stats.TotalGames))
	} else {
		fmt.Println("Both versions appear equally matched")
	}
	fmt.Println("===========================")
}

func CompareVersions(numGames int, searchDepth int8) (results []PerformanceResult) {

	for i := range evaluation.Models {
		for j := range evaluation.Models {
			if i >= j {
				continue // Avoid duplicate comparisons and self-comparisons
			}

			// Compare each model with every other model
			res := CompareCoefficients(evaluation.Models[i], evaluation.Models[j], numGames, searchDepth)
			results = append(results, res)
		}
	}

	return
}
