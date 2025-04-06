#ifndef CUDA_EXPORT_FIX_H
#define CUDA_EXPORT_FIX_H

// This header provides a fix for export issues in mixed C++/C environments

#ifdef _WIN32
#define OTHELLO_EXPORT __declspec(dllexport)
#else
#define OTHELLO_EXPORT
#endif

// These function declarations match the implementation in othello_cuda.cu
// and ensure they're properly exported

#ifdef __cplusplus
extern "C"
{
#endif

  OTHELLO_EXPORT int initCUDA();
  OTHELLO_EXPORT void initZobristTable();
  OTHELLO_EXPORT void setCoefficients(int *material, int *mobility, int *corners,
                                      int *parity, int *stability, int *frontier);
  OTHELLO_EXPORT void evaluateStates(int *boards, int *player_colors, int *scores, int num_states);
  OTHELLO_EXPORT int findBestMove(int *board, int player_color, int depth, int *best_row, int *best_col);
  OTHELLO_EXPORT int hasValidMoves(int *board, int player_color);
  OTHELLO_EXPORT int isGameFinished(int *board);
  OTHELLO_EXPORT void cleanupCUDA();
  OTHELLO_EXPORT void getGPUMemoryInfo(unsigned long long *free_memory, unsigned long long *total_memory);

#ifdef __cplusplus
}
#endif

#endif // CUDA_EXPORT_FIX_H
