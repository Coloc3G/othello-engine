#include <stdio.h>
#include <stdlib.h>
#include <cuda_runtime.h>
#include <time.h>
#include "othello_cuda.h"

// Constant board size
#define BOARD_SIZE 8
#define EMPTY 0
#define WHITE 1
#define BLACK 2

// Increase BLOCK_SIZE for better GPU utilization
#define BLOCK_SIZE 256

// Maximum number of positions to store in transposition table
#define MAX_POSITIONS_POOL 65536
#define POSITIONS_BATCH_SIZE 1024

// Game state structure
typedef struct
{
  int board[BOARD_SIZE][BOARD_SIZE];
  int player_color;
} GameState;

// Global device variables
__constant__ EvaluationCoefficients d_coeffs;

// Host-side copies
EvaluationCoefficients h_coeffs;
cudaError_t cuda_status = cudaSuccess;

// Debug mode flag for more verbose output
#define DEBUG_MODE 0

// Add debug flag for printing evaluation details
#define DEBUG_EVAL 1

// Global coefficient arrays on device
__device__ int d_material_coeffs[3];
__device__ int d_mobility_coeffs[3];
__device__ int d_corners_coeffs[3];
__device__ int d_parity_coeffs[3];
__device__ int d_stability_coeffs[3];
__device__ int d_frontier_coeffs[3];

//-----------------------------------------------------------------------
// Device-only functions (run on GPU)
//-----------------------------------------------------------------------

// Check if a move is valid for the given board and player
__device__ bool isValidMove(int board[BOARD_SIZE][BOARD_SIZE], int player, int row, int col)
{
  // Check if the position is empty
  if (board[row][col] != EMPTY)
    return false;

  // Get opponent color
  int opponent = (player == WHITE) ? BLACK : WHITE;

  // Direction vectors for all 8 directions
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  // Check all 8 directions
  for (int dir = 0; dir < 8; dir++)
  {
    int r = row + dx[dir];
    int c = col + dy[dir];

    // First step must have opponent piece
    if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && board[r][c] == opponent)
    {
      // Continue in this direction
      r += dx[dir];
      c += dy[dir];

      // Keep going until we find an empty cell, edge, or our own piece
      while (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE)
      {
        if (board[r][c] == EMPTY)
          break;
        if (board[r][c] == player)
          return true; // Found our own piece, move is valid

        // Continue in this direction
        r += dx[dir];
        c += dy[dir];
      }
    }
  }

  return false;
}

// Apply a move to the board and return a new board
__device__ void applyMove(int original[BOARD_SIZE][BOARD_SIZE], int result[BOARD_SIZE][BOARD_SIZE],
                          int player, int row, int col)
{
  // Copy the original board
  for (int i = 0; i < BOARD_SIZE; i++)
    for (int j = 0; j < BOARD_SIZE; j++)
      result[i][j] = original[i][j];

  // Place the piece
  result[row][col] = player;

  // Get opponent color
  int opponent = (player == WHITE) ? BLACK : WHITE;

  // Direction vectors for all 8 directions
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  // Check all 8 directions and flip pieces
  for (int dir = 0; dir < 8; dir++)
  {
    int r = row + dx[dir];
    int c = col + dy[dir];

    // Pieces to flip in this direction
    int flip_r[BOARD_SIZE * BOARD_SIZE], flip_c[BOARD_SIZE * BOARD_SIZE];
    int flip_count = 0;

    // Check if first piece is opponent
    if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && result[r][c] == opponent)
    {
      // Remember this piece
      flip_r[flip_count] = r;
      flip_c[flip_count] = c;
      flip_count++;

      // Continue in this direction
      r += dx[dir];
      c += dy[dir];

      // Find all opponent pieces in this direction
      while (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE)
      {
        if (result[r][c] == EMPTY)
        {
          flip_count = 0; // No pieces to flip
          break;
        }

        if (result[r][c] == player)
          break; // Found our piece, can flip

        // Remember opponent piece
        flip_r[flip_count] = r;
        flip_c[flip_count] = c;
        flip_count++;

        // Continue in this direction
        r += dx[dir];
        c += dy[dir];
      }

      // If we found our piece at the end, flip all pieces in between
      if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && result[r][c] == player)
      {
        for (int i = 0; i < flip_count; i++)
          result[flip_r[i]][flip_c[i]] = player;
      }
    }
  }
}

// Get all valid moves for a player
__device__ int getValidMoves(int board[BOARD_SIZE][BOARD_SIZE], int player, int moves_r[64], int moves_c[64])
{
  int count = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (isValidMove(board, player, i, j))
      {
        moves_r[count] = i;
        moves_c[count] = j;
        count++;
      }
    }
  }
  return count;
}

