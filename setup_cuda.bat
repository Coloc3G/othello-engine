@echo off
setlocal

echo Othello Engine CUDA Setup
echo ========================

REM Check if we're in the right directory
if not exist cuda\othello_cuda.cu (
    echo Error: This script must be run from the project root directory
    echo Expected to find cuda\othello_cuda.cu
    exit /b 1
)

echo Step 1: Building CUDA library
cd cuda
call build_cuda_windows.bat
if %ERRORLEVEL% NEQ 0 (
    echo CUDA build failed
    cd ..
    exit /b 1
)

echo Step 2: Verifying DLL
call verify_dll.bat
if %ERRORLEVEL% NEQ 0 (
    echo DLL verification failed
    cd ..
    exit /b 1
)
cd ..

echo Step 3: Ensuring DLL is in appropriate directories
if not exist models\ai\evaluation\cuda_othello.dll (
    mkdir models\ai\evaluation 2>nul
    copy cuda\cuda_othello.dll models\ai\evaluation\
)

if not exist models\ai\learning\cuda_othello.dll (
    mkdir models\ai\learning 2>nul
    copy cuda\cuda_othello.dll models\ai\learning\
)

copy cuda\cuda_othello.dll .\ /Y

echo.
echo Setup complete! You can now build and run the Othello Engine with GPU support.
echo.
echo To test, try running: go run main.go -benchmark
