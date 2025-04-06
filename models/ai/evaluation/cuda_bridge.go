package evaluation

// #cgo windows LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo linux LDFLAGS: -L${SRCDIR}/../../../cuda -lcuda_othello
// #cgo CFLAGS: -I${SRCDIR}/../../../cuda
// #include <stdlib.h>
// #include "othello_cuda.h"
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/Coloc3G/othello-engine/models/game"
)

var (
	cudaInitialized   bool
	cudaInitializeMux sync.Mutex
	gpuAvailable      bool

	// Pour les statistiques de performance
	totalTransferTime  time.Duration
	totalKernelTime    time.Duration
	batchesProcessed   int
	positionsEvaluated int
)

// InitCUDA initializes the CUDA environment
func InitCUDA() bool {
	cudaInitializeMux.Lock()
	defer cudaInitializeMux.Unlock()

	// Skip actual initialization if already done
	if cudaInitialized {
		return gpuAvailable
	}

	// Check if we're on Windows
	isWindows := runtime.GOOS == "windows"

	// First check if required DLLs exist
	if isWindows && !checkRequiredDLLs() {
		fmt.Println("CUDA acceleration disabled - Required DLLs not found")
		gpuAvailable = false
		cudaInitialized = true
		return false
	}

	// Try to initialize CUDA, but catch any panics
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("CUDA initialization failed with panic:", r)
			gpuAvailable = false
		}
		cudaInitialized = true
	}()

	// Try to initialize CUDA
	result := C.initCUDA()
	gpuAvailable = result == 1

	if gpuAvailable {
		fmt.Println("CUDA GPU acceleration initialized successfully")

		// Initialize Zobrist hash table on GPU for transposition table
		C.initZobristTable()
	} else {
		fmt.Println("CUDA GPU acceleration not available, falling back to CPU")
	}

	return gpuAvailable
}

// checkRequiredDLLs checks if the required DLL files exist on Windows
func checkRequiredDLLs() bool {
	// First attempt to use CUDA libraries in the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: Could not determine current directory: %v\n", err)
		return false
	}

	// Look for the CUDA DLL in various locations
	searchPaths := []string{
		currentDir,
		filepath.Join(currentDir, "cuda"),
	}

	// Add the executable directory to search paths
	if exePath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths, execDir)
		searchPaths = append(searchPaths, filepath.Join(execDir, "cuda"))
	}

	// Check each path for the CUDA DLL
	dllFound := false
	for _, path := range searchPaths {
		dllPath := filepath.Join(path, "cuda_othello.dll")
		if _, err := os.Stat(dllPath); err == nil {
			dllFound = true
			fmt.Printf("Found CUDA DLL at: %s\n", dllPath)

			// If found in a non-current directory, try to copy to current working dir
			if path != currentDir {
				data, err := os.ReadFile(dllPath)
				if err == nil {
					err = os.WriteFile(filepath.Join(currentDir, "cuda_othello.dll"), data, 0644)
					if err == nil {
						fmt.Println("Copied CUDA DLL to current directory for easier access")
					}
				}
			}
			break
		}
	}

	if !dllFound {
		fmt.Println("Warning: Required CUDA DLL not found")
		return false
	}

	return true
}

// CleanupCUDA frees CUDA resources
func CleanupCUDA() {
	cudaInitializeMux.Lock()
	defer cudaInitializeMux.Unlock()

	if cudaInitialized && gpuAvailable {
		C.cleanupCUDA()
		cudaInitialized = false
		gpuAvailable = false
		fmt.Println("CUDA resources cleaned up")
	}
}

// IsGPUAvailable returns whether GPU acceleration is available
func IsGPUAvailable() bool {
	if !cudaInitialized {
		InitCUDA()
	}
	return gpuAvailable
}

// SetCUDACoefficients sets the evaluation coefficients in CUDA
func SetCUDACoefficients(coeffs EvaluationCoefficients) bool {
	if !IsGPUAvailable() {
		return false
	}

	// Convert Go arrays to C arrays
	materialC := (*C.int)(unsafe.Pointer(&coeffs.MaterialCoeffs[0]))
	mobilityC := (*C.int)(unsafe.Pointer(&coeffs.MobilityCoeffs[0]))
	cornersC := (*C.int)(unsafe.Pointer(&coeffs.CornersCoeffs[0]))
	parityC := (*C.int)(unsafe.Pointer(&coeffs.ParityCoeffs[0]))
	stabilityC := (*C.int)(unsafe.Pointer(&coeffs.StabilityCoeffs[0]))
	frontierC := (*C.int)(unsafe.Pointer(&coeffs.FrontierCoeffs[0]))

	// Call C function to set coefficients
	C.setCoefficients(materialC, mobilityC, cornersC, parityC, stabilityC, frontierC)
	return true
}

