@echo off
setlocal EnableDelayedExpansion

REM CUDA build script for Othello Engine
echo Building CUDA components...

REM Check if NVCC is available
where nvcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Please ensure CUDA toolkit is installed and in your PATH
    exit /b 1
)

REM Compile CUDA code
echo Compiling CUDA kernels...
if not exist othello_cuda.cu (
    echo Error: othello_cuda.cu not found!
    exit /b 1
)

nvcc -c -o cuda_othello.obj othello_cuda.cu
if %ERRORLEVEL% NEQ 0 (
    echo Compilation failed!
    exit /b 1
)

REM Create DLL
echo Creating DLL...
nvcc --shared -o cuda_othello.dll cuda_othello.obj
if %ERRORLEVEL% NEQ 0 (
    echo DLL creation failed!
    exit /b 1
)

REM Check if build succeeded
if exist cuda_othello.dll (
    echo Build successful: cuda_othello.dll created
    
    REM Copy to parent directory for easy access
    copy cuda_othello.dll ..\ /Y
    echo Copied DLL to parent directory
) else (
    echo Build failed
    exit /b 1
)

echo CUDA build complete!
