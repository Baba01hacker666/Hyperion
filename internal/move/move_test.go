package move

import (
	"testing"
)

func TestMoveCreation(t *testing.T) {
	// E2 (12) to E4 (28)
	m := New(12, 28, DoublePawnPush)

	if m.From() != 12 {
		t.Errorf("Expected From square 12, got %v", m.From())
	}
	if m.To() != 28 {
		t.Errorf("Expected To square 28, got %v", m.To())
	}
	if m.Flag() != DoublePawnPush {
		t.Errorf("Expected Flag DoublePawnPush, got %v", m.Flag())
	}
	if m.String() != "e2e4" {
		t.Errorf("Expected String e2e4, got %s", m.String())
	}
	if m.IsPromotion() {
		t.Errorf("Did not expect promotion")
	}
	if m.IsCapture() {
		t.Errorf("Did not expect capture")
	}
}

func TestPromotionCapture(t *testing.T) {
	// E7 (52) to F8 (61)
	m := New(52, 61, QueenPromoCap)

	if m.From() != 52 {
		t.Errorf("Expected From square 52")
	}
	if m.To() != 61 {
		t.Errorf("Expected To square 61")
	}
	if !m.IsPromotion() {
		t.Errorf("Expected move to be a promotion")
	}
	if !m.IsCapture() {
		t.Errorf("Expected move to be a capture")
	}
	if m.PromotionType() != 4 { // Queen
		t.Errorf("Expected promotion type Queen (4)")
	}
	if m.String() != "e7f8q" {
		t.Errorf("Expected String e7f8q, got %s", m.String())
	}
}