// Evaluate stability of pieces
__device__ int evaluateStability(int board[BOARD_SIZE][BOARD_SIZE], int player, int opponent)
{
  // Pre-computed stability weights
  const int stability_map[BOARD_SIZE][BOARD_SIZE] = {
      {4, -3, 2, 2, 2, 2, -3, 4},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {4, -3, 2, 2, 2, 2, -3, 4}};

  int player_stability = 0;
  int opponent_stability = 0;

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] == player)
      {
        player_stability += stability_map[i][j];
      }
      else if (board[i][j] == opponent)
      {
        opponent_stability += stability_map[i][j];
      }
    }
  }

  return player_stability - opponent_stability;
}

// Count frontier discs (adjacent to empty spaces)
__device__ int countFrontierDiscs(int board[BOARD_SIZE][BOARD_SIZE], int player)
{
  int count = 0;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] == player)
      {
        // Check if this piece is adjacent to an empty square
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];

          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE &&
              board[r][c] == EMPTY)
          {
            count++;
            break; // Count each piece only once
          }
        }
      }
    }
  }

  return count;
}

// Enhanced board evaluation with more heuristics - made completely deterministic
__device__ int evaluateBoard(int *board, int player, int phase)
{
  // Convert the flat board array to 2D array
  int board_2d[BOARD_SIZE][BOARD_SIZE];
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      board_2d[i][j] = board[i * BOARD_SIZE + j];
    }
  }

  int opponent = (player == WHITE) ? BLACK : WHITE;

  // Count pieces for phase
  int piece_count = 0;
  int player_pieces = 0;
  int opponent_pieces = 0;

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] != EMPTY)
      {
        piece_count++;
        if (board_2d[i][j] == player)
          player_pieces++;
        else if (board_2d[i][j] == opponent)
          opponent_pieces++;
      }
    }
  }

  // Get all raw score components and coefficients
  // material
  int material_score = player_pieces - opponent_pieces;

  // corners
  int player_corners = 0;
  int opponent_corners = 0;
  if (board_2d[0][0] == player)
    player_corners++;
  if (board_2d[0][7] == player)
    player_corners++;
  if (board_2d[7][0] == player)
    player_corners++;
  if (board_2d[7][7] == player)
    player_corners++;
  if (board_2d[0][0] == opponent)
    opponent_corners++;
  if (board_2d[0][7] == opponent)
    opponent_corners++;
  if (board_2d[7][0] == opponent)
    opponent_corners++;
  if (board_2d[7][7] == opponent)
    opponent_corners++;
  int corner_score = player_corners - opponent_corners;

  // mobility
  int moves_r[64], moves_c[64];
  int player_moves = getValidMoves(board_2d, player, moves_r, moves_c);
  int opponent_moves = getValidMoves(board_2d, opponent, moves_r, moves_c);
  int mobility_score = player_moves - opponent_moves;

  // parity
  int empty_squares = 64 - player_pieces - opponent_pieces;
  int parity_score = 0;
  if (empty_squares % 2 == 0)
  {
    parity_score = (player == BLACK) ? -1 : 1;
  }
  else
  {
    parity_score = (player == BLACK) ? 1 : -1;
  }

  // stability calculation
  const int stability_map[BOARD_SIZE][BOARD_SIZE] = {
      {4, -3, 2, 2, 2, 2, -3, 4},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {4, -3, 2, 2, 2, 2, -3, 4}};

  int player_stability = 0;
  int opponent_stability = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] == player)
      {
        player_stability += stability_map[i][j];
      }
      else if (board_2d[i][j] == opponent)
      {
        opponent_stability += stability_map[i][j];
      }
    }
  }
  int stability_score = player_stability - opponent_stability;

  // frontier
  int player_frontier = 0;
  int opponent_frontier = 0;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] == player)
      {
        // Check if adjacent to empty
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];
          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE &&
              board_2d[r][c] == EMPTY)
          {
            player_frontier++;
            break;
          }
        }
      }
      else if (board_2d[i][j] == opponent)
      {
        // Check if adjacent to empty
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];
          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE &&
              board_2d[r][c] == EMPTY)
          {
            opponent_frontier++;
            break;
          }
        }
      }
    }
  }
  int frontier_score = opponent_frontier - player_frontier;

  // Calculate weighted components
  int material_contrib = d_coeffs.material_coeff[phase] * material_score;
  int mobility_contrib = d_coeffs.mobility_coeff[phase] * mobility_score;
  int corner_contrib = d_coeffs.corners_coeff[phase] * corner_score;
  int parity_contrib = d_coeffs.parity_coeff[phase] * parity_score;
  int stability_contrib = d_coeffs.stability_coeff[phase] * stability_score;
  int frontier_contrib = d_coeffs.frontier_coeff[phase] * frontier_score;

  // Final score
  int final_score = material_contrib + mobility_contrib + corner_contrib +
                    parity_contrib + stability_contrib + frontier_contrib;

  // Add detailed debug output if enabled
  if (DEBUG_EVAL)
  {
    printf("[GPU] P%d Phase=%d: Mat(%d*%d=%d) Mob(%d*%d=%d) Cor(%d*%d=%d) Par(%d*%d=%d) Stb(%d*%d=%d) Frt(%d*%d=%d) = %d\n",
           player, phase,
           d_coeffs.material_coeff[phase], material_score, material_contrib,
           d_coeffs.mobility_coeff[phase], mobility_score, mobility_contrib,
           d_coeffs.corners_coeff[phase], corner_score, corner_contrib,
           d_coeffs.parity_coeff[phase], parity_score, parity_contrib,
           d_coeffs.stability_coeff[phase], stability_score, stability_contrib,
           d_coeffs.frontier_coeff[phase], frontier_score, frontier_contrib,
           final_score);
  }

  return final_score;
}

