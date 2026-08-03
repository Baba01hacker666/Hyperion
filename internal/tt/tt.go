package tt

import (
	"hyperion/internal/move"
	"sync/atomic"
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

type ttEntry struct {
	key  atomic.Uint64
	data atomic.Uint64
}

type Table struct {
	entries []ttEntry
	mask    uint64
}

func packData(score int16, depth uint8, flag Flag, m move.Move) uint64 {
	return uint64(uint16(score)) |
		(uint64(depth) << 16) |
		(uint64(flag) << 24) |
		(uint64(m) << 32)
}

func unpackData(data uint64) (score int16, depth uint8, flag Flag, m move.Move) {
	score = int16(uint16(data & 0xFFFF))
	depth = uint8((data >> 16) & 0xFF)
	flag = Flag((data >> 24) & 0xFF)
	m = move.Move((data >> 32) & 0xFFFF)
	return
}

// NewTable creates a transposition table of the specified size in MB.
func NewTable(sizeMB int) *Table {
	entrySize := int(unsafe.Sizeof(ttEntry{}))
	numEntries := (sizeMB * 1024 * 1024) / entrySize

	// Round down to nearest power of 2 for fast bitwise masking
	size := 1
	for size*2 <= numEntries {
		size *= 2
	}

	return &Table{
		entries: make([]ttEntry, size),
		mask:    uint64(size - 1),
	}
}

func (t *Table) Store(hash uint64, depth int, flag Flag, score int, m move.Move) {
	idx := hash & t.mask
	entry := &t.entries[idx]

	currentKey := entry.key.Load()
	currentData := entry.data.Load()
	_, currentDepth, _, _ := unpackData(currentData)

	if currentKey == 0 || currentKey != hash || depth >= int(currentDepth) {
		data := packData(int16(score), uint8(depth), flag, m)
		entry.data.Store(data)
		entry.key.Store(hash)
	}
}

func (t *Table) Probe(hash uint64) (Entry, bool) {
	idx := hash & t.mask
	entry := &t.entries[idx]

	key := entry.key.Load()
	if key != hash {
		return Entry{}, false
	}
	data := entry.data.Load()
	if entry.key.Load() != hash {
		return Entry{}, false
	}

	score, depth, flag, m := unpackData(data)
	return Entry{
		Hash:  hash,
		Score: score,
		Depth: depth,
		Flag:  flag,
		Move:  m,
	}, true
}

func (t *Table) Clear() {
	for i := range t.entries {
		t.entries[i].key.Store(0)
		t.entries[i].data.Store(0)
	}
}
