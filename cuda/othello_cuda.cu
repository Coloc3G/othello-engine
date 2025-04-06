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
#define MAX_ZOBRIST_ENTRIES 1048576

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

// Transposition table entry
typedef struct
{
  unsigned long long key;
  int score;
  int depth;
  int best_move_row;
  int best_move_col;
} TranspositionEntry;

// Position cache for batching
typedef struct
{
  GameState states[MAX_POSITIONS_POOL];
  int scores[MAX_POSITIONS_POOL];
  unsigned long long hashes[MAX_POSITIONS_POOL];
  int cache_size;
} PositionCache;

// Global device variables
__constant__ EvaluationCoefficients d_coeffs;
__device__ unsigned long long d_zobrist_table[3][BOARD_SIZE][BOARD_SIZE];
__device__ TranspositionEntry d_tt[MAX_ZOBRIST_ENTRIES];
__device__ int d_tt_size = 0;
__device__ PositionCache d_position_cache;

// Host-side copies
EvaluationCoefficients h_coeffs;
unsigned long long h_zobrist_table[3][BOARD_SIZE][BOARD_SIZE];
cudaError_t cuda_status = cudaSuccess;

//-----------------------------------------------------------------------
// Zobrist hashing utilities
//-----------------------------------------------------------------------

// Compute Zobrist hash for a board position (device)
__device__ unsigned long long computeZobristHash(int board[BOARD_SIZE][BOARD_SIZE], int player)
{
  unsigned long long hash = player; // Include player in hash

  for (int row = 0; row < BOARD_SIZE; row++)
  {
    for (int col = 0; col < BOARD_SIZE; col++)
    {
      if (board[row][col] != EMPTY)
      {
        hash ^= d_zobrist_table[board[row][col] - 1][row][col];
      }
    }
  }

  return hash;
}

// Compute Zobrist hash for a board position (host)
unsigned long long computeZobristHashHost(int board[BOARD_SIZE][BOARD_SIZE], int player)
{
  unsigned long long hash = player; // Include player in hash

  for (int row = 0; row < BOARD_SIZE; row++)
  {
    for (int col = 0; col < BOARD_SIZE; col++)
    {
      if (board[row][col] != EMPTY)
      {
        hash ^= h_zobrist_table[board[row][col] - 1][row][col];
      }
    }
  }

  return hash;
}

//-----------------------------------------------------------------------
// Transposition table utilities
//-----------------------------------------------------------------------

// Store entry in transposition table (device)
__device__ void storeTranspositionEntry(unsigned long long key, int score,
                                        int depth, int best_move_row, int best_move_col)
{
  // Use key as index with modulo to handle collisions
  int index = key % MAX_ZOBRIST_ENTRIES;

  // Always replace for now (could implement more sophisticated replacement policy)
  d_tt[index].key = key;
  d_tt[index].score = score;
  d_tt[index].depth = depth;
  d_tt[index].best_move_row = best_move_row;
  d_tt[index].best_move_col = best_move_col;

  // Atomic increment to track table size
  atomicMin(&d_tt_size, MAX_ZOBRIST_ENTRIES);
}

// Lookup entry in transposition table (device)
__device__ bool lookupTranspositionEntry(unsigned long long key, int depth,
                                         int *score, int *best_move_row, int *best_move_col)
{
  int index = key % MAX_ZOBRIST_ENTRIES;

  // Check if we have a valid entry with sufficient depth
  if (d_tt[index].key == key && d_tt[index].depth >= depth)
  {
    *score = d_tt[index].score;
    *best_move_row = d_tt[index].best_move_row;
    *best_move_col = d_tt[index].best_move_col;
    return true;
  }

  return false;
}

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

