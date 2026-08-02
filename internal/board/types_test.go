package board

import (
	"testing"
)

func TestColor(t *testing.T) {
	if White.Opposite() != Black {
		t.Errorf("Expected White's opposite to be Black")
	}
	if Black.Opposite() != White {
		t.Errorf("Expected Black's opposite to be White")
	}
}

func TestPiece(t *testing.T) {
	tests := []struct {
		color Color
		pt    PieceType
		piece Piece
	}{
		{White, Pawn, WhitePawn},
		{White, Knight, WhiteKnight},
		{Black, Queen, BlackQueen},
		{Black, King, BlackKing},
		{Both, NoPieceType, NoPiece},
	}

	for _, tt := range tests {
		p := NewPiece(tt.color, tt.pt)
		if p != tt.piece {
			t.Errorf("Expected %v, got %v", tt.piece, p)
		}
		if p.Type() != tt.pt {
			t.Errorf("Expected piece type %v, got %v", tt.pt, p.Type())
		}
		if p != NoPiece && p.Color() != tt.color {
			t.Errorf("Expected piece color %v, got %v", tt.color, p.Color())
		}
	}
}

func TestSquare(t *testing.T) {
	if NewSquare(FileA, Rank1) != A1 {
		t.Errorf("Expected A1")
	}
	if NewSquare(FileH, Rank8) != H8 {
		t.Errorf("Expected H8")
	}
	if NewSquare(FileE, Rank4) != E4 {
		t.Errorf("Expected E4")
	}

	if E4.File() != FileE {
		t.Errorf("Expected FileE")
	}
	if E4.Rank() != Rank4 {
		t.Errorf("Expected Rank4")
	}
}
