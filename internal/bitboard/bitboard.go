package bitboard

import (
	"fmt"
	"math/bits"
	"strings"
)

// Bitboard represents a set of squares on a chessboard.
type Bitboard uint64

// Predefined commonly used bitboards.
const (
	Empty    Bitboard = 0
	Universe Bitboard = 0xFFFFFFFFFFFFFFFF

	FileA Bitboard = 0x0101010101010101
	FileB Bitboard = FileA << 1
	FileC Bitboard = FileA << 2
	FileD Bitboard = FileA << 3
	FileE Bitboard = FileA << 4
	FileF Bitboard = FileA << 5
	FileG Bitboard = FileA << 6
	FileH Bitboard = FileA << 7

	Rank1 Bitboard = 0xFF
	Rank2 Bitboard = Rank1 << 8
	Rank3 Bitboard = Rank1 << 16
	Rank4 Bitboard = Rank1 << 24
	Rank5 Bitboard = Rank1 << 32
	Rank6 Bitboard = Rank1 << 40
	Rank7 Bitboard = Rank1 << 48
	Rank8 Bitboard = Rank1 << 56
)

var FileMasks = [8]Bitboard{FileA, FileB, FileC, FileD, FileE, FileF, FileG, FileH}
var RankMasks = [8]Bitboard{Rank1, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8}

// Set sets the bit at the given square index (0-63).
func (b *Bitboard) Set(sq int) {
	*b |= (1 << sq)
}

// Clear clears the bit at the given square index (0-63).
func (b *Bitboard) Clear(sq int) {
	*b &= ^(1 << sq)
}

// Has checks if the bit at the given square index is set.
func (b Bitboard) Has(sq int) bool {
	return (b & (1 << sq)) != 0
}

// Toggle toggles the bit at the given square index (0-63).
func (b *Bitboard) Toggle(sq int) {
	*b ^= (1 << sq)
}

// PopCount returns the number of set bits.
func (b Bitboard) PopCount() int {
	return bits.OnesCount64(uint64(b))
}

// LSB returns the index of the least significant bit (0-63).
// Returns 64 if the bitboard is empty.
func (b Bitboard) LSB() int {
	if b == Empty {
		return 64
	}
	return bits.TrailingZeros64(uint64(b))
}

// MSB returns the index of the most significant bit (0-63).
// Returns 64 if the bitboard is empty.
func (b Bitboard) MSB() int {
	if b == Empty {
		return 64
	}
	return 63 - bits.LeadingZeros64(uint64(b))
}

// PopLSB clears the least significant bit and returns its index.
// Returns 64 if the bitboard is empty.
func (b *Bitboard) PopLSB() int {
	sq := b.LSB()
	if sq != 64 {
		*b &= *b - 1
	}
	return sq
}

// String returns a visual representation of the bitboard.
func (b Bitboard) String() string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file <= 7; file++ {
			sq := rank*8 + file
			if b.Has(sq) {
				sb.WriteString("1 ")
			} else {
				sb.WriteString(". ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Print prints the bitboard visual representation to standard output.
func (b Bitboard) Print() {
	fmt.Print(b.String())
}
