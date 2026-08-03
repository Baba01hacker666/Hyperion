package movegen

import (
	"hyperion/internal/board"
	"testing"
)

func TestPerftStartPos(t *testing.T) {
	b := board.New()
	b.SetFEN(board.StartFEN)

	expected := []uint64{
		1,      // Depth 0
		20,     // Depth 1
		400,    // Depth 2
		8902,   // Depth 3
		197281, // Depth 4
	}

	for depth := 1; depth <= 3; depth++ {
		nodes := Perft(b, depth)
		if nodes != expected[depth] {
			t.Errorf("Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], nodes)
		}

		pNodes := PerftParallel(b, depth, 4)
		if pNodes != expected[depth] {
			t.Errorf("Parallel Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], pNodes)
		}
	}
}

func TestPerftKiwipete(t *testing.T) {
	b := board.New()
	b.SetFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")

	expected := []uint64{
		1,     // Depth 0
		48,    // Depth 1
		2039,  // Depth 2
		97862, // Depth 3
	}

	for depth := 1; depth <= 2; depth++ {
		nodes := Perft(b, depth)
		if nodes != expected[depth] {
			t.Errorf("Kiwipete Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], nodes)
		}

		pNodes := PerftParallel(b, depth, 4)
		if pNodes != expected[depth] {
			t.Errorf("Kiwipete Parallel Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], pNodes)
		}
	}
}