// CUDA kernel to evaluate multiple game states in parallel with shared memory
__global__ void evaluateStatesKernel(GameState *states, int *scores, int num_states)
{
  // Use shared memory for faster access
  __shared__ int shared_results[BLOCK_SIZE];

  int idx = blockIdx.x * blockDim.x + threadIdx.x;
  int tid = threadIdx.x;

  if (idx < num_states)
  {
    GameState state = states[idx];
    int board[BOARD_SIZE][BOARD_SIZE];
    int player_color = state.player_color;

    // Copy the board to local memory for faster access
    for (int i = 0; i < BOARD_SIZE; i++)
    {
      for (int j = 0; j < BOARD_SIZE; j++)
      {
        board[i][j] = state.board[i][j];
      }
    }

    // Flatten the 2D board into a 1D array
    int flatBoard[64];
    for (int i = 0; i < BOARD_SIZE; i++)
    {
      for (int j = 0; j < BOARD_SIZE; j++)
      {
        flatBoard[i * BOARD_SIZE + j] = board[i][j];
      }
    }

    // Determine game phase
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

    int phase;
    if (piece_count < 20)
      phase = 0;
    else if (piece_count <= 58)
      phase = 1;
    else
      phase = 2;

    // Calculate and store the evaluation score
    // Note: evaluateBoard returns score with opposite sign to CPU version
    shared_results[tid] = evaluateBoard(flatBoard, player_color, phase);
  }
  else
  {
    // Default value for unused threads
    shared_results[tid] = 0;
  }

  // Synchronize threads in the block
  __syncthreads();

  // Copy result to global memory
  if (idx < num_states)
  {
    scores[idx] = shared_results[tid];
  }
}

//-----------------------------------------------------------------------
// Host-only code (CPU side)
//-----------------------------------------------------------------------

// Host function to check if a move is valid (CPU implementation)
bool isValidMoveHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int row, int col)
{
  if (board[row][col] != EMPTY)
    return false;

  int opponent = (player == WHITE) ? BLACK : WHITE;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int dir = 0; dir < 8; dir++)
  {
    int r = row + dx[dir];
    int c = col + dy[dir];

    if (r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || board[r][c] != opponent)
      continue;

    r += dx[dir];
    c += dy[dir];
    bool foundPlayerPiece = false;

    while (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE)
    {
      if (board[r][c] == EMPTY)
        break;
      if (board[r][c] == player)
      {
        foundPlayerPiece = true;
        break;
      }
      r += dx[dir];
      c += dy[dir];
    }

    if (foundPlayerPiece)
      return true;
  }

  return false;
}

// Host function to apply a move (CPU implementation)
void applyMoveHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int row, int col)
{
  board[row][col] = player;
  int opponent = (player == WHITE) ? BLACK : WHITE;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int dir = 0; dir < 8; dir++)
  {
    int r = row + dx[dir];
    int c = col + dy[dir];

    if (r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || board[r][c] != opponent)
      continue;

    // Store positions to flip
    int flipPositions[BOARD_SIZE * BOARD_SIZE][2];
    int flipCount = 0;

    while (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && board[r][c] == opponent)
    {
      flipPositions[flipCount][0] = r;
      flipPositions[flipCount][1] = c;
      flipCount++;
      r += dx[dir];
      c += dy[dir];
    }

    // If we found our piece at the end, flip all pieces in between
    if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && board[r][c] == player)
    {
      for (int i = 0; i < flipCount; i++)
      {
        board[flipPositions[i][0]][flipPositions[i][1]] = player;
      }
    }
  }
}

// Host function to get all valid moves (CPU implementation)
int getValidMovesHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int moves_r[64], int moves_c[64])
{
  int count = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (isValidMoveHost(board, player, i, j))
      {
        moves_r[count] = i;
        moves_c[count] = j;
        count++;
      }
    }
  }
  return count;
}

