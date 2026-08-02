package board

import (
	"hyperion/internal/move"
	"hyperion/internal/zobrist"
)

// Undo holds the state necessary to unmake a move.
type Undo struct {
	Move      move.Move
	Captured  Piece
	Castle    CastlingRight
	EnPassant Square
	HalfMove  int
	Hash      uint64
}

// MakeMove applies a move to the board. It does not check for legality.
func (b *Board) MakeMove(m move.Move, undo *Undo) {
	from := Square(m.From())
	to := Square(m.To())
	flag := m.Flag()
	us := b.SideToMove
	them := us.Opposite()

	p := b.Squares[from]
	captured := b.Squares[to]

	// Save undo state
	undo.Move = m
	undo.Captured = captured
	undo.Castle = b.Castle
	undo.EnPassant = b.EnPassant
	undo.HalfMove = b.HalfMove
	undo.Hash = b.Hash

	// Incremental Zobrist Hash updates
	b.Hash ^= zobrist.SideToMove
	b.Hash ^= zobrist.PieceSquare[us][p.Type()][from]

	if captured != NoPiece && flag != move.EPCapture {
		b.Hash ^= zobrist.PieceSquare[them][captured.Type()][to]
	}

	// Update halfmove clock
	b.HalfMove++
	if p.Type() == Pawn || captured != NoPiece {
		b.HalfMove = 0
	}

	// Move the piece
	b.RemovePiece(from)
	if captured != NoPiece && flag != move.EPCapture {
		b.RemovePiece(to)
	}

	// Handle promotions
	finalPiece := p
	if m.IsPromotion() {
		finalPiece = NewPiece(us, PieceType(m.PromotionType()))
	}
	b.Hash ^= zobrist.PieceSquare[us][finalPiece.Type()][to]

	b.AddPiece(to, finalPiece)

	// Handle En Passant capture
	if flag == move.EPCapture {
		epPawnSq := to
		if us == White {
			epPawnSq -= 8
		} else {
			epPawnSq += 8
		}
		undo.Captured = b.Squares[epPawnSq]
		b.Hash ^= zobrist.PieceSquare[them][Pawn][epPawnSq]
		b.RemovePiece(epPawnSq)
	}

	// Hash out old En Passant
	if undo.EnPassant != NoSquare {
		b.Hash ^= zobrist.EnPassant[undo.EnPassant]
	}

	// Reset En Passant
	b.EnPassant = NoSquare

	// Handle Double Pawn Push (set En Passant)
	if flag == move.DoublePawnPush {
		if us == White {
			b.EnPassant = to - 8
		} else {
			b.EnPassant = to + 8
		}
		b.Hash ^= zobrist.EnPassant[b.EnPassant]
	}

	// Handle Castling (move the rook)
	if flag == move.KingCastle {
		if us == White {
			b.RemovePiece(H1)
			b.AddPiece(F1, WhiteRook)
			b.Hash ^= zobrist.PieceSquare[White][Rook][H1] ^ zobrist.PieceSquare[White][Rook][F1]
		} else {
			b.RemovePiece(H8)
			b.AddPiece(F8, BlackRook)
			b.Hash ^= zobrist.PieceSquare[Black][Rook][H8] ^ zobrist.PieceSquare[Black][Rook][F8]
		}
	} else if flag == move.QueenCastle {
		if us == White {
			b.RemovePiece(A1)
			b.AddPiece(D1, WhiteRook)
			b.Hash ^= zobrist.PieceSquare[White][Rook][A1] ^ zobrist.PieceSquare[White][Rook][D1]
		} else {
			b.RemovePiece(A8)
			b.AddPiece(D8, BlackRook)
			b.Hash ^= zobrist.PieceSquare[Black][Rook][A8] ^ zobrist.PieceSquare[Black][Rook][D8]
		}
	}

	// Update castling rights
	b.Hash ^= zobrist.Castle[b.Castle]
	b.Castle &= castlingMask[from] & castlingMask[to]
	b.Hash ^= zobrist.Castle[b.Castle]

	// Update fullmove and turn
	if b.SideToMove == Black {
		b.FullMove++
	}
	b.SideToMove = them
}

// UnmakeMove reverts a move on the board using the Undo struct.
func (b *Board) UnmakeMove(undo *Undo) {
	m := undo.Move
	us := b.SideToMove.Opposite()
	from := Square(m.From())
	to := Square(m.To())
	flag := m.Flag()

	// Restore state
	b.SideToMove = us
	if b.SideToMove == Black {
		b.FullMove--
	}
	b.Castle = undo.Castle
	b.EnPassant = undo.EnPassant
	b.HalfMove = undo.HalfMove

	// Move piece back
	p := b.Squares[to]
	if m.IsPromotion() {
		p = NewPiece(us, Pawn)
	}

	b.RemovePiece(to)
	b.AddPiece(from, p)

	// Restore captured piece
	if flag == move.EPCapture {
		epPawnSq := to
		if us == White {
			epPawnSq -= 8
		} else {
			epPawnSq += 8
		}
		b.AddPiece(epPawnSq, undo.Captured)
	} else if undo.Captured != NoPiece {
		b.AddPiece(to, undo.Captured)
	}

	// Restore rook for castling
	if flag == move.KingCastle {
		if us == White {
			b.RemovePiece(F1)
			b.AddPiece(H1, WhiteRook)
		} else {
			b.RemovePiece(F8)
			b.AddPiece(H8, BlackRook)
		}
	} else if flag == move.QueenCastle {
		if us == White {
			b.RemovePiece(D1)
			b.AddPiece(A1, WhiteRook)
		} else {
			b.RemovePiece(D8)
			b.AddPiece(A8, BlackRook)
		}
	}

	// Restore exact hash in 1 cycle
	b.Hash = undo.Hash
}

// castlingMask removes castling rights when a king or rook moves or is captured.
var castlingMask [64]CastlingRight

func init() {
	for i := 0; i < 64; i++ {
		castlingMask[i] = AnyCastling
	}
	castlingMask[A1] &^= WhiteOOO
	castlingMask[H1] &^= WhiteOO
	castlingMask[E1] &^= (WhiteOO | WhiteOOO)
	castlingMask[A8] &^= BlackOOO
	castlingMask[H8] &^= BlackOO
	castlingMask[E8] &^= (BlackOO | BlackOOO)
}
