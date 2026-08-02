package board

// Color represents the side to move or the color of a piece.
type Color uint8

const (
	White Color = 0
	Black Color = 1
	Both  Color = 2
)

// Opposite returns the opposite color.
func (c Color) Opposite() Color {
	return c ^ 1
}

// PieceType represents the type of a piece (Pawn, Knight, etc.).
type PieceType uint8

const (
	Pawn PieceType = iota
	Knight
	Bishop
	Rook
	Queen
	King
	NoPieceType
)

// Piece represents a specific colored piece (WhitePawn, BlackKnight, etc.).
type Piece uint8

const (
	WhitePawn Piece = iota
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
	NoPiece
)

// NewPiece creates a Piece from a Color and a PieceType.
func NewPiece(color Color, pt PieceType) Piece {
	if pt == NoPieceType {
		return NoPiece
	}
	return Piece(int(color)*6 + int(pt))
}

// Type returns the PieceType of the piece.
func (p Piece) Type() PieceType {
	if p == NoPiece {
		return NoPieceType
	}
	return PieceType(p % 6)
}

// Color returns the Color of the piece.
func (p Piece) Color() Color {
	if p == NoPiece {
		return Both
	}
	return Color(p / 6)
}

// File represents a column on the chessboard.
type File uint8

const (
	FileA File = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

// Rank represents a row on the chessboard.
type Rank uint8

const (
	Rank1 Rank = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

// Square represents a square on the chessboard (0-63).
type Square uint8

const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	NoSquare
)

// NewSquare returns a Square from a File and a Rank.
func NewSquare(file File, rank Rank) Square {
	return Square(uint8(rank)<<3 | uint8(file))
}

// File returns the File of the square.
func (s Square) File() File {
	return File(s & 7)
}

// Rank returns the Rank of the square.
func (s Square) Rank() Rank {
	return Rank(s >> 3)
}
