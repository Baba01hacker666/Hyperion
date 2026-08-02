package attack

import (
	"hyperion/internal/bitboard"
	"hyperion/internal/board"
)

// Precalculated attack tables for non-sliding pieces.
var (
	PawnAttacks   [2][64]bitboard.Bitboard
	KnightAttacks [64]bitboard.Bitboard
	KingAttacks   [64]bitboard.Bitboard
)

func init() {
	initLeapers()
}

func initLeapers() {
	for sq := 0; sq < 64; sq++ {
		// 1. Pawn Attacks
		var wAttacks, bAttacks bitboard.Bitboard
		r := sq / 8
		f := sq % 8

		if r < 7 { // White pawns attack upwards
			if f > 0 {
				wAttacks.Set(sq + 7)
			}
			if f < 7 {
				wAttacks.Set(sq + 9)
			}
		}
		PawnAttacks[board.White][sq] = wAttacks

		if r > 0 { // Black pawns attack downwards
			if f > 0 {
				bAttacks.Set(sq - 9)
			}
			if f < 7 {
				bAttacks.Set(sq - 7)
			}
		}
		PawnAttacks[board.Black][sq] = bAttacks

		// 2. Knight Attacks
		KnightAttacks[sq] = maskKnightAttacks(sq)

		// 3. King Attacks
		KingAttacks[sq] = maskKingAttacks(sq)
	}
}

func maskKnightAttacks(sq int) bitboard.Bitboard {
	var attacks bitboard.Bitboard
	r := sq / 8
	f := sq % 8

	moves := [][2]int{
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
		{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
	}

	for _, m := range moves {
		nr := r + m[0]
		nf := f + m[1]
		if nr >= 0 && nr <= 7 && nf >= 0 && nf <= 7 {
			attacks.Set(nr*8 + nf)
		}
	}
	return attacks
}

func maskKingAttacks(sq int) bitboard.Bitboard {
	var attacks bitboard.Bitboard
	r := sq / 8
	f := sq % 8

	moves := [][2]int{
		{1, 0}, {1, 1}, {1, -1},
		{0, 1}, {0, -1},
		{-1, 0}, {-1, 1}, {-1, -1},
	}

	for _, m := range moves {
		nr := r + m[0]
		nf := f + m[1]
		if nr >= 0 && nr <= 7 && nf >= 0 && nf <= 7 {
			attacks.Set(nr*8 + nf)
		}
	}
	return attacks
}
