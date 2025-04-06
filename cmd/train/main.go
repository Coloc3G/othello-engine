package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Coloc3G/othello-engine/models/ai/evaluation"
	"github.com/Coloc3G/othello-engine/models/ai/learning"
	"github.com/Coloc3G/othello-engine/models/opening"
)

// We're keeping this simple since we don't have the full implementation
var gpuOnlyMode bool = false

func main() {
	// Define command line flags
	benchmarkPtr := flag.Bool("benchmark", false, "Run GPU benchmark")
	gpuOnlyPtr := flag.Bool("gpu-only", false, "Use GPU only (no fallback to CPU)")

	// Training parameters
	generations := flag.Int("generations", 10, "Number of generations to train")
	populationSize := flag.Int("population", 30, "Population size")
	loadFile := flag.String("load", "", "Load existing model file")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads to use")
	tournamentMode := flag.Bool("tournament", false, "Use tournament mode for training")
	numGames := flag.Int("games", len(opening.KNOWN_OPENINGS), "Number of games to play between models (max: number of openings)")

	// Force CPU mode
	cpuOnly := flag.Bool("cpu-only", false, "Force CPU-only mode even if GPU is available")

	flag.Parse()

	// Run benchmark if requested
	if *benchmarkPtr {
		fmt.Println("Running GPU benchmarks...")
		evaluation.RunGPUBenchmark()
		return
	}

	// Configure GPU usage
	if *gpuOnlyPtr {
		fmt.Println("Running in GPU-only mode (no CPU fallback)")
		// Set a global flag in the evaluation package
		gpuOnlyMode = true
	}

	// Initialize CUDA early to verify GPU availability
	gpuAvailable := evaluation.InitCUDA()
	if gpuAvailable {
		fmt.Println("GPU acceleration is enabled")

		// Clean up CUDA resources when the program exits
		defer evaluation.CleanupCUDA()
	} else {
		fmt.Println("GPU acceleration is not available - using CPU only")

		// If GPU-only mode was requested but GPU is not available, exit
		if *gpuOnlyPtr {
			fmt.Println("Error: GPU-only mode specified but GPU is not available")
			os.Exit(1)
		}
	}

	// Ensure CUDA DLL file is properly located
	ensureCUDAFiles()

	// Set GOMAXPROCS to control parallelism
	runtime.GOMAXPROCS(*threads)

	// Run the trainer
	runTrainer(*generations, *populationSize, *loadFile, *threads, *tournamentMode, *cpuOnly, *numGames)
}

// runTrainer runs the training process
func runTrainer(generations int, populationSize int, loadFile string, threads int, tournamentMode bool, cpuOnly bool, numGames int) {
	fmt.Println("Othello AI Trainer")
	fmt.Printf("Running with %d threads\n", threads)

	// Record start time
	startTime := time.Now()

	var trainer interface {
		InitializePopulation()
		StartTraining(int)
		TournamentTraining(int)
		LoadModel(string) (learning.EvaluationModel, error)
	}

	// Create GPU-enabled trainer if requested and available
	useGPU := !cpuOnly

	if useGPU {
		fmt.Println("Attempting to use GPU acceleration...")
		gpuAvailable := ensureGPUSupport()

		if gpuAvailable {
			fmt.Println("GPU acceleration enabled successfully")
			gpuTrainer := learning.NewGPUTrainer(populationSize)
			gpuTrainer.NumGames = numGames // Set number of games
			trainer = gpuTrainer

			// Load existing model if specified
			if loadFile != "" {
				fmt.Printf("Loading model from %s\n", loadFile)
				model, err := gpuTrainer.LoadModel(loadFile)
				if err != nil {
					fmt.Printf("Error loading model: %v\n", err)
					os.Exit(1)
				}
				gpuTrainer.BestModel = model
				gpuTrainer.Models = append(gpuTrainer.Models, model)
				fmt.Println("Model loaded successfully")
			} else {
				fmt.Println("Initializing new population")
				gpuTrainer.InitializePopulation()
			}
		} else {
			useGPU = false
			fmt.Println("GPU acceleration not available, falling back to CPU")
		}
	}

	// Fall back to CPU if GPU not available or disabled
	if !useGPU {
		fmt.Println("Using CPU for training")
		cpuTrainer := learning.NewTrainer(populationSize)
		cpuTrainer.NumGames = numGames // Set number of games
		trainer = cpuTrainer

		// Load existing model if specified
		if loadFile != "" {
			fmt.Printf("Loading model from %s\n", loadFile)
			model, err := trainer.LoadModel(loadFile)
			if err != nil {
				fmt.Printf("Error loading model: %v\n", err)
				os.Exit(1)
			}
			cpuTrainer.BestModel = model
			cpuTrainer.Models = append(cpuTrainer.Models, model)
			fmt.Println("Model loaded successfully")
		} else {
			fmt.Println("Initializing new population")
			trainer.InitializePopulation()
		}
	}

	// Start training
	fmt.Printf("Starting training for %d generations with population size %d, playing %d games per match\n",
		generations, populationSize, numGames)

	// Use tournament mode or standard training
	if tournamentMode {
		fmt.Println("Using tournament mode for evaluation")
		trainer.TournamentTraining(generations)
	} else {
		trainer.StartTraining(generations)
	}

	// Calculate total duration
	duration := time.Since(startTime)

	// Show results
	fmt.Println("\nTraining completed")
	fmt.Printf("Total training time: %s\n", duration.Round(time.Second))
}

