package attack

import (
	"hyperion/internal/board"
	"testing"
)

func TestIsSquareAttacked(t *testing.T) {
	b := board.New()
	b.SetFEN(board.StartFEN)

	// E4 is empty in start pos
	if IsSquareAttacked(b, board.E4, board.White) {
		t.Errorf("E4 should not be attacked by white in start pos")
	}

	// But D3 is attacked by the C2 and E2 pawns
	if !IsSquareAttacked(b, board.D3, board.White) {
		t.Errorf("D3 should be attacked by white pawns in start pos")
	}

	// Test knight attack
	if !IsSquareAttacked(b, board.C3, board.White) {
		t.Errorf("C3 should be attacked by B1 Knight")
	}

	// Custom position to test sliders
	b.SetFEN("8/8/8/3Q4/8/8/8/8 w - - 0 1")
	if !IsSquareAttacked(b, board.D8, board.White) { // Rook-wise attack
		t.Errorf("D8 should be attacked by Q on D5")
	}
	if !IsSquareAttacked(b, board.A2, board.White) { // Bishop-wise attack
		t.Errorf("A2 should be attacked by Q on D5")
	}
	if IsSquareAttacked(b, board.A2, board.Black) {
		t.Errorf("A2 should not be attacked by black")
	}

	// Blocked slider attack
	b.SetFEN("8/8/8/3Q4/3P4/8/3K4/8 w - - 0 1") // D4 blocks D2
	if IsSquareAttacked(b, board.D2, board.White) {
		t.Errorf("D2 should NOT be attacked by Q on D5 because D4 blocks it")
	}
}
