#include <stdio.h>
#include <stdlib.h>
#include <cuda_runtime.h>

// Constant board size
#define BOARD_SIZE 8
#define EMPTY 0
#define WHITE 1
#define BLACK 2

// Number of threads per block (can be tuned)
#define BLOCK_SIZE 256

// Coefficient structure for evaluation
typedef struct
{
  int material_coeff[3];
  int mobility_coeff[3];
  int corners_coeff[3];
  int parity_coeff[3];
  int stability_coeff[3];
  int frontier_coeff[3];
} EvaluationCoefficients;

// Game state structure
typedef struct
{
  int board[BOARD_SIZE][BOARD_SIZE];
  int player_color;
} GameState;

// Copy coefficients to device constant memory
__constant__ EvaluationCoefficients d_coeffs;

// CUDA kernel to evaluate multiple game states in parallel
__global__ void evaluateStatesKernel(GameState *states, int *scores, int num_states)
{
  int idx = blockIdx.x * blockDim.x + threadIdx.x;

  if (idx < num_states)
  {
    GameState state = states[idx];
    int board[BOARD_SIZE][BOARD_SIZE];
    int player_color = state.player_color;
    int opponent_color = (player_color == WHITE) ? BLACK : WHITE;

    // Copy board to local memory for faster access
    for (int i = 0; i < BOARD_SIZE; i++)
    {
      for (int j = 0; j < BOARD_SIZE; j++)
      {
        board[i][j] = state.board[i][j];
      }
    }

    // Count pieces for phase determination
    int piece_count = 0;
    for (int i = 0; i < BOARD_SIZE; i++)
    {
      for (int j = 0; j < BOARD_SIZE; j++)
      {
        if (board[i][j] != EMPTY)
        {
          piece_count++;
        }
      }
    }

    // Determine game phase (0=early, 1=mid, 2=late)
    int phase;
    if (piece_count < 20)
    {
      phase = 0;
    }
    else if (piece_count <= 58)
    {
      phase = 1;
    }
    else
    {
      phase = 2;
    }

    // Material evaluation
    int player_pieces = 0;
    int opponent_pieces = 0;
    for (int i = 0; i < BOARD_SIZE; i++)
    {
      for (int j = 0; j < BOARD_SIZE; j++)
      {
        if (board[i][j] == player_color)
        {
          player_pieces++;
        }
        else if (board[i][j] == opponent_color)
        {
          opponent_pieces++;
        }
      }
    }
    int material_score = player_pieces - opponent_pieces;

    // Corner evaluation
    int player_corners = 0;
    int opponent_corners = 0;
    if (board[0][0] == player_color)
      player_corners++;
    if (board[0][7] == player_color)
      player_corners++;
    if (board[7][0] == player_color)
      player_corners++;
    if (board[7][7] == player_color)
      player_corners++;

    if (board[0][0] == opponent_color)
      opponent_corners++;
    if (board[0][7] == opponent_color)
      opponent_corners++;
    if (board[7][0] == opponent_color)
      opponent_corners++;
    if (board[7][7] == opponent_color)
      opponent_corners++;

    int corner_score;
    if (player_corners + opponent_corners == 0)
    {
      corner_score = 0;
    }
    else
    {
      corner_score = 100 * (player_corners - opponent_corners) / (player_corners + opponent_corners);
    }

    // Simplified evaluation - using only material and corner for GPU implementation
    int final_score = d_coeffs.material_coeff[phase] * material_score +
                      d_coeffs.corners_coeff[phase] * corner_score;

    scores[idx] = final_score;
  }
}

// C wrapper functions for Go
extern "C"
{

  // Initialize CUDA and return success status
  __declspec(dllexport) int initCUDA()
  {
    cudaError_t error;
    int deviceCount;

    error = cudaGetDeviceCount(&deviceCount);
    if (error != cudaSuccess)
    {
      printf("CUDA Error: %s\n", cudaGetErrorString(error));
      return 0;
    }

    if (deviceCount == 0)
    {
      printf("No CUDA-capable devices found\n");
      return 0;
    }

    // Choose device 0 by default
    error = cudaSetDevice(0);
    if (error != cudaSuccess)
    {
      printf("CUDA Error: %s\n", cudaGetErrorString(error));
      return 0;
    }

    return 1;
  }

  // Set evaluation coefficients
  __declspec(dllexport) void setCoefficients(int *material, int *mobility, int *corners,
                                             int *parity, int *stability, int *frontier)
  {
    EvaluationCoefficients h_coeffs;

    // Copy coefficients from host arrays to host struct
    for (int i = 0; i < 3; i++)
    {
      h_coeffs.material_coeff[i] = material[i];
      h_coeffs.mobility_coeff[i] = mobility[i];
      h_coeffs.corners_coeff[i] = corners[i];
      h_coeffs.parity_coeff[i] = parity[i];
      h_coeffs.stability_coeff[i] = stability[i];
      h_coeffs.frontier_coeff[i] = frontier[i];
    }

    // Copy coefficients to device constant memory
    cudaMemcpyToSymbol(d_coeffs, &h_coeffs, sizeof(EvaluationCoefficients));
  }

  // Evaluate multiple game states in parallel
  __declspec(dllexport) void evaluateStates(int *boards, int *player_colors, int *scores, int num_states)
  {
    GameState *h_states = (GameState *)malloc(num_states * sizeof(GameState));
    GameState *d_states;
    int *d_scores;

    // Prepare game states
    for (int s = 0; s < num_states; s++)
    {
      h_states[s].player_color = player_colors[s];
      for (int i = 0; i < BOARD_SIZE; i++)
      {
        for (int j = 0; j < BOARD_SIZE; j++)
        {
          h_states[s].board[i][j] = boards[s * BOARD_SIZE * BOARD_SIZE + i * BOARD_SIZE + j];
        }
      }
    }

    // Allocate device memory
    cudaMalloc((void **)&d_states, num_states * sizeof(GameState));
    cudaMalloc((void **)&d_scores, num_states * sizeof(int));

    // Copy data to device
    cudaMemcpy(d_states, h_states, num_states * sizeof(GameState), cudaMemcpyHostToDevice);

    // Calculate grid dimensions
    int blocks = (num_states + BLOCK_SIZE - 1) / BLOCK_SIZE;

    // Launch kernel
    evaluateStatesKernel<<<blocks, BLOCK_SIZE>>>(d_states, d_scores, num_states);

    // Copy results back to host
    cudaMemcpy(scores, d_scores, num_states * sizeof(int), cudaMemcpyDeviceToHost);

    // Free memory
    cudaFree(d_states);
    cudaFree(d_scores);
    free(h_states);
  }

  // Free CUDA resources
  __declspec(dllexport) void cleanupCUDA()
  {
    cudaDeviceReset();
  }

} // extern "C"
