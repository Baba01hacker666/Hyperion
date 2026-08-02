package board

import (
	"testing"
)

func TestAddRemovePiece(t *testing.T) {
	b := New()

	if b.PieceAt(E4) != NoPiece {
		t.Errorf("Expected NoPiece at E4")
	}

	b.AddPiece(E4, WhiteKnight)
	if b.PieceAt(E4) != WhiteKnight {
		t.Errorf("Expected WhiteKnight at E4")
	}

	if !b.Pieces[Knight].Has(int(E4)) {
		t.Errorf("Expected WhiteKnight to be set in Knight bitboard")
	}
	if !b.Colors[White].Has(int(E4)) {
		t.Errorf("Expected WhiteKnight to be set in White bitboard")
	}

	p := b.RemovePiece(E4)
	if p != WhiteKnight {
		t.Errorf("Expected RemovePiece to return WhiteKnight")
	}
	if b.PieceAt(E4) != NoPiece {
		t.Errorf("Expected NoPiece at E4 after removal")
	}
	if b.Pieces[Knight].Has(int(E4)) {
		t.Errorf("Expected WhiteKnight to be cleared in Knight bitboard")
	}
	if b.Colors[White].Has(int(E4)) {
		t.Errorf("Expected WhiteKnight to be cleared in White bitboard")
	}
}

func TestAllPieces(t *testing.T) {
	b := New()
	b.AddPiece(E4, WhiteKnight)
	b.AddPiece(D5, BlackPawn)

	all := b.AllPieces()
	if !all.Has(int(E4)) || !all.Has(int(D5)) {
		t.Errorf("Expected AllPieces to contain E4 and D5")
	}
	if all.PopCount() != 2 {
		t.Errorf("Expected AllPieces to have 2 bits set")
	}
}
