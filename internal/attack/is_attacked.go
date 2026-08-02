package attack

import (
	"hyperion/internal/board"
	"hyperion/internal/magic"
)

// IsSquareAttacked returns true if the given square is attacked by any piece of the specified color.
func IsSquareAttacked(b *board.Board, sq board.Square, byColor board.Color) bool {
	// 1. Pawns
	// Note: We use the attack pattern of a pawn of the OPPOSITE color to see if a pawn of the BYCOLOR is attacking the square.
	// E.g., to see if white pawns attack E4, we place a black pawn on E4 and see if it hits any white pawns.
	if PawnAttacks[byColor.Opposite()][sq]&b.Pieces[board.Pawn]&b.Colors[byColor] != 0 {
		return true
	}

	// 2. Knights
	if KnightAttacks[sq]&b.Pieces[board.Knight]&b.Colors[byColor] != 0 {
		return true
	}

	// 3. Kings
	if KingAttacks[sq]&b.Pieces[board.King]&b.Colors[byColor] != 0 {
		return true
	}

	occ := b.AllPieces()

	// 4. Bishops and Queens
	if magic.GetBishopAttacks(int(sq), occ)&(b.Pieces[board.Bishop]|b.Pieces[board.Queen])&b.Colors[byColor] != 0 {
		return true
	}

	// 5. Rooks and Queens
	if magic.GetRookAttacks(int(sq), occ)&(b.Pieces[board.Rook]|b.Pieces[board.Queen])&b.Colors[byColor] != 0 {
		return true
	}

	return false
}
