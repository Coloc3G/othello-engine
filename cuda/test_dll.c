#include <stdio.h>
#include <stdlib.h>
#include <windows.h>

typedef void (*InitZobristTableFunc)();
typedef int (*InitCUDAFunc)();

int main()
{
  HMODULE hDll = LoadLibrary("cuda_othello.dll");

  if (hDll == NULL)
  {
    printf("Error: Could not load DLL. Error code: %lu\n", GetLastError());
    return 1;
  }

  printf("DLL loaded successfully.\n");

  // Try to get the address of the initZobristTable function
  InitZobristTableFunc initZobristTable = (InitZobristTableFunc)GetProcAddress(hDll, "initZobristTable");
  if (initZobristTable == NULL)
  {
    printf("Error: Could not find initZobristTable function. Error code: %lu\n", GetLastError());
  }
  else
  {
    printf("Found initZobristTable function.\n");
    // Call the function
    initZobristTable();
    printf("Called initZobristTable successfully.\n");
  }

  // Try another function for comparison
  InitCUDAFunc initCUDA = (InitCUDAFunc)GetProcAddress(hDll, "initCUDA");
  if (initCUDA == NULL)
  {
    printf("Error: Could not find initCUDA function. Error code: %lu\n", GetLastError());
  }
  else
  {
    printf("Found initCUDA function.\n");
  }

  FreeLibrary(hDll);
  return 0;
}
