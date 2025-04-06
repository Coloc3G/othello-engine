@echo off
echo Deploying CUDA DLL to required locations...

if not exist cuda_othello.dll (
    echo Error: cuda_othello.dll not found!
    exit /b 1
)

REM Copy to project root
copy cuda_othello.dll ..\ /Y
echo Copied DLL to project root

REM Create directories if they don't exist
set EVAL_DIR=..\models\ai\evaluation
set LEARN_DIR=..\models\ai\learning

if not exist %EVAL_DIR% mkdir %EVAL_DIR%
if not exist %LEARN_DIR% mkdir %LEARN_DIR%

copy cuda_othello.dll %EVAL_DIR%\ /Y
echo Copied DLL to evaluation directory

copy cuda_othello.dll %LEARN_DIR%\ /Y
echo Copied DLL to learning directory

echo DLL deployment complete.
