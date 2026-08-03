package movegen

import (
	"hyperion/internal/board"
	"hyperion/internal/move"
	"sync"
	"sync/atomic"
)

// Perft calculates the number of leaf nodes at a given depth recursively.
func Perft(b *board.Board, depth int) uint64 {
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
		nodes += Perft(b, depth-1)
		b.UnmakeMove(&undo)
	}
	return nodes
}

// PerftParallel calculates Perft using multiple worker goroutines at the root level.
func PerftParallel(b *board.Board, depth int, workers int) uint64 {
	if depth == 0 {
		return 1
	}

	list := &MoveList{}
	GenerateLegalMoves(b, list)

	if list.Count == 0 {
		return 0
	}

	if workers <= 1 {
		return Perft(b, depth)
	}

	var totalNodes atomic.Uint64
	var wg sync.WaitGroup

	moveChan := make(chan move.Move, list.Count)
	for i := 0; i < list.Count; i++ {
		moveChan <- list.Moves[i]
	}
	close(moveChan)

	numWorkers := workers
	if numWorkers > list.Count {
		numWorkers = list.Count
	}

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			boardCopy := b.Clone()
			var undo board.Undo

			for m := range moveChan {
				boardCopy.MakeMove(m, &undo)
				subNodes := Perft(boardCopy, depth-1)
				boardCopy.UnmakeMove(&undo)
				totalNodes.Add(subNodes)
			}
		}()
	}

	wg.Wait()
	return totalNodes.Load()
}