// Enhanced board evaluation with more heuristics
__device__ int evaluateBoard(int board[BOARD_SIZE][BOARD_SIZE], int player, int phase)
{
  int opponent = (player == WHITE) ? BLACK : WHITE;

  // Material evaluation
  int player_pieces = 0;
  int opponent_pieces = 0;
  for (int i = 0; i < BOARD_SIZE; i++)
  {
    for (int j = 0; j < BOARD_SIZE; j++)
    {
      if (board[i][j] == player)
        player_pieces++;
      else if (board[i][j] == opponent)
        opponent_pieces++;
    }
  }
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

  int corner_score = 100 * (player_corners - opponent_corners) / (player_corners + opponent_corners + 1);

  // Mobility evaluation
  int moves_r[64], moves_c[64];
  int player_moves = getValidMoves(board, player, moves_r, moves_c);
  int opponent_moves = getValidMoves(board, opponent, moves_r, moves_c);
  int mobility_score = 100 * (player_moves - opponent_moves) / (player_moves + opponent_moves + 1);

  // Parity evaluation (beneficial to have the last move)
  int empty_squares = 64 - player_pieces - opponent_pieces;
  int parity_score = (empty_squares % 2 == 0) ? -1 : 1; // Even is bad for next player

  // Stability evaluation
  int stability_score = evaluateStability(board, player, opponent);

  // Frontier evaluation
  int player_frontier = countFrontierDiscs(board, player);
  int opponent_frontier = countFrontierDiscs(board, opponent);
  int frontier_score = -100 * (player_frontier - opponent_frontier) / (player_frontier + opponent_frontier + 1);

  // Final score using coefficients
  return d_coeffs.material_coeff[phase] * material_score +
         d_coeffs.mobility_coeff[phase] * mobility_score +
         d_coeffs.corners_coeff[phase] * corner_score +
         d_coeffs.parity_coeff[phase] * parity_score +
         d_coeffs.stability_coeff[phase] * stability_score +
         d_coeffs.frontier_coeff[phase] * frontier_score;
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
    shared_results[tid] = evaluateBoard(board, player_color, phase);
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

  int corner_score = 100 * (player_corners - opponent_corners) / (player_corners + opponent_corners + 1);

  // Mobility calculation
  int moves_r[64], moves_c[64];
  int player_moves = getValidMovesHost(board, player, moves_r, moves_c);
  int opponent_moves = getValidMovesHost(board, opponent, moves_r, moves_c);
  int mobility_score = 100 * (player_moves - opponent_moves) / (player_moves + opponent_moves + 1);

  return coeffs.material_coeff[phase] * material_score +
         coeffs.corners_coeff[phase] * corner_score +
         coeffs.mobility_coeff[phase] * mobility_score;
}

// Recursive minimax search (CPU implementation)
int minimaxHost(int board[BOARD_SIZE][BOARD_SIZE], int player, int depth, bool maximizing,
                int alpha, int beta, int *best_row, int *best_col, EvaluationCoefficients coeffs)
{
  // Leaf node evaluation
  if (depth == 0)
  {
    return evaluateBoardHost(board, maximizing ? player : (player == WHITE ? BLACK : WHITE), coeffs);
  }

  int opponent = (player == WHITE) ? BLACK : WHITE;
  int moves_r[64], moves_c[64];
  int move_count = getValidMovesHost(board, player, moves_r, moves_c);

  // No moves - check if game is over or we need to pass
  if (move_count == 0)
  {
    int opp_moves_r[64], opp_moves_c[64];
    int opp_move_count = getValidMovesHost(board, opponent, opp_moves_r, opp_moves_c);

    // Game is over
    if (opp_move_count == 0)
    {
      return evaluateBoardHost(board, maximizing ? player : opponent, coeffs);
    }

    // Pass turn
    return minimaxHost(board, opponent, depth - 1, !maximizing, alpha, beta, best_row, best_col, coeffs);
  }

  // Maximizing player
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
      int score = minimaxHost(new_board, opponent, depth - 1, false, alpha, beta, &dummy_r, &dummy_c, coeffs);

      if (score > best_score)
      {
        best_score = score;
        best_move_r = moves_r[i];
        best_move_c = moves_c[i];
      }

      // Alpha-beta pruning
      alpha = alpha > best_score ? alpha : best_score;
      if (beta <= alpha)
        break;
    }

    *best_row = best_move_r;
    *best_col = best_move_c;
    return best_score;
  }
  // Minimizing player
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
      applyMoveHost(new_board, player, moves_r[i], moves_c[i]);

      // Recursive search
      int dummy_r = -1, dummy_c = -1;
      int score = minimaxHost(new_board, opponent, depth - 1, true, alpha, beta, &dummy_r, &dummy_c, coeffs);

      if (score < best_score)
      {
        best_score = score;
        best_move_r = moves_r[i];
        best_move_c = moves_c[i];
      }

      // Alpha-beta pruning
      beta = beta < best_score ? beta : best_score;
      if (beta <= alpha)
        break;
    }

    *best_row = best_move_r;
    *best_col = best_move_c;
    return best_score;
  }
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

// Initialize Zobrist hash table
__declspec(dllexport) void initZobristTable()
{
  srand((unsigned int)time(NULL));
  for (int piece = 0; piece < 3; piece++)
  {
    for (int row = 0; row < BOARD_SIZE; row++)
    {
      for (int col = 0; col < BOARD_SIZE; col++)
      {
        // Generate random 64-bit value for each board position and piece
        h_zobrist_table[piece][row][col] =
            ((unsigned long long)rand() << 32) | rand();
      }
    }
  }

  // Copy Zobrist table to device
  cudaMemcpyToSymbol(d_zobrist_table, h_zobrist_table,
                     sizeof(h_zobrist_table));
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

  // Get current evaluation coefficients from device
  EvaluationCoefficients h_coeffs;
  cudaMemcpyFromSymbol(&h_coeffs, d_coeffs, sizeof(EvaluationCoefficients));

  // Initialize best move coordinates
  int br = -1, bc = -1;

  // Perform minimax search
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