// Evaluate a board state (CPU implementation)
int evaluateBoardHost(int board[BOARD_SIZE][BOARD_SIZE], int player, EvaluationCoefficients coeffs)
{
  int opponent = (player == WHITE) ? BLACK : WHITE;

  // Count pieces for phase determination
  int piece_count = 0;
  int player_pieces = 0;
  int opponent_pieces = 0;

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] != EMPTY)
      {
        piece_count++;
        if (board[i][j] == player)
          player_pieces++;
        else if (board[i][j] == opponent)
          opponent_pieces++;
      }
    }
  }

  int phase;
  if (piece_count < 20)
    phase = 0; // Early game
  else if (piece_count <= 58)
    phase = 1; // Mid game
  else
    phase = 2; // Late game

  // Material evaluation
  int material_score = player_pieces - opponent_pieces;

  // Corner evaluation
  int player_corners = 0;
  int opponent_corners = 0;
  if (board[0][0] == player)
    player_corners++;
  if (board[0][7] == player)
    player_corners++;
  if (board[7][0] == player)
    player_corners++;
  if (board[7][7] == player)
    player_corners++;

  if (board[0][0] == opponent)
    opponent_corners++;
  if (board[0][7] == opponent)
    opponent_corners++;
  if (board[7][0] == opponent)
    opponent_corners++;
  if (board[7][7] == opponent)
    opponent_corners++;

  // Simple difference for corner score - match device/Go implementation
  int corner_score = player_corners - opponent_corners;

  // Mobility calculation
  int moves_r[64], moves_c[64];
  int player_moves = getValidMovesHost(board, player, moves_r, moves_c);
  int opponent_moves = getValidMovesHost(board, opponent, moves_r, moves_c);

  // Simple difference for mobility - match device/Go implementation
  int mobility_score = player_moves - opponent_moves;

  // Parity evaluation - exactly match device implementation
  int empty_squares = 64 - player_pieces - opponent_pieces;
  int parity_score = 0;
  if (empty_squares % 2 == 0)
  {
    parity_score = (player == BLACK) ? -1 : 1;
  }
  else
  {
    parity_score = (player == BLACK) ? 1 : -1;
  }

  // Stability evaluation using the same map
  const int stability_map[BOARD_SIZE][BOARD_SIZE] = {
      {4, -3, 2, 2, 2, 2, -3, 4},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {4, -3, 2, 2, 2, 2, -3, 4}};

  int player_stability = 0;
  int opponent_stability = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] == player)
      {
        player_stability += stability_map[i][j];
      }
      else if (board[i][j] == opponent)
      {
        opponent_stability += stability_map[i][j];
      }
    }
  }
  int stability_score = player_stability - opponent_stability;

  // Frontier discs calculation - must match device implementation exactly
  int player_frontier = 0;
  int opponent_frontier = 0;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] == player)
      {
        // Check if this piece is adjacent to an empty square
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];

          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && board[r][c] == EMPTY)
          {
            player_frontier++;
            break; // Count each piece only once
          }
        }
      }
      else if (board[i][j] == opponent)
      {
        // Check if this piece is adjacent to an empty square
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];

          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE && board[r][c] == EMPTY)
          {
            opponent_frontier++;
            break; // Count each piece only once
          }
        }
      }
    }
  }

  // Simple difference for frontier score - match device/Go implementation
  int frontier_score = opponent_frontier - player_frontier;

  // Calculate each component's contribution
  int material_contrib = coeffs.material_coeff[phase] * material_score;
  int mobility_contrib = coeffs.mobility_coeff[phase] * mobility_score;
  int corner_contrib = coeffs.corners_coeff[phase] * corner_score;
  int parity_contrib = coeffs.parity_coeff[phase] * parity_score;
  int stability_contrib = coeffs.stability_coeff[phase] * stability_score;
  int frontier_contrib = coeffs.frontier_coeff[phase] * frontier_score;

  // Final weighted score
  int final_score = material_contrib + mobility_contrib + corner_contrib +
                    parity_contrib + stability_contrib + frontier_contrib;

  // Add detailed debug output matching the GPU version
  if (DEBUG_EVAL)
  {
    printf("[CPU] P%d Phase=%d: Mat(%d*%d=%d) Mob(%d*%d=%d) Cor(%d*%d=%d) Par(%d*%d=%d) Stb(%d*%d=%d) Frt(%d*%d=%d) = %d\n",
           player, phase,
           coeffs.material_coeff[phase], material_score, material_contrib,
           coeffs.mobility_coeff[phase], mobility_score, mobility_contrib,
           coeffs.corners_coeff[phase], corner_score, corner_contrib,
           coeffs.parity_coeff[phase], parity_score, parity_contrib,
           coeffs.stability_coeff[phase], stability_score, stability_contrib,
           coeffs.frontier_coeff[phase], frontier_score, frontier_contrib,
           final_score);
  }

  return final_score;
}

