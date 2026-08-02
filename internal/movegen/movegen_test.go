package movegen

import (
	"hyperion/internal/board"
	"testing"
)

func perft(b *board.Board, depth int) uint64 {
	if depth == 0 {
		return 1
	}

	var nodes uint64 = 0
	list := &MoveList{}
	GenerateLegalMoves(b, list)

	var undo board.Undo
	for i := 0; i < list.Count; i++ {
		m := list.Moves[i]
		b.MakeMove(m, &undo)
		nodes += perft(b, depth-1)
		b.UnmakeMove(&undo)
	}
	return nodes
}

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

	for depth := 1; depth <= 3; depth++ { // Testing up to depth 3 for speed
		nodes := perft(b, depth)
		if nodes != expected[depth] {
			t.Errorf("Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], nodes)
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

	for depth := 1; depth <= 2; depth++ { // Testing up to depth 2 for speed
		nodes := perft(b, depth)
		if nodes != expected[depth] {
			t.Errorf("Kiwipete Perft Depth %d Failed. Expected: %d, Got: %d", depth, expected[depth], nodes)
		}
	}
}
