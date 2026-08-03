# Hyperion

Hyperion is a modern, high-performance chess engine written entirely in Go (Golang) with zero external dependencies and no CGO. It features a clean modular architecture, lockless atomic Transposition Tables, multi-threaded Lazy SMP search, and full compliance with the Universal Chess Interface (UCI) protocol.

## Architecture

*   **`internal/bitboard`**: Provides foundational 64-bit board representations and fast bit-twiddling operations (`PopCount`, `LSB`, `MSB`).
*   **`internal/board`**: Represents the chess board state, Piece/Color enums, FEN parsing, lightweight `Undo` stack, and `Clone()` for thread-safe concurrent searches.
*   **`internal/tt`**: Lockless, thread-safe Transposition Table using atomic bit packing for zero-lock concurrent reads and writes across worker threads.
*   **`internal/magic`**: Implements Magic Bitboards to instantaneously generate sliding piece attacks (Rooks, Bishops, Queens) using a precomputed hash table for O(1) lookups.
*   **`internal/attack`**: Provides static attack tables for leapers (Knights, Kings, Pawns) and exposes a fast `IsSquareAttacked` function to validate king safety and castling rights.
*   **`internal/move`**: Encodes a chess move into a highly compressed, standard 16-bit integer minimizing memory overhead across search trees.
*   **`internal/movegen`**: Generates pseudo-legal moves and strictly legal moves, with single-threaded (`Perft`) and multi-threaded (`PerftParallel`) validation functions.
*   **`internal/evaluation`**: Provides static board evaluation taking into account basic material balances and simplified Piece-Square Tables (PST) to encourage rapid development and centralized control.
*   **`internal/search`**: Multi-threaded Lazy SMP search algorithm with Iterative Deepening, Aspiration Windows, PVS, Null-Move Pruning, and Late Move Reductions (LMR).
*   **`internal/uci`**: Implements the Universal Chess Interface protocol, supporting `Threads`, `Hash`, and `Style` options.
*   **`cmd/hyperion`**: The main entry point to compile the engine executable.
*   **`cmd/magics`**: An offline generator script (`go generate`) used to discover collision-free magic numbers.

## Building

```bash
go build -o bin/hyperion cmd/hyperion/main.go
```

## Running

Launch the binary to enter the UCI loop:

```bash
./bin/hyperion
```

You can type UCI commands directly:
```text
uci
setoption name Threads value 4
position startpos
go depth 6
quit
```

## Testing

The engine is heavily unit-tested. To run the full suite (including parallel perft and multi-threaded search tests):

```bash
go test -v ./...
```

## Features

*   **Lazy SMP Multithreading**: Multi-worker concurrent iterative deepening with zero-overhead thread synchronization.
*   **Lockless Atomic Transposition Table**: Atomic hash and payload packing allowing safe parallel TT access without mutex lock contention.
*   **Parallel Perft**: Multi-threaded root move distribution for high-speed perft tree verification.