// Recursive minimax search (CPU implementation)
int minimaxHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int depth, bool maximizing,
                int alpha, int beta, int *best_row, int *best_col, EvaluationCoefficients coeffs)
{
  // Leaf node evaluation
  if (depth == 0 || isGameFinishedHost(board))
  {
    return evaluateBoardHost(board, player, coeffs);
  }

  int opponent = (player == WHITE) ? BLACK : WHITE;
  int moves_r[64], moves_c[64];
  int move_count;

  // Determine whose turn it is
  int current_player = maximizing ? player : opponent;

  // Get valid moves for current player
  move_count = getValidMovesHost(board, current_player, moves_r, moves_c);

  // Sort moves to ensure consistent ordering with Go implementation
  for (int i = 0; i < move_count - 1; i++)
  {
    for (int j = i + 1; j < move_count; j++)
    {
      // Sort by row first, then column (same as Go implementation)
      if (moves_r[i] > moves_r[j] || (moves_r[i] == moves_r[j] && moves_c[i] > moves_c[j]))
      {
        // Swap rows
        int temp_r = moves_r[i];
        moves_r[i] = moves_r[j];
        moves_r[j] = temp_r;

        // Swap columns
        int temp_c = moves_c[i];
        moves_c[i] = moves_c[j];
        moves_c[j] = temp_c;
      }
    }
  }

  // No moves - check if game is over or we need to pass
  if (move_count == 0)
  {
    // Check if opponent has moves
    int opp_player = maximizing ? opponent : player;
    int opp_moves_r[64], opp_moves_c[64];
    int opp_move_count = getValidMovesHost(board, opp_player, opp_moves_r, opp_moves_c);

    // Game is over
    if (opp_move_count == 0)
    {
      return evaluateBoardHost(board, player, coeffs);
    }

    // Pass turn to other player - same depth but flip maximizing
    return minimaxHost(board, player, depth - 1, !maximizing, alpha, beta, best_row, best_col, coeffs);
  }

  // Maximizing player (our player)
  if (maximizing)
  {
    int best_score = -1000000;
    int best_move_r = -1;
    int best_move_c = -1;

    for (int i = 0; i < move_count; i++)
    {
      // Copy board
      int new_board[BOARD_SIZE][BOARD_SIZE];
      for (int r = 0; r < BOARD_SIZE; r++)
        for (int c = 0; c < BOARD_SIZE; c++)
          new_board[r][c] = board[r][c];

      // Apply move
      applyMoveHost(new_board, player, moves_r[i], moves_c[i]);

      // Recursive search
      int dummy_r = -1, dummy_c = -1;
      int score = minimaxHost(new_board, player, depth - 1, false, alpha, beta, &dummy_r, &dummy_c, coeffs);

      // Use strict '>' comparison for consistent selection with Go implementation
      if (score > best_score)
      {
        best_score = score;
        best_move_r = moves_r[i];
        best_move_c = moves_c[i];
      }

      // Alpha-beta pruning
      alpha = (alpha > best_score) ? alpha : best_score;
      if (beta <= alpha)
        break;
    }

    *best_row = best_move_r;
    *best_col = best_move_c;
    return best_score;
  }
  // Minimizing player (opponent)
  else
  {
    int best_score = 1000000;
    int best_move_r = -1;
    int best_move_c = -1;

    for (int i = 0; i < move_count; i++)
    {
      // Copy board
      int new_board[BOARD_SIZE][BOARD_SIZE];
      for (int r = 0; r < BOARD_SIZE; r++)
        for (int c = 0; c < BOARD_SIZE; c++)
          new_board[r][c] = board[r][c];

      // Apply move
      applyMoveHost(new_board, opponent, moves_r[i], moves_c[i]);

      // Recursive search
      int dummy_r = -1, dummy_c = -1;
      int score = minimaxHost(new_board, player, depth - 1, true, alpha, beta, &dummy_r, &dummy_c, coeffs);

      // Use strict '<' comparison for consistent selection with Go implementation
      if (score < best_score)
      {
        best_score = score;
        best_move_r = moves_r[i];
        best_move_c = moves_c[i];
      }

      // Alpha-beta pruning
      beta = (beta < best_score) ? beta : best_score;
      if (beta <= alpha)
        break;
    }

    *best_row = best_move_r;
    *best_col = best_move_c;
    return best_score;
  }
}

// Add a helper function to check if game is finished
bool isGameFinishedHost(int board[BOARD_SIZE][BOARD_SIZE])
{
  int black_moves_r[64], black_moves_c[64];
  int white_moves_r[64], white_moves_c[64];

  int black_move_count = getValidMovesHost(board, BLACK, black_moves_r, black_moves_c);
  int white_move_count = getValidMovesHost(board, WHITE, white_moves_r, white_moves_c);

  return black_move_count == 0 && white_move_count == 0;
}

//-----------------------------------------------------------------------
// External C interface (exported functions)
//-----------------------------------------------------------------------

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

