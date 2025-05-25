package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/Coloc3G/othello-engine/models/ai/learning"
)

func main() {
	// Parse command-line flags
	generations := flag.Int("generations", 50, "Number of generations to run")
	populationSize := flag.Int("population", 60, "Population size")
	numGames := flag.Int("games", 20, "Number of games per model evaluation")
	depth := flag.Int("depth", 5, "Search depth for AI")
	tournamentMode := flag.Bool("tournament", false, "Use tournament-based evaluation")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	flag.Parse()

	// Set max parallelism
	runtime.GOMAXPROCS(*threads)
	fmt.Printf("Running with %d threads\n", *threads)

	// Create appropriate trainer
	var trainer learning.TrainerInterface
	cpuTrainer := learning.NewTrainer(*populationSize)
	cpuTrainer.NumGames = *numGames
	cpuTrainer.MaxDepth = *depth
	trainer = cpuTrainer

	// Print training configuration
	fmt.Println("Othello AI Trainer")
	fmt.Printf("Initializing new population\n")
	trainer.InitializePopulation()

	fmt.Printf("Starting training for %d generations with population size %d, playing %d games per match\n\n",
		*generations, *populationSize, *numGames)

	// Run training
	if *tournamentMode {
		fmt.Println("Using tournament mode for evaluation")
		trainer.TournamentTraining(*generations)
	} else {
		trainer.StartTraining(*generations)
	}

	fmt.Println("Training completed!")
}
