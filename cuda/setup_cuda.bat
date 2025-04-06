@echo off
setlocal

REM Find CUDA installation
set CUDA_PATH=
for %%v in (12.8 12.0 11.8 11.7 11.6 11.5 11.4 11.3 11.2 11.1 11.0) do (
    if exist "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v%%v\bin\cudart64_*.dll" (
        set "CUDA_PATH=C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v%%v"
        goto :found
    )
)

:found
if "%CUDA_PATH%"=="" (
    echo Error: CUDA installation not found
    exit /b 1
)

echo Found CUDA installation at %CUDA_PATH%

REM Copy required CUDA libraries to our cuda directory
echo Copying CUDA runtime libraries...
copy "%CUDA_PATH%\bin\cudart64_*.dll" .
echo Done!

echo Setup complete. You can now build and run the application.