// Evaluate multiple game states in parallel
__declspec(dllexport) void evaluateStates(int *boards, int *player_colors, int *scores, int num_states)
{
  // Measure transfer and execution time for profiling
  cudaEvent_t start, stop;
  cudaEventCreate(&start);
  cudaEventCreate(&stop);

  cudaEventRecord(start);

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

  // Allocate device memory and copy data - use pinned memory for faster transfers
  cudaMalloc((void **)&d_states, num_states * sizeof(GameState));
  cudaMalloc((void **)&d_scores, num_states * sizeof(int));

  // Copy data to device
  cudaMemcpy(d_states, h_states, num_states * sizeof(GameState), cudaMemcpyHostToDevice);

  // Calculate grid dimensions
  int threads = BLOCK_SIZE;
  int blocks = (num_states + threads - 1) / threads;

  // Launch kernel
  evaluateStatesKernel<<<blocks, threads>>>(d_states, d_scores, num_states);

  // Check for kernel launch errors
  cudaError_t err = cudaGetLastError();
  if (err != cudaSuccess)
  {
    printf("CUDA Kernel Error: %s\n", cudaGetErrorString(err));
    // In case of error, zero all scores
    memset(scores, 0, num_states * sizeof(int));

    // Free memory and return
    cudaFree(d_states);
    cudaFree(d_scores);
    free(h_states);
    return;
  }

  // Copy results back to host
  cudaMemcpy(scores, d_scores, num_states * sizeof(int), cudaMemcpyDeviceToHost);

  // Remove extra sign flip (previously, we had: for (int i = 0; i < num_states; i++) { scores[i] = -scores[i]; })

  cudaEventRecord(stop);
  cudaEventSynchronize(stop);

  float milliseconds = 0;
  cudaEventElapsedTime(&milliseconds, start, stop);

  if (num_states > 1000)
  {
    printf("GPU processed %d states in %.2f ms (%.2f states/ms)\n",
           num_states, milliseconds, num_states / milliseconds);
  }

  // Free memory
  cudaFree(d_states);
  cudaFree(d_scores);
  free(h_states);

  cudaEventDestroy(start);
  cudaEventDestroy(stop);
}

// Find the best move using minimax
__declspec(dllexport) int findBestMove(int *board, int player_color, int depth, int *best_row, int *best_col)
{
  // Convert the flat board array to 2D array
  int board_2d[BOARD_SIZE][BOARD_SIZE];
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      board_2d[i][j] = board[i * BOARD_SIZE + j];
    }
  }

  // Get current evaluation coefficients from global memory
  EvaluationCoefficients h_coeffs;
  cudaMemcpyFromSymbol(&h_coeffs, d_coeffs, sizeof(EvaluationCoefficients));

  // Initialize best move coordinates
  int br = -1, bc = -1;

  // Get valid moves and check for special cases first
  int moves_r[64], moves_c[64];
  int move_count = getValidMovesHost(board_2d, player_color, moves_r, moves_c);

  // No valid moves
  if (move_count == 0)
  {
    *best_row = -1;
    *best_col = -1;
    return -1000000;
  }

  // If only one move, return it immediately
  if (move_count == 1)
  {
    *best_row = moves_r[0];
    *best_col = moves_c[0];

    // Apply the move to get a more accurate score
    int new_board[BOARD_SIZE][BOARD_SIZE];
    for (int i = 0; i < BOARD_SIZE; i++)
      for (int j = 0; j < BOARD_SIZE; j++)
        new_board[i][j] = board_2d[i][j];

    applyMoveHost(new_board, player_color, moves_r[0], moves_c[0]);

    // Get score from the position after our move
    return evaluateBoardHost(new_board, player_color, h_coeffs);
  }

  // Sort the moves for deterministic processing
  for (int i = 0; i < move_count - 1; i++)
  {
    for (int j = i + 1; j < move_count; j++)
    {
      if (moves_r[i] > moves_r[j] || (moves_r[i] == moves_r[j] && moves_c[i] > moves_c[j]))
      {
        // Swap rows
        int temp_r = moves_r[i];
        moves_r[i] = moves_r[j];
        moves_r[j] = temp_r;

        // Swap columns
        int temp_c = moves_c[i];
        moves_c[i] = moves_c[j];
        moves_c[j] = temp_c;
      }
    }
  }

  // Perform the minimax search with the sorted moves
  int best_score = minimaxHost(
      board_2d,     // game board
      player_color, // current player
      depth,        // search depth
      true,         // maximizing player
      -1000000,     // alpha
      1000000,      // beta
      &br,          // best row (output)
      &bc,          // best column (output)
      h_coeffs      // evaluation coefficients
  );

  // Set output parameters
  *best_row = br;
  *best_col = bc;

  // Return consistent score (no sign flip needed here since minimaxHost already uses CPU evaluation)
  return best_score;
}

// Check if a player has valid moves
__declspec(dllexport) int hasValidMoves(int *board, int player_color)
{
  // Convert the flat board array to 2D array
  int board_2d[BOARD_SIZE][BOARD_SIZE];
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      board_2d[i][j] = board[i * BOARD_SIZE + j];
    }
  }

  // Check for any valid move
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (isValidMoveHost(board_2d, player_color, i, j))
      {
        return 1; // At least one valid move exists
      }
    }
  }

  return 0; // No valid moves
}

// Check if game is finished (no valid moves for either player)
__declspec(dllexport) int isGameFinished(int *board)
{
  // No player can move = game is finished
  return !hasValidMoves(board, BLACK) && !hasValidMoves(board, WHITE);
}

// Get GPU memory information
__declspec(dllexport) void getGPUMemoryInfo(unsigned long long *free_memory, unsigned long long *total_memory)
{
  size_t free, total;
  cudaMemGetInfo(&free, &total);

  *free_memory = free;
  *total_memory = total;
}

// Free CUDA resources
__declspec(dllexport) void cleanupCUDA()
{
  cudaDeviceReset();
}

