#ifndef OTHELLO_CUDA_H
#define OTHELLO_CUDA_H

#ifdef __cplusplus
extern "C"
{
#endif

  // Initialize CUDA and return success status (1=success, 0=failure)
  __declspec(dllexport) int initCUDA();

  // Initialize Zobrist hash table for transposition table
  __declspec(dllexport) void initZobristTable();

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

#ifdef __cplusplus
}
#endif

#endif // OTHELLO_CUDA_H
