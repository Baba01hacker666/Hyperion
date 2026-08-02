package evaluation

import (
	"hyperion/internal/board"
	"testing"
)

func TestEvaluateStartPos(t *testing.T) {
	b := board.New()
	b.SetFEN(board.StartFEN)

	eval := Evaluate(b)

	// Start position should be perfectly equal (0)
	if eval != 0 {
		t.Errorf("Expected evaluation of 0 for starting position, got %d", eval)
	}
}

func TestEvaluateAdvantage(t *testing.T) {
	b := board.New()
	// White has an extra Queen
	b.SetFEN("rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	eval := Evaluate(b)
	if eval <= 500 {
		t.Errorf("Expected evaluation > 500, got %d", eval)
	}

	// Black has an extra Queen, but it's Black to move.
	// Since evaluate is from the perspective of the side to move, black being ahead means the evaluation for black should be positive!
	b.SetFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR b KQkq - 0 1")
	eval2 := Evaluate(b)

	if eval2 <= 500 {
		t.Errorf("Expected evaluation > 500 for black, got %d", eval2)
	}
}