// Evaluate and find best moves for multiple positions in parallel
__declspec(dllexport) void evaluateAndFindBestMoves(int *boards, int *player_colors, int *depths,
                                                    int *scores, int *best_rows, int *best_cols, int num_states)
{
  // Measure transfer and execution time for profiling
  cudaEvent_t start, stop;
  cudaEventCreate(&start);
  cudaEventCreate(&stop);

  cudaEventRecord(start);

  // Process each position sequentially
  // In a more optimized implementation, this would be done in parallel on the GPU
  for (int i = 0; i < num_states; i++)
  {
    // Extract the current board
    int *current_board = &boards[i * 64];
    int player_color = player_colors[i];
    int depth = depths[i];

    // Find best move for this position
    int row = -1, col = -1;
    int score = findBestMove(current_board, player_color, depth, &row, &col);

    // Store results
    scores[i] = score;
    best_rows[i] = row;
    best_cols[i] = col;
  }

  cudaEventRecord(stop);
  cudaEventSynchronize(stop);

  float milliseconds = 0;
  cudaEventElapsedTime(&milliseconds, start, stop);

  if (num_states > 10)
  {
    printf("GPU processed %d minimax positions in %.2f ms (%.2f positions/ms)\n",
           num_states, milliseconds, num_states / milliseconds);
  }

  cudaEventDestroy(start);
  cudaEventDestroy(stop);
}

