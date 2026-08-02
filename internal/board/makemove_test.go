package board

import (
	"hyperion/internal/move"
	"testing"
)

func TestMakeUnmakeMove(t *testing.T) {
	b := New()
	b.SetFEN(StartFEN)
	startFEN := b.FEN()

	// e2e4
	m := move.New(int(E2), int(E4), move.DoublePawnPush)
	var undo Undo
	b.MakeMove(m, &undo)

	if b.PieceAt(E2) != NoPiece {
		t.Errorf("Expected e2 to be empty after move")
	}
	if b.PieceAt(E4) != WhitePawn {
		t.Errorf("Expected white pawn on e4 after move")
	}
	if b.SideToMove != Black {
		t.Errorf("Expected side to move to be black")
	}
	if b.EnPassant != E3 {
		t.Errorf("Expected en passant square to be e3")
	}

	b.UnmakeMove(&undo)

	if b.FEN() != startFEN {
		t.Errorf("FEN mismatch after unmake.\nExpected: %s\nGot:      %s", startFEN, b.FEN())
	}
}

func TestMakeUnmakeCastling(t *testing.T) {
	b := New()
	b.SetFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	startFEN := b.FEN()

	// White short castle
	m := move.New(int(E1), int(G1), move.KingCastle)
	var undo Undo
	b.MakeMove(m, &undo)

	if b.PieceAt(G1) != WhiteKing || b.PieceAt(F1) != WhiteRook {
		t.Errorf("White kingside castling failed")
	}
	if b.Castle&(WhiteOO|WhiteOOO) != 0 {
		t.Errorf("White castling rights should be lost after castling")
	}

	b.UnmakeMove(&undo)
	if b.FEN() != startFEN {
		t.Errorf("FEN mismatch after unmake castling.\nExpected: %s\nGot:      %s", startFEN, b.FEN())
	}
}
