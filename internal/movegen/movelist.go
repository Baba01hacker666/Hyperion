package movegen

import (
	"hyperion/internal/move"
)

// MoveList represents a pre-allocated array of moves to avoid dynamic allocations during search.
type MoveList struct {
	Moves [256]move.Move
	Count int
}

// Add appends a move to the list.
func (l *MoveList) Add(m move.Move) {
	l.Moves[l.Count] = m
	l.Count++
}

// Clear resets the move list.
func (l *MoveList) Clear() {
	l.Count = 0
}

// Slice returns a slice containing only the valid moves in the list.
func (l *MoveList) Slice() []move.Move {
	return l.Moves[:l.Count]
}
