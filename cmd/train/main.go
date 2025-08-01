package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/learning"
)

func main() {
	// Parse command-line flags
	generations := flag.Int("generations", 50, "Number of generations to run")
	populationSize := flag.Int("population", 50, "Population size")
	numGames := flag.Int("games", 20, "Number of games per model evaluation")
	depth := flag.Int("depth", 5, "Search depth for AI")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	baseModel := flag.String("base", "V1", "Base model to use for training (default: V1)")
	modelName := flag.String("name", "", "Name of the model to save after training")
	flag.Parse()

	if *modelName == "" {
		fmt.Println("Please provide a name for the model using the -name flag.")
		flag.Usage()
		return
	}

	// Set max parallelism
	runtime.GOMAXPROCS(*threads)
	fmt.Printf("Running with %d threads\n", *threads)

	baseModelCoeffs, found := evaluation.GetCoefficientsByName(*baseModel)
	if !found {
		fmt.Printf("Base model '%s' not found. Available models: ", *baseModel)
		for _, model := range evaluation.Models {
			fmt.Printf("%s ", model.Name)
		}
		fmt.Println()
		return
	}

	// Create appropriate trainer
	trainer := learning.NewTrainer(*modelName, *populationSize, *numGames, int8(*depth), baseModelCoeffs)

	// Print training configuration
	fmt.Println("Othello AI Trainer")
	fmt.Printf("Starting training for %d generations with population size %d, playing %d matches\n\n",
		*generations, *populationSize, *numGames)
	trainer.StartTraining(*generations)
}
