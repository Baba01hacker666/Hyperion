package bitboard

import (
	"testing"
)

func TestBitboard(t *testing.T) {
	var b Bitboard

	if b != Empty {
		t.Errorf("Expected empty bitboard, got %d", b)
	}

	b.Set(0)
	if !b.Has(0) {
		t.Errorf("Expected bit 0 to be set")
	}
	if b.Has(1) {
		t.Errorf("Expected bit 1 to not be set")
	}
	if b.PopCount() != 1 {
		t.Errorf("Expected PopCount 1, got %d", b.PopCount())
	}

	b.Set(63)
	if !b.Has(63) {
		t.Errorf("Expected bit 63 to be set")
	}
	if b.PopCount() != 2 {
		t.Errorf("Expected PopCount 2, got %d", b.PopCount())
	}

	b.Clear(0)
	if b.Has(0) {
		t.Errorf("Expected bit 0 to be clear")
	}
	if b.PopCount() != 1 {
		t.Errorf("Expected PopCount 1, got %d", b.PopCount())
	}

	b.Toggle(63)
	if b.Has(63) {
		t.Errorf("Expected bit 63 to be clear after toggle")
	}

	b.Toggle(8)
	if !b.Has(8) {
		t.Errorf("Expected bit 8 to be set after toggle")
	}
}

func TestLSB_MSB_PopLSB(t *testing.T) {
	b := Empty

	if b.LSB() != 64 {
		t.Errorf("Expected LSB of empty bitboard to be 64, got %d", b.LSB())
	}
	if b.MSB() != 64 {
		t.Errorf("Expected MSB of empty bitboard to be 64, got %d", b.MSB())
	}
	if b.PopLSB() != 64 {
		t.Errorf("Expected PopLSB of empty bitboard to be 64, got %d", b.PopLSB())
	}

	b.Set(10)
	b.Set(20)

	if b.LSB() != 10 {
		t.Errorf("Expected LSB 10, got %d", b.LSB())
	}
	if b.MSB() != 20 {
		t.Errorf("Expected MSB 20, got %d", b.MSB())
	}

	sq := b.PopLSB()
	if sq != 10 {
		t.Errorf("Expected PopLSB to return 10, got %d", sq)
	}
	if b.Has(10) {
		t.Errorf("Expected bit 10 to be cleared after PopLSB")
	}
	if b.LSB() != 20 {
		t.Errorf("Expected LSB 20, got %d", b.LSB())
	}
}

func TestString(t *testing.T) {
	b := Empty
	b.Set(0)
	b.Set(63)
	str := b.String()

	if len(str) == 0 {
		t.Errorf("Expected non-empty string representation")
	}
}
