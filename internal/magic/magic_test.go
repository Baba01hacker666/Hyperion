package magic

import (
	"hyperion/internal/bitboard"
	"testing"
)

func TestRookAttacks(t *testing.T) {
	// Empty board, rook on E4
	occ := bitboard.Empty
	attacks := GetRookAttacks(28, occ) // 28 is E4

	// Expected squares: file E (4, 12, 20, 36, 44, 52, 60), rank 4 (24, 25, 26, 27, 29, 30, 31)
	expectedSq := []int{4, 12, 20, 36, 44, 52, 60, 24, 25, 26, 27, 29, 30, 31}
	for _, sq := range expectedSq {
		if !attacks.Has(sq) {
			t.Errorf("Expected Rook on E4 to attack square %d", sq)
		}
	}
	if attacks.PopCount() != 14 {
		t.Errorf("Expected 14 attack squares, got %d", attacks.PopCount())
	}

	// Blocked board
	occ.Set(36) // Block at E5
	occ.Set(26) // Block at C4
	attacks = GetRookAttacks(28, occ)

	if attacks.Has(44) {
		t.Errorf("Expected E6 to not be attacked because of E5 block")
	}
	if !attacks.Has(36) {
		t.Errorf("Expected block E5 to be attacked")
	}
	if attacks.Has(25) {
		t.Errorf("Expected B4 to not be attacked because of C4 block")
	}
	if !attacks.Has(26) {
		t.Errorf("Expected block C4 to be attacked")
	}
}

func TestBishopAttacks(t *testing.T) {
	// Empty board, bishop on E4
	occ := bitboard.Empty
	attacks := GetBishopAttacks(28, occ) // 28 is E4

	expectedSq := []int{
		37, 46, 55, // Top Right
		35, 42, 49, 56, // Top Left
		21, 14, 7, // Bottom Right
		19, 10, 1, // Bottom Left
	}
	for _, sq := range expectedSq {
		if !attacks.Has(sq) {
			t.Errorf("Expected Bishop on E4 to attack square %d", sq)
		}
	}
	if attacks.PopCount() != 13 {
		t.Errorf("Expected 13 attack squares, got %d", attacks.PopCount())
	}
}
