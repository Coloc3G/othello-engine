@echo off
setlocal

echo Othello Engine Runner
echo ====================

REM Check if CUDA DLLs exist in root directory
if exist "cuda_othello.dll" (
    echo CUDA DLL found in root directory
) else (
    echo CUDA DLL not found in root directory, checking cuda directory...
    if exist "cuda\cuda_othello.dll" (
        echo CUDA DLL found in cuda directory, copying to root...
        copy "cuda\cuda_othello.dll" .
    ) else (
        echo CUDA DLL not found! GPU acceleration will be disabled.
    )
)

REM Check for CUDA runtime
if exist "cudart64_*.dll" (
    echo CUDA runtime found.
) else (
    echo CUDA runtime not found in root, checking cuda directory...
    if exist "cuda\cudart64_*.dll" (
        echo CUDA runtime found in cuda directory, copying to root...
        copy "cuda\cudart64_*.dll" .
    ) else (
        echo CUDA runtime not found! GPU acceleration will be disabled.
    )
)

REM Set CGO_ENABLED=0 to avoid CUDA issues when not needed
if "%1"=="-mode" (
    if "%2"=="train" (
        echo Training mode detected, using CPU-only for reliability
        set CGO_ENABLED=0
    )
    if "%2"=="compare" (
        echo Comparison mode detected, using CPU-only for reliability
        set CGO_ENABLED=0
    )
)

REM Run the application with the specified arguments
echo.
echo Running Othello Engine with arguments: %*
go run main.go %*

echo.
echo Program execution completed
