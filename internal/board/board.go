package board

import (
	"hyperion/internal/bitboard"
)

// CastlingRight represents the castling rights using a bitmask.
type CastlingRight uint8

const (
	WhiteOO CastlingRight = 1 << iota
	WhiteOOO
	BlackOO
	BlackOOO
	NoCastling  CastlingRight = 0
	AnyCastling CastlingRight = WhiteOO | WhiteOOO | BlackOO | BlackOOO
)

// Board represents the state of a chess game.
type Board struct {
	Pieces     [6]bitboard.Bitboard // Bitboards for each PieceType (Pawn..King)
	Colors     [2]bitboard.Bitboard // Bitboards for each Color (White, Black)
	Squares    [64]Piece            // Array storing the Piece at each Square
	SideToMove Color                // White or Black
	Castle     CastlingRight        // Current castling rights
	EnPassant  Square               // En passant target square (NoSquare if none)
	HalfMove   int                  // Halfmove clock for fifty-move rule
	FullMove     int                  // Fullmove number
	Hash         uint64               // Zobrist hash key
	PosHistory   [256]uint64          // Array storing hash history for repetition checking
	HistoryCount int                  // Total recorded positions in history
}

// New creates and returns a new empty Board.
func New() *Board {
	b := &Board{
		SideToMove: White,
		Castle:     NoCastling,
		EnPassant:  NoSquare,
		HalfMove:   0,
		FullMove:   1,
	}
	for i := 0; i < 64; i++ {
		b.Squares[i] = NoPiece
	}
	return b
}

// AddPiece adds a Piece to the board at the specified Square.
func (b *Board) AddPiece(sq Square, p Piece) {
	if b.Squares[sq] != NoPiece {
		b.RemovePiece(sq)
	}
	b.Squares[sq] = p
	b.Pieces[p.Type()].Set(int(sq))
	b.Colors[p.Color()].Set(int(sq))
}

// RemovePiece removes and returns the Piece at the specified Square.
func (b *Board) RemovePiece(sq Square) Piece {
	p := b.Squares[sq]
	if p != NoPiece {
		b.Pieces[p.Type()].Clear(int(sq))
		b.Colors[p.Color()].Clear(int(sq))
		b.Squares[sq] = NoPiece
	}
	return p
}

// PieceAt returns the Piece at the specified Square.
func (b *Board) PieceAt(sq Square) Piece {
	return b.Squares[sq]
}

// AllPieces returns a Bitboard containing all pieces on the board.
func (b *Board) AllPieces() bitboard.Bitboard {
	return b.Colors[White] | b.Colors[Black]
}

// IsRepetition checks if the current board position has occurred at least twice before
// (i.e. we are about to make it a 3-fold repetition, which is a draw).
func (b *Board) IsRepetition() bool {
	if b.HistoryCount < 4 {
		return false
	}
	currHash := b.Hash
	count := 0
	// Only check positions of the same side to move (every 2 plies back)
	// and stop at irreversible moves (halfmove clock resets)
	for i := b.HistoryCount - 2; i >= 0; i -= 2 {
		if b.PosHistory[i] == currHash {
			count++
			if count >= 2 {
				return true
			}
		}
		// Stop searching back past an irreversible move
		if i < b.HistoryCount-int(b.HalfMove) {
			break
		}
	}
	return false
}

// Is2FoldRepetition returns true if current position has occurred at least once before
// (i.e., this is the 2nd occurrence = 2-fold repetition). Used in search to score as draw.
func (b *Board) Is2FoldRepetition() bool {
	if b.HistoryCount < 2 {
		return false
	}
	currHash := b.Hash
	for i := b.HistoryCount - 2; i >= 0; i -= 2 {
		if b.PosHistory[i] == currHash {
			return true
		}
		if i < b.HistoryCount-int(b.HalfMove) {
			break
		}
	}
	return false
}

// Clone creates a deep copy of the board state.
func (b *Board) Clone() *Board {
	cb := *b
	return &cb
}
