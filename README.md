# Hyperion

Hyperion is a modern, high-performance chess engine written entirely in Go (Golang) with zero external dependencies and no CGO. It features a clean modular architecture and fully complies with the Universal Chess Interface (UCI) protocol.

## Architecture

*   **`internal/bitboard`**: Provides foundational 64-bit board representations and fast bit-twiddling operations (`PopCount`, `LSB`, `MSB`).
*   **`internal/board`**: Represents the chess board state, Piece/Color enums, FEN parsing, and highly efficient `MakeMove`/`UnmakeMove` tracking via a lightweight `Undo` stack.
*   **`internal/magic`**: Implements Magic Bitboards to instantaneously generate sliding piece attacks (Rooks, Bishops, Queens) using a precomputed hash table for O(1) lookups.
*   **`internal/attack`**: Provides static attack tables for leapers (Knights, Kings, Pawns) and exposes a fast `IsSquareAttacked` function to validate king safety and castling rights.
*   **`internal/move`**: Encodes a chess move into a highly compressed, standard 16-bit integer minimizing memory overhead across search trees.
*   **`internal/movegen`**: Generates pseudo-legal moves and strictly legal moves. Verified for absolute correctness using recursive Perft tests on the starting and Kiwipete positions.
*   **`internal/evaluation`**: Provides static board evaluation taking into account basic material balances and simplified Piece-Square Tables (PST) to encourage rapid development and centralized control.
*   **`internal/search`**: A recursive Alpha-Beta pruning search algorithm with mate-distance biases to evaluate and select the best moves.
*   **`internal/uci`**: Implements the Universal Chess Interface protocol, allowing Hyperion to connect to any standard chess GUI.
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
position startpos
go depth 5
quit
```

## Testing

The engine is heavily unit-tested. To run the full suite (including Perft tests):

```bash
go test -v ./...
```

## Roadmap

The engine's foundation is structurally complete. Upcoming milestones include:
*   Zobrist Hashing & Transposition Tables (TT)
*   Iterative Deepening & Quiescence Search (QS)
*   Move Ordering (MVV-LVA, Killer Heuristics)
*   Null-Move Pruning (NMP) & Late Move Reductions (LMR)
*   Syzygy Tablebase Probing
