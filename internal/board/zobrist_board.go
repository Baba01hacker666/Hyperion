package board

import (
	"hyperion/internal/zobrist"
)

// ComputeHash computes and returns the Zobrist hash of the current board state from scratch.
func (b *Board) ComputeHash() uint64 {
	var hash uint64

	for sq := 0; sq < 64; sq++ {
		p := b.Squares[sq]
		if p != NoPiece {
			hash ^= zobrist.PieceSquare[p.Color()][p.Type()][sq]
		}
	}

	if b.SideToMove == Black {
		hash ^= zobrist.SideToMove
	}

	hash ^= zobrist.Castle[b.Castle]

	if b.EnPassant != NoSquare {
		hash ^= zobrist.EnPassant[b.EnPassant]
	}

	return hash
}