// Set the evaluation coefficients on the GPU
extern "C" void setCoefficients(int *material, int *mobility, int *corners, int *parity, int *stability, int *frontier)
{
  cudaError_t error;

  // Also set host-side coefficients for deterministic CPU evaluation
  memcpy(h_coeffs.material_coeff, material, 3 * sizeof(int));
  memcpy(h_coeffs.mobility_coeff, mobility, 3 * sizeof(int));
  memcpy(h_coeffs.corners_coeff, corners, 3 * sizeof(int));
  memcpy(h_coeffs.parity_coeff, parity, 3 * sizeof(int));
  memcpy(h_coeffs.stability_coeff, stability, 3 * sizeof(int));
  memcpy(h_coeffs.frontier_coeff, frontier, 3 * sizeof(int));

  // Use cudaMemcpyToSymbol with explicit size and offset
  error = cudaMemcpyToSymbol(d_material_coeffs, material, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting material coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  error = cudaMemcpyToSymbol(d_mobility_coeffs, mobility, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting mobility coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  error = cudaMemcpyToSymbol(d_corners_coeffs, corners, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting corners coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  error = cudaMemcpyToSymbol(d_parity_coeffs, parity, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting parity coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  error = cudaMemcpyToSymbol(d_stability_coeffs, stability, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting stability coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  error = cudaMemcpyToSymbol(d_frontier_coeffs, frontier, 3 * sizeof(int), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting frontier coefficients: %s\n", cudaGetErrorString(error));
    return;
  }

  // Ensure consistency by copying to coeffs structure for both CPU and GPU
  error = cudaMemcpyToSymbol(d_coeffs, &h_coeffs, sizeof(EvaluationCoefficients), 0, cudaMemcpyHostToDevice);
  if (error != cudaSuccess)
  {
    printf("Error setting full coefficients structure: %s\n", cudaGetErrorString(error));
    return;
  }

  // Force synchronization to ensure all memory operations complete
  error = cudaDeviceSynchronize();
  if (error != cudaSuccess)
  {
    printf("Error in synchronization: %s\n", cudaGetErrorString(error));
  }
}

// Add a debug function to export
__declspec(dllexport) int debugEvaluateBoard(int *board, int player_color, int *debug_info)
{
  // Convert the flat board array to 2D array
  int board_2d[BOARD_SIZE][BOARD_SIZE];
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      board_2d[i][j] = board[i * BOARD_SIZE + j];
    }
  }

  int opponent = (player_color == WHITE) ? BLACK : WHITE;

  // Count pieces for phase
  int piece_count = 0;
  int player_pieces = 0;
  int opponent_pieces = 0;

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] != EMPTY)
      {
        piece_count++;
        if (board_2d[i][j] == player_color)
          player_pieces++;
        else if (board_2d[i][j] == opponent)
          opponent_pieces++;
      }
    }
  }

  int phase = piece_count < 20 ? 0 : (piece_count <= 58 ? 1 : 2);

  // Get current evaluation coefficients from global memory
  EvaluationCoefficients h_coeffs;
  cudaMemcpyFromSymbol(&h_coeffs, d_coeffs, sizeof(EvaluationCoefficients));

  // Get all raw score components and coefficients
  // material
  int material_score = player_pieces - opponent_pieces;

  // corners
  int player_corners = 0;
  int opponent_corners = 0;
  if (board_2d[0][0] == player_color)
    player_corners++;
  if (board_2d[0][7] == player_color)
    player_corners++;
  if (board_2d[7][0] == player_color)
    player_corners++;
  if (board_2d[7][7] == player_color)
    player_corners++;
  if (board_2d[0][0] == opponent)
    opponent_corners++;
  if (board_2d[0][7] == opponent)
    opponent_corners++;
  if (board_2d[7][0] == opponent)
    opponent_corners++;
  if (board_2d[7][7] == opponent)
    opponent_corners++;
  int corner_score = player_corners - opponent_corners;

  // mobility
  int moves_r[64], moves_c[64];
  int player_moves = getValidMovesHost(board_2d, player_color, moves_r, moves_c);
  int opponent_moves = getValidMovesHost(board_2d, opponent, moves_r, moves_c);
  int mobility_score = player_moves - opponent_moves;

  // parity
  int empty_squares = 64 - player_pieces - opponent_pieces;
  int parity_score = 0;
  if (empty_squares % 2 == 0)
  {
    parity_score = (player_color == BLACK) ? -1 : 1;
  }
  else
  {
    parity_score = (player_color == BLACK) ? 1 : -1;
  }

  // stability calculation
  const int stability_map[BOARD_SIZE][BOARD_SIZE] = {
      {4, -3, 2, 2, 2, 2, -3, 4},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 0, 1, 1, 0, -1, 2},
      {2, -1, 1, 0, 0, 1, -1, 2},
      {-3, -4, -1, -1, -1, -1, -4, -3},
      {4, -3, 2, 2, 2, 2, -3, 4}};

  int player_stability = 0;
  int opponent_stability = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] == player_color)
      {
        player_stability += stability_map[i][j];
      }
      else if (board_2d[i][j] == opponent)
      {
        opponent_stability += stability_map[i][j];
      }
    }
  }
  int stability_score = player_stability - opponent_stability;

  // frontier
  int player_frontier = 0;
  int opponent_frontier = 0;
  int dx[8] = {-1, -1, -1, 0, 0, 1, 1, 1};
  int dy[8] = {-1, 0, 1, -1, 1, -1, 0, 1};

  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board_2d[i][j] == player_color)
      {
        // Check if adjacent to empty
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];
          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE &&
              board_2d[r][c] == EMPTY)
          {
            player_frontier++;
            break;
          }
        }
      }
      else if (board_2d[i][j] == opponent)
      {
        // Check if adjacent to empty
        for (int dir = 0; dir < 8; dir++)
        {
          int r = i + dx[dir];
          int c = j + dy[dir];
          if (r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE &&
              board_2d[r][c] == EMPTY)
          {
            opponent_frontier++;
            break;
          }
        }
      }
    }
  }
  int frontier_score = opponent_frontier - player_frontier;

  // Calculate weighted components
  int material_contrib = h_coeffs.material_coeff[phase] * material_score;
  int mobility_contrib = h_coeffs.mobility_coeff[phase] * mobility_score;
  int corner_contrib = h_coeffs.corners_coeff[phase] * corner_score;
  int parity_contrib = h_coeffs.parity_coeff[phase] * parity_score;
  int stability_contrib = h_coeffs.stability_coeff[phase] * stability_score;
  int frontier_contrib = h_coeffs.frontier_coeff[phase] * frontier_score;

  // Final score
  int final_score = material_contrib + mobility_contrib + corner_contrib +
                    parity_contrib + stability_contrib + frontier_contrib;

  // Show full breakdown if debug array provided
  if (debug_info != NULL)
  {
    // Store all raw values and weighted values into debug_info array
    debug_info[0] = phase;
    debug_info[1] = material_score;
    debug_info[2] = h_coeffs.material_coeff[phase];
    debug_info[3] = mobility_score;
    debug_info[4] = h_coeffs.mobility_coeff[phase];
    debug_info[5] = corner_score;
    debug_info[6] = h_coeffs.corners_coeff[phase];
    debug_info[7] = parity_score;
    debug_info[8] = h_coeffs.parity_coeff[phase];
    debug_info[9] = stability_score;
    debug_info[10] = h_coeffs.stability_coeff[phase];
    debug_info[11] = frontier_score;
    debug_info[12] = h_coeffs.frontier_coeff[phase];
    debug_info[13] = material_contrib;
    debug_info[14] = mobility_contrib;
    debug_info[15] = corner_contrib;
    debug_info[16] = parity_contrib;
    debug_info[17] = stability_contrib;
    debug_info[18] = frontier_contrib;
    debug_info[19] = final_score;

    // Print full details
    printf("[DEBUG-HOST] P%d Phase=%d: Mat(%d*%d=%d) Mob(%d*%d=%d) Cor(%d*%d=%d) Par(%d*%d=%d) Stb(%d*%d=%d) Frt(%d*%d=%d) => %d\n",
           player_color, phase,
           h_coeffs.material_coeff[phase], material_score, material_contrib,
           h_coeffs.mobility_coeff[phase], mobility_score, mobility_contrib,
           h_coeffs.corners_coeff[phase], corner_score, corner_contrib,
           h_coeffs.parity_coeff[phase], parity_score, parity_contrib,
           h_coeffs.stability_coeff[phase], stability_score, stability_contrib,
           h_coeffs.frontier_coeff[phase], frontier_score, frontier_contrib,
           final_score);
  }

  return final_score;
}
