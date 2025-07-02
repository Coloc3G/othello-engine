package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Coloc3G/othello-engine/models/opening"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Results structure to hold performance data
type PerformanceResult struct {
	Version1Name   string
	Version2Name   string
	Version1Wins   int
	Version2Wins   int
	Draws          int
	TotalGames     int
	Version1WinPct float64
	Version2WinPct float64
	DrawPct        float64
}

func main() {
	// Parse command line flags
	numGames := flag.Int("games", 200, "Number of games to run for each comparison")
	searchDepth := flag.Int("depth", 5, "Search depth for AI")
	generateHTML := flag.Bool("html", false, "Generate HTML visualization files")
	showASCII := flag.Bool("ascii", true, "Show ASCII visualization in terminal")
	flag.Parse()

	if *numGames > len(opening.KNOWN_OPENINGS) {
		*numGames = len(opening.KNOWN_OPENINGS)
	}

	searchDepth8 := int8(*searchDepth)

	fmt.Println("Othello AI Performance Visualization")
	fmt.Printf("Running with %d matches (2 matches/game) at depth %d\n", *numGames*2, searchDepth8)

	// Run all comparisons and generate results
	results := runAllComparisons(*numGames, searchDepth8)

	// Create visualizations
	if *generateHTML {
		generateHTMLVisualizations(results)
	}

	if *showASCII {
		showASCIIVisualizations(results)
	}
}

// runAllComparisons runs all comparisons and returns results
func runAllComparisons(numGames int, searchDepth int8) []PerformanceResult {
	// Compare V1 vs V2
	results := CompareVersions(numGames, searchDepth)
	return results
}

// generateHTMLVisualizations creates HTML visualizations of the results
func generateHTMLVisualizations(results []PerformanceResult) {
	// Create a new page
	page := components.NewPage()
	page.PageTitle = "Othello AI Performance Comparison"
	page.Layout = components.PageFlexLayout

	// Add bar chart for each comparison
	for _, result := range results {
		// Create a horizontal bar chart
		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    fmt.Sprintf("Comparison: %s vs %s", result.Version1Name, result.Version2Name),
				Subtitle: fmt.Sprintf("Total Games: %d", result.TotalGames),
			}),
			charts.WithTooltipOpts(opts.Tooltip{}),
			charts.WithLegendOpts(opts.Legend{Right: "10%"}),
			charts.WithColorsOpts(opts.Colors{"#2f4554", "#61a0a8", "#d48265"}),
		)

		// Set x axis with categories
		bar.SetXAxis([]string{result.Version1Name, result.Version2Name})

		// Add data for wins, draws, losses
		bar.AddSeries("Wins", []opts.BarData{
			{Value: result.Version1Wins},
			{Value: result.Version2Wins},
		})
		bar.AddSeries("Draws", []opts.BarData{
			{Value: result.Draws},
			{Value: result.Draws},
		})
		bar.AddSeries("Losses", []opts.BarData{
			{Value: result.Version2Wins},
			{Value: result.Version1Wins},
		})

		// Set series options
		bar.SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{Position: "right"}),
		)

		// Add chart to page
		page.AddCharts(bar)

		// Add percentage bar chart
		percentBar := charts.NewBar()
		percentBar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: fmt.Sprintf("Win Percentages: %s vs %s", result.Version1Name, result.Version2Name),
			}),
			charts.WithTooltipOpts(opts.Tooltip{}),
			charts.WithLegendOpts(opts.Legend{Right: "10%"}),
			charts.WithColorsOpts(opts.Colors{"#5470c6", "#91cc75", "#fac858"}),
		)

		// Set x axis with categories
		percentBar.SetXAxis([]string{result.Version1Name, result.Version2Name})

		// Add data for win percentages
		percentBar.AddSeries("Win %", []opts.BarData{
			{Value: result.Version1WinPct},
			{Value: result.Version2WinPct},
		})
		percentBar.AddSeries("Draw %", []opts.BarData{
			{Value: result.DrawPct},
			{Value: result.DrawPct},
		})

		// Set series options
		percentBar.SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{Position: "right"}),
		)

		// Add chart to page
		page.AddCharts(percentBar)
	}

	// Create timestamp for filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("performance_visualization_%s.html", timestamp)

	// Create file
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	// Render the page to file
	err = page.Render(f)
	if err != nil {
		log.Fatalf("Failed to render chart: %v", err)
	}

	fmt.Printf("Visualization saved to %s\n", filename)
}

// showASCIIVisualizations displays ASCII visualizations in the terminal
func showASCIIVisualizations(results []PerformanceResult) {
	fmt.Println("\n===== ASCII Visualization =====")

	// For each comparison
	for _, result := range results {
		fmt.Printf("\n%s vs %s Comparison (Total: %d games)\n\n",
			result.Version1Name, result.Version2Name, result.TotalGames)

		// Calculate the max value for scaling
		maxValue := max(max(result.Version1Wins, result.Version2Wins), result.Draws)
		scale := 50.0 / float64(maxValue) // Scale to fit in 50 chars width

		// Draw the bars
		drawASCIIBar(fmt.Sprintf("%s Wins", result.Version1Name), result.Version1Wins, scale)
		drawASCIIBar("Draws", result.Draws, scale)
		drawASCIIBar(fmt.Sprintf("%s Wins", result.Version2Name), result.Version2Wins, scale)

		// Show percentages
		fmt.Printf("\nWin Percentages:\n")
		fmt.Printf("%s: %.1f%%  |  Draw: %.1f%%  |  %s: %.1f%%\n",
			result.Version1Name, result.Version1WinPct,
			result.DrawPct,
			result.Version2Name, result.Version2WinPct)

		fmt.Println(strings.Repeat("-", 60))

		if result.Version1Wins > result.Version2Wins {
			diff := result.Version1Wins - result.Version2Wins
			fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
				result.Version1Name, diff, float64(diff)*100.0/float64(result.TotalGames))
		} else if result.Version2Wins > result.Version1Wins {
			diff := result.Version2Wins - result.Version1Wins
			fmt.Printf("%s is stronger by %d games (%.1f%% margin)\n",
				result.Version2Name, diff, float64(diff)*100.0/float64(result.TotalGames))
		} else {
			fmt.Println("Both versions appear equally matched")
		}
	}

}

// drawASCIIBar draws a simple ASCII bar with label
func drawASCIIBar(label string, value int, scale float64) {
	barWidth := int(float64(value) * scale)
	fmt.Printf("%-10s [%s%s] %d\n",
		label,
		strings.Repeat("█", barWidth),
		strings.Repeat(" ", 50-barWidth),
		value)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
