package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/learning"
	"github.com/Coloc3G/othello-engine/test"
)

func main() {
	// Parse command line arguments
	generations := flag.Int("generations", 10, "Number of generations to train")
	populationSize := flag.Int("population", 30, "Population size")
	loadFile := flag.String("load", "", "Load existing model file")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	tournamentMode := flag.Bool("tournament", false, "Use tournament mode for training")
	compareVersions := flag.Bool("compare", false, "Compare coefficient versions")
	compareGames := flag.Int("compare-games", 200, "Number of games for version comparison")
	compareDepth := flag.Int("compare-depth", 5, "Search depth for version comparison")
	useOpenings := flag.Bool("openings", true, "Use random openings in comparisons")
	flag.Parse()

	// Set GOMAXPROCS to control parallelism
	runtime.GOMAXPROCS(*threads)

	fmt.Println("Othello AI Trainer")
	fmt.Printf("Running with %d threads\n", *threads)

	// If comparison mode is enabled, run comparison and exit
	if *compareVersions {
		if *useOpenings {
			test.CompareVersionsWithOpenings(*compareGames, *compareDepth)
		} else {
			test.CompareVersions(*compareGames, *compareDepth)
		}
		return
	}

	// Record start time
	startTime := time.Now()

	// Create trainer
	trainer := learning.NewTrainer(*populationSize)
	trainer.NumGames = 30 // Adjust as needed

	// Load existing model if specified
	if *loadFile != "" {
		fmt.Printf("Loading model from %s\n", *loadFile)
		model, err := trainer.LoadModel(*loadFile)
		if err != nil {
			fmt.Printf("Error loading model: %v\n", err)
			os.Exit(1)
		}
		trainer.BestModel = model
		trainer.Models = append(trainer.Models, model)
		fmt.Println("Model loaded successfully")
	} else {
		fmt.Println("Initializing new population")
		trainer.InitializePopulation()
	}

	// Start training
	fmt.Printf("Starting training for %d generations with population size %d\n",
		*generations, *populationSize)
	fmt.Println("Each generation will play", trainer.NumGames, "games per model")

	// Use tournament mode or standard training
	if *tournamentMode {
		fmt.Println("Using tournament mode for evaluation")
		trainer.TournamentTraining(*generations)
	} else {
		trainer.StartTraining(*generations)
	}

	// Calculate total duration
	duration := time.Since(startTime)

	// Show results
	fmt.Println("\nTraining completed")
	fmt.Printf("Total training time: %s\n", duration.Round(time.Second))
	fmt.Printf("Best model has fitness: %.2f\n", trainer.BestModel.Fitness)
	fmt.Printf("Win rate: %.2f%%\n",
		float64(trainer.BestModel.Wins)/float64(trainer.BestModel.Wins+trainer.BestModel.Losses+trainer.BestModel.Draws)*100)
	fmt.Printf("Saved best model to best_model.json\n")
}
