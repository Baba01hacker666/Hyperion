package search

import (
	"hyperion/internal/board"
	"testing"
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
