package zobrist

import (
	"math/rand"
)

var (
	// PieceSquare keys: [color][pieceType][square]
	PieceSquare [2][6][64]uint64
	SideToMove  uint64
	Castle      [16]uint64
	EnPassant   [64]uint64
)

func init() {
	// Use a deterministic PRNG seed for reproducible Zobrist keys
	rng := rand.New(rand.NewSource(1070372))

	for c := 0; c < 2; c++ {
		for pt := 0; pt < 6; pt++ {
			for sq := 0; sq < 64; sq++ {
				PieceSquare[c][pt][sq] = rng.Uint64()
			}
		}
	}

	SideToMove = rng.Uint64()

	for i := 0; i < 16; i++ {
		Castle[i] = rng.Uint64()
	}

	for i := 0; i < 64; i++ {
		EnPassant[i] = rng.Uint64()
	}
}
