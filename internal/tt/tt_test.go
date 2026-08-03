package tt

import (
	"hyperion/internal/move"
	"sync"
	"testing"
)

func TestTTStoreProbe(t *testing.T) {
	table := NewTable(1)
	hash := uint64(0x123456789ABCDEF0)
	m := move.Move(0x4321)
	score := -150
	depth := 5
	flag := Exact

	table.Store(hash, depth, flag, score, m)

	entry, ok := table.Probe(hash)
	if !ok {
		t.Fatalf("Expected to find entry in TT")
	}

	if entry.Hash != hash {
		t.Errorf("Expected hash %x, got %x", hash, entry.Hash)
	}
	if entry.Score != int16(score) {
		t.Errorf("Expected score %d, got %d", score, entry.Score)
	}
	if entry.Depth != uint8(depth) {
		t.Errorf("Expected depth %d, got %d", depth, entry.Depth)
	}
	if entry.Flag != flag {
		t.Errorf("Expected flag %d, got %d", flag, entry.Flag)
	}
	if entry.Move != m {
		t.Errorf("Expected move %d, got %d", m, entry.Move)
	}
}

func TestTTConcurrentAccess(t *testing.T) {
	table := NewTable(4)
	var wg sync.WaitGroup

	numGoroutines := 16
	opsPerGoroutine := 1000

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				hash := uint64((gid * 100000) + i)
				m := move.Move(i & 0xFFFF)
				score := (i % 500) - 250
				depth := (i % 20) + 1

				table.Store(hash, depth, Exact, score, m)
				table.Probe(hash)
			}
		}(g)
	}

	wg.Wait()
}
