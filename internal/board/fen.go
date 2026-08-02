package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// SetFEN sets the board state from a FEN string.
func (b *Board) SetFEN(fen string) error {
	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return errors.New("invalid FEN string: not enough fields")
	}

	// 1. Piece placement
	for i := 0; i < 64; i++ {
		b.RemovePiece(Square(i))
	}

	rank := Rank8
	file := FileA
	for _, char := range parts[0] {
		if char == '/' {
			rank--
			file = FileA
			continue
		}

		if char >= '1' && char <= '8' {
			file += File(char - '0')
			continue
		}

		var p Piece
		switch char {
		case 'P':
			p = WhitePawn
		case 'N':
			p = WhiteKnight
		case 'B':
			p = WhiteBishop
		case 'R':
			p = WhiteRook
		case 'Q':
			p = WhiteQueen
		case 'K':
			p = WhiteKing
		case 'p':
			p = BlackPawn
		case 'n':
			p = BlackKnight
		case 'b':
			p = BlackBishop
		case 'r':
			p = BlackRook
		case 'q':
			p = BlackQueen
		case 'k':
			p = BlackKing
		default:
			return fmt.Errorf("invalid piece character in FEN: %c", char)
		}

		b.AddPiece(NewSquare(file, rank), p)
		file++
	}

	// 2. Active color
	switch parts[1] {
	case "w":
		b.SideToMove = White
	case "b":
		b.SideToMove = Black
	default:
		return fmt.Errorf("invalid active color in FEN: %s", parts[1])
	}

	// 3. Castling availability
	b.Castle = NoCastling
	if parts[2] != "-" {
		for _, char := range parts[2] {
			switch char {
			case 'K':
				b.Castle |= WhiteOO
			case 'Q':
				b.Castle |= WhiteOOO
			case 'k':
				b.Castle |= BlackOO
			case 'q':
				b.Castle |= BlackOOO
			default:
				return fmt.Errorf("invalid castling rights in FEN: %c", char)
			}
		}
	}

	// 4. En passant target square
	b.EnPassant = NoSquare
	if parts[3] != "-" {
		if len(parts[3]) != 2 {
			return fmt.Errorf("invalid en passant square in FEN: %s", parts[3])
		}
		f := File(parts[3][0] - 'a')
		r := Rank(parts[3][1] - '1')
		if f >= FileA && f <= FileH && r >= Rank1 && r <= Rank8 {
			b.EnPassant = NewSquare(f, r)
		} else {
			return fmt.Errorf("invalid en passant square in FEN: %s", parts[3])
		}
	}

	// 5. Halfmove clock
	b.HalfMove = 0
	if len(parts) >= 5 {
		hm, err := strconv.Atoi(parts[4])
		if err != nil {
			return fmt.Errorf("invalid halfmove clock in FEN: %s", parts[4])
		}
		b.HalfMove = hm
	}

	// 6. Fullmove number
	b.FullMove = 1
	if len(parts) >= 6 {
		fm, err := strconv.Atoi(parts[5])
		if err != nil {
			return fmt.Errorf("invalid fullmove number in FEN: %s", parts[5])
		}
		b.FullMove = fm
	}

	b.Hash = b.ComputeHash()
	return nil
}

// FEN returns the FEN string representation of the board state.
func (b *Board) FEN() string {
	var sb strings.Builder

	// 1. Piece placement
	for r := 7; r >= 0; r-- {
		emptyCount := 0
		for f := 0; f <= 7; f++ {
			p := b.PieceAt(NewSquare(File(f), Rank(r)))
			if p == NoPiece {
				emptyCount++
			} else {
				if emptyCount > 0 {
					sb.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				sb.WriteByte(pieceChar(p))
			}
		}
		if emptyCount > 0 {
			sb.WriteString(strconv.Itoa(emptyCount))
		}
		if r > 0 {
			sb.WriteByte('/')
		}
	}

	sb.WriteByte(' ')

	// 2. Active color
	if b.SideToMove == White {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('b')
	}

	sb.WriteByte(' ')

	// 3. Castling availability
	if b.Castle == NoCastling {
		sb.WriteByte('-')
	} else {
		if b.Castle&WhiteOO != 0 {
			sb.WriteByte('K')
		}
		if b.Castle&WhiteOOO != 0 {
			sb.WriteByte('Q')
		}
		if b.Castle&BlackOO != 0 {
			sb.WriteByte('k')
		}
		if b.Castle&BlackOOO != 0 {
			sb.WriteByte('q')
		}
	}

	sb.WriteByte(' ')

	// 4. En passant target square
	if b.EnPassant == NoSquare {
		sb.WriteByte('-')
	} else {
		sb.WriteByte(byte('a' + b.EnPassant.File()))
		sb.WriteByte(byte('1' + b.EnPassant.Rank()))
	}

	// 5. Halfmove clock
	sb.WriteString(fmt.Sprintf(" %d ", b.HalfMove))

	// 6. Fullmove number
	sb.WriteString(fmt.Sprintf("%d", b.FullMove))

	return sb.String()
}

func pieceChar(p Piece) byte {
	switch p {
	case WhitePawn:
		return 'P'
	case WhiteKnight:
		return 'N'
	case WhiteBishop:
		return 'B'
	case WhiteRook:
		return 'R'
	case WhiteQueen:
		return 'Q'
	case WhiteKing:
		return 'K'
	case BlackPawn:
		return 'p'
	case BlackKnight:
		return 'n'
	case BlackBishop:
		return 'b'
	case BlackRook:
		return 'r'
	case BlackQueen:
		return 'q'
	case BlackKing:
		return 'k'
	default:
		return '?'
	}
}
