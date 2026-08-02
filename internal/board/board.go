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
	FullMove   int                  // Fullmove number
	Hash       uint64               // Zobrist hash key
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
