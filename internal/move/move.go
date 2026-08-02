package move

import (
	"fmt"
)

// Flag represents the special characteristics of a move (4 bits).
type Flag uint16

const (
	Quiet           Flag = 0
	DoublePawnPush  Flag = 1
	KingCastle      Flag = 2
	QueenCastle     Flag = 3
	Capture         Flag = 4
	EPCapture       Flag = 5
	KnightPromotion Flag = 8
	BishopPromotion Flag = 9
	RookPromotion   Flag = 10
	QueenPromotion  Flag = 11
	KnightPromoCap  Flag = 12
	BishopPromoCap  Flag = 13
	RookPromoCap    Flag = 14
	QueenPromoCap   Flag = 15
)

// Move represents a chess move using 16 bits.
// Bits 0-5: From square (0-63)
// Bits 6-11: To square (0-63)
// Bits 12-15: Flag (0-15)
type Move uint16

const NullMove Move = 0

// New creates a new move from its components.
func New(from, to int, flag Flag) Move {
	return Move(uint16(from) | (uint16(to) << 6) | (uint16(flag) << 12))
}

// From returns the source square of the move (0-63).
func (m Move) From() int {
	return int(m & 0x3F)
}

// To returns the destination square of the move (0-63).
func (m Move) To() int {
	return int((m >> 6) & 0x3F)
}

// Flag returns the move flag.
func (m Move) Flag() Flag {
	return Flag(m >> 12)
}

// IsCapture returns true if the move is any kind of capture.
func (m Move) IsCapture() bool {
	return m.Flag() >= Capture && m.Flag() != KnightPromotion && m.Flag() != BishopPromotion && m.Flag() != RookPromotion && m.Flag() != QueenPromotion
}

// IsPromotion returns true if the move is a promotion.
func (m Move) IsPromotion() bool {
	return m.Flag() >= KnightPromotion
}

// PromotionType returns the piece type the pawn is promoting to (integer corresponding to board.PieceType).
// Returns 6 (NoPieceType) if not a promotion.
func (m Move) PromotionType() int {
	switch m.Flag() {
	case KnightPromotion, KnightPromoCap:
		return 1 // Knight
	case BishopPromotion, BishopPromoCap:
		return 2 // Bishop
	case RookPromotion, RookPromoCap:
		return 3 // Rook
	case QueenPromotion, QueenPromoCap:
		return 4 // Queen
	default:
		return 6 // NoPieceType
	}
}

// String returns the UCI notation of the move (e.g., "e2e4", "e7e8q").
func (m Move) String() string {
	if m == NullMove {
		return "0000"
	}
	from := m.From()
	to := m.To()

	fromFile := from & 7
	fromRank := from >> 3
	toFile := to & 7
	toRank := to >> 3

	fromStr := fmt.Sprintf("%c%c", 'a'+fromFile, '1'+fromRank)
	toStr := fmt.Sprintf("%c%c", 'a'+toFile, '1'+toRank)

	promoStr := ""
	if m.IsPromotion() {
		switch m.PromotionType() {
		case 1:
			promoStr = "n" // Knight
		case 2:
			promoStr = "b" // Bishop
		case 3:
			promoStr = "r" // Rook
		case 4:
			promoStr = "q" // Queen
		}
	}

	return fromStr + toStr + promoStr
}