// ensureCUDAFiles ensures the CUDA DLL files are in the right locations
func ensureCUDAFiles() {
	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		// In development mode, use current working directory
		execPath, _ = os.Getwd()
	} else {
		execPath = filepath.Dir(execPath)
	}

	// Create cuda directory if it doesn't exist
	cudaDir := filepath.Join(execPath, "cuda")
	os.MkdirAll(cudaDir, 0755)

	// Check if DLL is in root directory and copy it to cuda dir if needed
	dllPath := filepath.Join(execPath, "cuda_othello.dll")
	if _, err := os.Stat(dllPath); err == nil {
		cudaDllPath := filepath.Join(cudaDir, "cuda_othello.dll")
		if _, err := os.Stat(cudaDllPath); os.IsNotExist(err) {
			// Copy DLL file to cuda directory
			data, _ := os.ReadFile(dllPath)
			os.WriteFile(cudaDllPath, data, 0644)
			fmt.Println("Copied CUDA DLL to cuda directory")
		}
	}

	// Same for CUDA runtime DLL
	rtDllPath := filepath.Join(execPath, "cudart64_*.dll")
	matches, _ := filepath.Glob(rtDllPath)
	if len(matches) > 0 {
		for _, match := range matches {
			baseName := filepath.Base(match)
			cudaRtPath := filepath.Join(cudaDir, baseName)
			if _, err := os.Stat(cudaRtPath); os.IsNotExist(err) {
				// Copy DLL file to cuda directory
				data, _ := os.ReadFile(match)
				os.WriteFile(cudaRtPath, data, 0644)
				fmt.Println("Copied CUDA runtime DLL to cuda directory")
			}
		}
	}
}

// ensureGPUSupport checks and ensures GPU support is properly configured
func ensureGPUSupport() bool {
	// Check if CUDA files exist
	cudaDir := filepath.Join("cuda")
	if _, err := os.Stat(cudaDir); os.IsNotExist(err) {
		os.MkdirAll(cudaDir, 0755)
	}

	// Check if DLL files exist in working directory or cuda directory
	cudaDllPath := filepath.Join(cudaDir, "cuda_othello.dll")
	mainDllPath := "cuda_othello.dll"

	_, cudaDllExists := os.Stat(cudaDllPath)
	_, mainDllExists := os.Stat(mainDllPath)

	if os.IsNotExist(cudaDllExists) && os.IsNotExist(mainDllExists) {
		fmt.Println("ERROR: cuda_othello.dll not found in either working directory or cuda/ directory")
		return false
	}

	// Check if cudart runtime is available
	cudartFound := false

	filepath.Walk(cudaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), "cudart") {
			cudartFound = true
		}
		return nil
	})

	if !cudartFound {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.Contains(info.Name(), "cudart") {
				cudartFound = true
			}
			return nil
		})
	}

	if !cudartFound {
		fmt.Println("WARNING: CUDA runtime library not found, GPU support may not work correctly")
	}

	// Create GPU evaluation header if needed
	headerPath := filepath.Join(cudaDir, "othello_cuda.h")
	if _, err := os.Stat(headerPath); os.IsNotExist(err) {
		fmt.Println("Creating GPU support header file...")
		createGPUHeader(headerPath)
	}

	return true
}

// createGPUHeader creates the GPU evaluation header file
func createGPUHeader(path string) {
	headerContent := `#ifndef OTHELLO_CUDA_H
#define OTHELLO_CUDA_H

#ifdef __cplusplus
extern "C" {
#endif

// Initialize CUDA and return success status
__declspec(dllexport) int initCUDA();

// Set evaluation coefficients
__declspec(dllexport) void setCoefficients(int* material, int* mobility, int* corners, 
                    int* parity, int* stability, int* frontier);

// Evaluate multiple game states in parallel
__declspec(dllexport) void evaluateStates(int* boards, int* player_colors, int* scores, int num_states);

// Free CUDA resources
__declspec(dllexport) void cleanupCUDA();

#ifdef __cplusplus
}
#endif

#endif // OTHELLO_CUDA_H`

	err := os.WriteFile(path, []byte(headerContent), 0644)
	if err != nil {
		fmt.Printf("Error creating GPU header: %v\n", err)
	}
}
