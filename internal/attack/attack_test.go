package attack

import (
	"hyperion/internal/board"
	"testing"
)

func TestPawnAttacks(t *testing.T) {
	// White Pawn on E4
	wAttacks := PawnAttacks[board.White][int(board.E4)]
	if !wAttacks.Has(int(board.D5)) || !wAttacks.Has(int(board.F5)) {
		t.Errorf("Expected White Pawn on E4 to attack D5 and F5")
	}
	if wAttacks.PopCount() != 2 {
		t.Errorf("Expected exactly 2 attack squares for pawn on E4")
	}

	// Black Pawn on A7
	bAttacks := PawnAttacks[board.Black][int(board.A7)]
	if !bAttacks.Has(int(board.B6)) {
		t.Errorf("Expected Black Pawn on A7 to attack B6")
	}
	if bAttacks.PopCount() != 1 {
		t.Errorf("Expected exactly 1 attack square for A file pawn")
	}
}

func TestKnightAttacks(t *testing.T) {
	// Knight on E4
	attacks := KnightAttacks[int(board.E4)]
	expectedSquares := []board.Square{
		board.D6, board.F6, board.C5, board.G5,
		board.C3, board.G3, board.D2, board.F2,
	}

	for _, sq := range expectedSquares {
		if !attacks.Has(int(sq)) {
			t.Errorf("Expected Knight on E4 to attack %v", sq)
		}
	}
	if attacks.PopCount() != 8 {
		t.Errorf("Expected exactly 8 attack squares for Knight on E4")
	}

	// Knight on A1
	a1Attacks := KnightAttacks[int(board.A1)]
	if !a1Attacks.Has(int(board.B3)) || !a1Attacks.Has(int(board.C2)) {
		t.Errorf("Expected Knight on A1 to attack B3 and C2")
	}
	if a1Attacks.PopCount() != 2 {
		t.Errorf("Expected exactly 2 attack squares for Knight on A1")
	}
}

func TestKingAttacks(t *testing.T) {
	// King on E4
	attacks := KingAttacks[int(board.E4)]
	expectedSquares := []board.Square{
		board.D5, board.E5, board.F5,
		board.D4, board.F4,
		board.D3, board.E3, board.F3,
	}

	for _, sq := range expectedSquares {
		if !attacks.Has(int(sq)) {
			t.Errorf("Expected King on E4 to attack %v", sq)
		}
	}
	if attacks.PopCount() != 8 {
		t.Errorf("Expected exactly 8 attack squares for King on E4")
	}
}
