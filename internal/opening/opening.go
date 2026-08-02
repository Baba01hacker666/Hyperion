package opening

import (
	"hyperion/internal/board"
	"hyperion/internal/move"
	"hyperion/internal/movegen"
	"math/rand"
	"time"
)

// Built-in Grandmaster Opening Book lines (FEN -> moves)
var defaultBook = map[string][]string{
	// Starting position (1. e4, 1. d4, 1. Nf3, 1. c4)
	board.StartFEN: {"e2e4", "d2d4", "g1f3", "c2c4"},

	// 1. e4 responses
	"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1": {"c7c5", "e7e5", "e7e6", "c7c6"},
	"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2": {"g1f3", "b1c3", "f2f4"},

	// 1. d4 responses
	"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 1": {"g8f6", "d7d5", "e7e6"},

	// Italian Game / Ruy Lopez
	"rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2":   {"b8c6", "g8f6"},
	"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3": {"f1c4", "f1b5", "d2d4"},

	// Sicilian Defense (1. e4 c5 2. Nf3)
	"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2": {"g1f3", "c2c3", "b1c3"},
	"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2": {"d7d6", "e7e6", "b8c6", "g7g6"},

	// Queen's Gambit (1. d4 d5 2. c4)
	"rnbqkbnr/ppp1pppp/8/3p4/3P4/8/PPP1PPPP/RNBQKBNR w KQkq d6 0 2": {"c2c4", "g1f3", "c1f4"},
	"rnbqkbnr/ppp1pppp/8/3p4/2PP4/8/PP2PPPP/RNBQKBNR b KQkq c3 0 2": {"e7e6", "c7c6", "d5c4"},
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetBookMove checks if the current position is in the opening book.
// Returns move.NullMove if not found.
func GetBookMove(b *board.Board) move.Move {
	fen := b.FEN()
	moves, exists := defaultBook[fen]
	if !exists || len(moves) == 0 {
		return move.NullMove
	}

	// Pick a random GM move from the opening book candidates
	chosenStr := moves[rand.Intn(len(moves))]

	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)

	for i := 0; i < list.Count; i++ {
		m := list.Moves[i]
		if m.String() == chosenStr {
			return m
		}
	}

	return move.NullMove
}
