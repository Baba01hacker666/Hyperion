package tt

import (
	"hyperion/internal/move"
	"unsafe"
)

type Flag uint8

const (
	Exact Flag = iota
	LowerBound
	UpperBound
)

type Entry struct {
	Hash  uint64
	Score int16
	Depth uint8
	Flag  Flag
	Move  move.Move
}

type Table struct {
	entries []Entry
	mask    uint64
}

// NewTable creates a transposition table of the specified size in MB.
func NewTable(sizeMB int) *Table {
	entrySize := int(unsafe.Sizeof(Entry{}))
	numEntries := (sizeMB * 1024 * 1024) / entrySize

	// Round down to nearest power of 2 for fast bitwise masking
	size := 1
	for size*2 <= numEntries {
		size *= 2
	}

	return &Table{
		entries: make([]Entry, size),
		mask:    uint64(size - 1),
	}
}

func (t *Table) Store(hash uint64, depth int, flag Flag, score int, m move.Move) {
	idx := hash & t.mask
	entry := &t.entries[idx]

	// Always replace if new search depth is greater or equal
	if entry.Hash == 0 || depth >= int(entry.Depth) {
		entry.Hash = hash
		entry.Score = int16(score)
		entry.Depth = uint8(depth)
		entry.Flag = flag
		entry.Move = m
	}
}

func (t *Table) Probe(hash uint64) (Entry, bool) {
	idx := hash & t.mask
	entry := t.entries[idx]

	if entry.Hash == hash {
		return entry, true
	}
	return Entry{}, false
}

func (t *Table) Clear() {
	for i := range t.entries {
		t.entries[i] = Entry{}
	}
}