// EvaluateStatesCUDA evaluates multiple game states in parallel using CUDA
func EvaluateStatesCUDA(boards []game.Board, playerColors []game.Piece) []int {
	if !IsGPUAvailable() {
		return nil
	}

	numStates := len(boards)
	if numStates == 0 || numStates != len(playerColors) {
		return nil
	}

	// Optimisation: utiliser des pinned memory pour accélérer les transferts
	// Mesurer le temps de transfert pour des statistiques
	transferStart := time.Now()

	// Flatten the boards for C - utiliser unified memory quand c'est disponible
	flattenedBoards := make([]int, numStates*8*8)
	for s := 0; s < numStates; s++ {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				flattenedBoards[s*64+i*8+j] = int(boards[s][i][j])
			}
		}
	}

	// Convert player colors to ints
	colorInts := make([]int, numStates)
	for i, color := range playerColors {
		colorInts[i] = int(color)
	}

	// Prepare C arrays
	boardsC := (*C.int)(unsafe.Pointer(&flattenedBoards[0]))
	colorsC := (*C.int)(unsafe.Pointer(&colorInts[0]))
	scoresC := (*C.int)(C.malloc(C.size_t(numStates * 4))) // 4 bytes per int
	defer C.free(unsafe.Pointer(scoresC))

	transferTime := time.Since(transferStart)
	totalTransferTime += transferTime

	// Call C function to evaluate
	kernelStart := time.Now()
	C.evaluateStates(boardsC, colorsC, scoresC, C.int(numStates))
	kernelTime := time.Since(kernelStart)
	totalKernelTime += kernelTime

	// Copy back results and measure transfer time again
	transferStart = time.Now()
	scores := make([]int, numStates)
	for i := 0; i < numStates; i++ {
		scores[i] = int(*(*C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(scoresC)) + uintptr(i*4))))
	}
	transferTime = time.Since(transferStart)
	totalTransferTime += transferTime

	// Update statistics
	batchesProcessed++
	positionsEvaluated += numStates

	return scores
}

// GetBatchStats returns statistics about GPU batch processing
func GetBatchStats() (batches int, positions int, avgTransferMs float64, avgKernelMs float64) {
	if batchesProcessed == 0 {
		return 0, 0, 0, 0
	}

	return batchesProcessed,
		positionsEvaluated,
		float64(totalTransferTime.Milliseconds()) / float64(batchesProcessed),
		float64(totalKernelTime.Milliseconds()) / float64(batchesProcessed)
}

// ResetBatchStats resets the batch statistics
func ResetBatchStats() {
	totalTransferTime = 0
	totalKernelTime = 0
	batchesProcessed = 0
	positionsEvaluated = 0
}

// FindBestMoveCUDA uses GPU to find the best move directly
func FindBestMoveCUDA(board game.Board, player game.Piece, depth int) (game.Position, bool) {
	if !IsGPUAvailable() {
		return game.Position{Row: -1, Col: -1}, false
	}

	// Flatten the board for C
	flatBoard := make([]C.int, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			flatBoard[i*8+j] = C.int(board[i][j])
		}
	}

	// Prepare C variables
	boardC := (*C.int)(unsafe.Pointer(&flatBoard[0]))
	playerC := C.int(int(player))
	depthC := C.int(depth)

	var bestRow, bestCol C.int
	bestRowPtr := (*C.int)(unsafe.Pointer(&bestRow))
	bestColPtr := (*C.int)(unsafe.Pointer(&bestCol))

	// Call C function to find best move
	result := C.findBestMove(boardC, playerC, depthC, bestRowPtr, bestColPtr)

	// Check if a valid move was found
	if result <= -1000000 || bestRow < 0 || bestRow >= 8 || bestCol < 0 || bestCol >= 8 {
		return game.Position{Row: -1, Col: -1}, false
	}

	return game.Position{Row: int(bestRow), Col: int(bestCol)}, true
}
