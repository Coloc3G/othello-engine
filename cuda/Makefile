# Makefile for CUDA Othello library

CUDA_PATH ?= /usr/local/cuda
NVCC       = nvcc
CFLAGS     = -O3
CUDA_FLAGS = -O3 -Xcompiler -fPIC

SOURCES = othello_cuda.cu
OUTPUT  = libcuda_othello.so
OUTPUT_WIN = cuda_othello.dll

all: $(OUTPUT)

windows: $(OUTPUT_WIN)

$(OUTPUT): $(SOURCES)
	$(NVCC) $(CUDA_FLAGS) --shared -o $@ $<

$(OUTPUT_WIN): $(SOURCES)
	$(NVCC) $(CUDA_FLAGS) --shared -o $@ $<

clean:
	rm -f $(OUTPUT) $(OUTPUT_WIN)
