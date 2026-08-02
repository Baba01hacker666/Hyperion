package magic

import (
	"hyperion/internal/bitboard"
)

var (
	RookMasks   [64]bitboard.Bitboard
	BishopMasks [64]bitboard.Bitboard

	RookShifts   [64]int
	BishopShifts [64]int

	RookAttacks   [64][]bitboard.Bitboard
	BishopAttacks [64][]bitboard.Bitboard
)

func init() {
	for sq := 0; sq < 64; sq++ {
		RookMasks[sq] = maskRook(sq)
		BishopMasks[sq] = maskBishop(sq)

		RookShifts[sq] = 64 - RookMasks[sq].PopCount()
		BishopShifts[sq] = 64 - BishopMasks[sq].PopCount()

		RookAttacks[sq] = make([]bitboard.Bitboard, 1<<(64-RookShifts[sq]))
		BishopAttacks[sq] = make([]bitboard.Bitboard, 1<<(64-BishopShifts[sq]))

		initAttacks(sq, true)
		initAttacks(sq, false)
	}
}

func maskRook(sq int) bitboard.Bitboard {
	var mask bitboard.Bitboard
	r, f := sq/8, sq%8
	for tr := r + 1; tr <= 6; tr++ {
		mask.Set(tr*8 + f)
	}
	for tr := r - 1; tr >= 1; tr-- {
		mask.Set(tr*8 + f)
	}
	for tf := f + 1; tf <= 6; tf++ {
		mask.Set(r*8 + tf)
	}
	for tf := f - 1; tf >= 1; tf-- {
		mask.Set(r*8 + tf)
	}
	return mask
}

func maskBishop(sq int) bitboard.Bitboard {
	var mask bitboard.Bitboard
	r, f := sq/8, sq%8
	for tr, tf := r+1, f+1; tr <= 6 && tf <= 6; tr, tf = tr+1, tf+1 {
		mask.Set(tr*8 + tf)
	}
	for tr, tf := r+1, f-1; tr <= 6 && tf >= 1; tr, tf = tr+1, tf-1 {
		mask.Set(tr*8 + tf)
	}
	for tr, tf := r-1, f+1; tr >= 1 && tf <= 6; tr, tf = tr-1, tf+1 {
		mask.Set(tr*8 + tf)
	}
	for tr, tf := r-1, f-1; tr >= 1 && tf >= 1; tr, tf = tr-1, tf-1 {
		mask.Set(tr*8 + tf)
	}
	return mask
}

func attackRookSlow(sq int, block bitboard.Bitboard) bitboard.Bitboard {
	var attacks bitboard.Bitboard
	r, f := sq/8, sq%8
	for tr := r + 1; tr <= 7; tr++ {
		attacks.Set(tr*8 + f)
		if block.Has(tr*8 + f) {
			break
		}
	}
	for tr := r - 1; tr >= 0; tr-- {
		attacks.Set(tr*8 + f)
		if block.Has(tr*8 + f) {
			break
		}
	}
	for tf := f + 1; tf <= 7; tf++ {
		attacks.Set(r*8 + tf)
		if block.Has(r*8 + tf) {
			break
		}
	}
	for tf := f - 1; tf >= 0; tf-- {
		attacks.Set(r*8 + tf)
		if block.Has(r*8 + tf) {
			break
		}
	}
	return attacks
}

func attackBishopSlow(sq int, block bitboard.Bitboard) bitboard.Bitboard {
	var attacks bitboard.Bitboard
	r, f := sq/8, sq%8
	for tr, tf := r+1, f+1; tr <= 7 && tf <= 7; tr, tf = tr+1, tf+1 {
		attacks.Set(tr*8 + tf)
		if block.Has(tr*8 + tf) {
			break
		}
	}
	for tr, tf := r+1, f-1; tr <= 7 && tf >= 0; tr, tf = tr+1, tf-1 {
		attacks.Set(tr*8 + tf)
		if block.Has(tr*8 + tf) {
			break
		}
	}
	for tr, tf := r-1, f+1; tr >= 0 && tf <= 7; tr, tf = tr-1, tf+1 {
		attacks.Set(tr*8 + tf)
		if block.Has(tr*8 + tf) {
			break
		}
	}
	for tr, tf := r-1, f-1; tr >= 0 && tf >= 0; tr, tf = tr-1, tf-1 {
		attacks.Set(tr*8 + tf)
		if block.Has(tr*8 + tf) {
			break
		}
	}
	return attacks
}

func set2bb(index int, bits int, mask bitboard.Bitboard) bitboard.Bitboard {
	var b bitboard.Bitboard
	for i := 0; i < bits; i++ {
		sq := mask.PopLSB()
		if (index & (1 << i)) != 0 {
			b.Set(sq)
		}
	}
	return b
}

func initAttacks(sq int, isRook bool) {
	var mask bitboard.Bitboard
	var bitsCount int
	if isRook {
		mask = RookMasks[sq]
		bitsCount = mask.PopCount()
	} else {
		mask = BishopMasks[sq]
		bitsCount = mask.PopCount()
	}

	numOccupancies := 1 << bitsCount
	for i := 0; i < numOccupancies; i++ {
		occ := set2bb(i, bitsCount, mask)
		if isRook {
			magicIndex := (uint64(occ) * RookMagics[sq]) >> RookShifts[sq]
			RookAttacks[sq][magicIndex] = attackRookSlow(sq, occ)
		} else {
			magicIndex := (uint64(occ) * BishopMagics[sq]) >> BishopShifts[sq]
			BishopAttacks[sq][magicIndex] = attackBishopSlow(sq, occ)
		}
	}
}

// GetRookAttacks returns the rook attacks for a given square and occupancy.
func GetRookAttacks(sq int, occ bitboard.Bitboard) bitboard.Bitboard {
	occ &= RookMasks[sq]
	occ = bitboard.Bitboard((uint64(occ) * RookMagics[sq]) >> RookShifts[sq])
	return RookAttacks[sq][occ]
}

// GetBishopAttacks returns the bishop attacks for a given square and occupancy.
func GetBishopAttacks(sq int, occ bitboard.Bitboard) bitboard.Bitboard {
	occ &= BishopMasks[sq]
	occ = bitboard.Bitboard((uint64(occ) * BishopMagics[sq]) >> BishopShifts[sq])
	return BishopAttacks[sq][occ]
}

// GetQueenAttacks returns the queen attacks (rook + bishop).
func GetQueenAttacks(sq int, occ bitboard.Bitboard) bitboard.Bitboard {
	return GetRookAttacks(sq, occ) | GetBishopAttacks(sq, occ)
}
