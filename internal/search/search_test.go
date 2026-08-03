package search

import (
	"hyperion/internal/board"
	"hyperion/internal/movegen"
	"testing"
	"time"
)

func TestSearchSingleAndMultiThreaded(t *testing.T) {
	b := board.New()
	b.SetFEN(board.StartFEN)

	// Single threaded
	s1 := NewSearcher(16)
	s1.Threads = 1
	m1, score1 := s1.Search(b, 5)

	if m1.String() == "" {
		t.Fatalf("Single-threaded search failed to return a move")
	}

	// Multi-threaded (4 threads)
	s4 := NewSearcher(16)
	s4.Threads = 4
	m4, score4 := s4.Search(b, 5)

	if m4.String() == "" {
		t.Fatalf("Multi-threaded search failed to return a move")
	}

	if s4.Nodes.Load() == 0 {
		t.Errorf("Expected positive node count for multi-threaded search")
	}

	t.Logf("1 Thread: bestmove=%s score=%d nodes=%d", m1.String(), score1, s1.Nodes.Load())
	t.Logf("4 Threads: bestmove=%s score=%d nodes=%d", m4.String(), score4, s4.Nodes.Load())
}

func TestSearchRootDoesNotInventBestMoveWhenStopped(t *testing.T) {
	b := board.New()
	if err := b.SetFEN("6k1/6pp/8/8/8/8/6PP/6K1 b - - 0 1"); err != nil {
		t.Fatalf("SetFEN failed: %v", err)
	}

	s := NewSearcher(16)
	w := &Worker{searcher: s, board: b}
	s.Stopped.Store(true)

	result := w.searchRoot(b, 8, -Infinity, Infinity)
	if result.complete {
		t.Fatalf("stopped root search should be marked incomplete")
	}
	if result.bestMove != 0 {
		t.Fatalf("stopped root search returned fabricated move %s", result.bestMove.String())
	}
}

func TestSearchWithExpiredClockReturnsLegalFallback(t *testing.T) {
	b := board.New()
	if err := b.SetFEN("6k1/6pp/8/8/8/8/6PP/6K1 b - - 0 1"); err != nil {
		t.Fatalf("SetFEN failed: %v", err)
	}

	s := NewSearcher(16)
	bestMove, _ := s.SearchWithLimits(b, Limits{Depth: 8, MoveTime: time.Nanosecond})
	if bestMove == 0 {
		t.Fatalf("expected legal fallback move, got null move")
	}

	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)
	for i := 0; i < list.Count; i++ {
		if list.Moves[i] == bestMove {
			return
		}
	}
	t.Fatalf("fallback move %s is not legal", bestMove.String())
}
