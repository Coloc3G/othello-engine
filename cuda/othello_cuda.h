#ifndef OTHELLO_CUDA_H
#define OTHELLO_CUDA_H

#ifdef __cplusplus
extern "C"
{
#endif

  // Initialize CUDA and return success status (1=success, 0=failure)
  int initCUDA();

  // Set evaluation coefficients
  void setCoefficients(int *material, int *mobility, int *corners,
                       int *parity, int *stability, int *frontier);

  // Evaluate multiple game states in parallel
  void evaluateStates(int *boards, int *player_colors, int *scores, int num_states);

  // Free CUDA resources
  void cleanupCUDA();

#ifdef __cplusplus
}
#endif

#endif // OTHELLO_CUDA_H
