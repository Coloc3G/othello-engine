#include <stdbool.h>

#ifndef OTHELLO_CUDA_H
#define OTHELLO_CUDA_H

// Define constants first so they can be used throughout the file
#define BOARD_SIZE 8
#define EMPTY 0
#define WHITE 1
#define BLACK 2

// Define evaluation coefficient structure before it's used
typedef struct
{
  int material_coeff[3];
  int mobility_coeff[3];
  int corners_coeff[3];
  int parity_coeff[3];
  int stability_coeff[3];
  int frontier_coeff[3];
} EvaluationCoefficients;

#ifdef __cplusplus
extern "C"
{
#endif

  // Initialize CUDA and return success status (1=success, 0=failure)
  __declspec(dllexport) int initCUDA();

  // Set evaluation coefficients
  __declspec(dllexport) void setCoefficients(int *material, int *mobility, int *corners,
                                             int *parity, int *stability, int *frontier);
  // Evaluate multiple game states in parallel
  __declspec(dllexport) void evaluateStates(int *boards, int *player_colors, int *scores, int num_states);

  // Evaluate and find best moves for multiple positions in parallel
  __declspec(dllexport) void evaluateAndFindBestMoves(int *boards, int *player_colors, int *depths,
                                                      int *scores, int *best_rows, int *best_cols, int num_states);

  // Perform minimax search to find best move for a board
  __declspec(dllexport) int findBestMove(int *board, int player_color, int depth, int *best_row, int *best_col);

  // Check if a player has valid moves
  __declspec(dllexport) int hasValidMoves(int *board, int player_color);

  // Check if game is finished (no valid moves for either player)
  __declspec(dllexport) int isGameFinished(int *board);

  // Free CUDA resources
  __declspec(dllexport) void cleanupCUDA();

  // Get GPU memory usage statistics
  __declspec(dllexport) void getGPUMemoryInfo(unsigned long long *free_memory, unsigned long long *total_memory);

  // Add debug evaluation function
  __declspec(dllexport) int debugEvaluateBoard(int *board, int player_color, int *debug_info);

  // Host-side helper function prototypes
  bool isValidMoveHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int row, int col);
  void applyMoveHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int row, int col);
  int getValidMovesHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int moves_r[64], int moves_c[64]);
  int evaluateBoardHost(int board[BOARD_SIZE][BOARD_SIZE], int player, EvaluationCoefficients coeffs);
  bool isGameFinishedHost(int board[BOARD_SIZE][BOARD_SIZE]);
  int minimaxHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int depth, bool maximizing,
                  int alpha, int beta, int *best_row, int *best_col, EvaluationCoefficients coeffs);

#ifdef __cplusplus
}
#endif

#endif // OTHELLO_CUDA_H
