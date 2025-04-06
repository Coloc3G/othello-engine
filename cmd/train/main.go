package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/learning"
)

func main() {
	// Parse command-line flags
	generations := flag.Int("generations", 50, "Number of generations to run")
	populationSize := flag.Int("population", 60, "Population size")
	numGames := flag.Int("games", 10, "Number of games per model evaluation")
	depth := flag.Int("depth", 5, "Search depth for AI")
	useGPU := flag.Bool("gpu", true, "Use GPU acceleration if available")
	gpuOnly := flag.Bool("gpu-only", false, "Run in GPU-only mode (no CPU fallback)")
	tournamentMode := flag.Bool("tournament", false, "Use tournament-based evaluation")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	flag.Parse()

	// Set max parallelism
	runtime.GOMAXPROCS(*threads)
	fmt.Printf("Running with %d threads\n", *threads)

	// Create appropriate trainer
	var trainer learning.TrainerInterface

	// Check GPU availability if requested
	if *useGPU || *gpuOnly {
		fmt.Println("Attempting to use GPU acceleration...")
		gpuAvailable := evaluation.InitCUDA()

		if gpuAvailable {
			fmt.Println("GPU acceleration enabled successfully")
			gpuTrainer := learning.NewGPUTrainer(*populationSize)
			gpuTrainer.NumGames = *numGames
			gpuTrainer.MaxDepth = *depth

			// Add debug info to trainer description
			fmt.Printf("Using GPU trainer with %d games per model, search depth %d\n",
				gpuTrainer.NumGames, gpuTrainer.MaxDepth)

			trainer = gpuTrainer
		} else if *gpuOnly {
			fmt.Println("Error: GPU-only mode requested but GPU is not available")
			os.Exit(1)
		} else {
			fmt.Println("GPU not available, falling back to CPU trainer")
			cpuTrainer := learning.NewTrainer(*populationSize)
			cpuTrainer.NumGames = *numGames
			cpuTrainer.MaxDepth = *depth
			trainer = cpuTrainer
		}
	} else {
		fmt.Println("Using CPU-only trainer")
		cpuTrainer := learning.NewTrainer(*populationSize)
		cpuTrainer.NumGames = *numGames
		cpuTrainer.MaxDepth = *depth
		trainer = cpuTrainer
	}

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

	// Clean up GPU resources if applicable
	if gpuTrainer, ok := trainer.(*learning.GPUTrainer); ok {
		gpuTrainer.CleanupGPU()
	}

	fmt.Println("Training completed!")
}
