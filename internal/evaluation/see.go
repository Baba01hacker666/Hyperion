package evaluation

import (
	"hyperion/internal/attack"
	"hyperion/internal/bitboard"
	"hyperion/internal/board"
	"hyperion/internal/magic"
	"hyperion/internal/move"
)

var seeValues = [6]int{100, 320, 330, 500, 900, 20000}

// SEE (Static Exchange Evaluation) evaluates a series of captures on a target square.
// Returns true if the net gain for the attacker is >= threshold.
func SEE(b *board.Board, m move.Move, threshold int) bool {
	from := board.Square(m.From())
	to := board.Square(m.To())

	attackerPiece := b.PieceAt(from)
	if attackerPiece == board.NoPiece {
		return false
	}

	targetPiece := b.PieceAt(to)
	victimVal := 0
	if targetPiece != board.NoPiece {
		victimVal = seeValues[targetPiece.Type()]
	} else if m.Flag() == move.EPCapture {
		victimVal = seeValues[board.Pawn]
	}

	// Early cutoff if initial gain is already below threshold
	if victimVal-seeValues[attackerPiece.Type()] >= threshold {
		return true
	}

	// Swap list array
	var gain [32]int
	gain[0] = victimVal

	occ := b.AllPieces()
	occ.Clear(int(from))

	// Initial attacker
	currentAttackerVal := seeValues[attackerPiece.Type()]
	side := b.SideToMove.Opposite()
	depth := 1

	for {
		// Find least valuable attacker for current side
		attackerSq, attackerType, found := getLeastValuableAttacker(b, to, side, occ)
		if !found {
			break
		}

		gain[depth] = currentAttackerVal - gain[depth-1]
		if max(-gain[depth], gain[depth-1]) < threshold {
			break
		}

		currentAttackerVal = seeValues[attackerType]
		occ.Clear(int(attackerSq))
		side = side.Opposite()
		depth++
	}

	// Minimax back up the swap list
	for depth--; depth > 0; depth-- {
		gain[depth-1] = -max(-gain[depth], -gain[depth-1])
	}

	return gain[0] >= threshold
}

func getLeastValuableAttacker(b *board.Board, to board.Square, side board.Color, occ bitboard.Bitboard) (board.Square, board.PieceType, bool) {
	// 1. Pawns
	pawnAttacks := attack.PawnAttacks[side.Opposite()][to] & b.Pieces[board.Pawn] & b.Colors[side] & occ
	if pawnAttacks != 0 {
		sq := board.Square(pawnAttacks.LSB())
		return sq, board.Pawn, true
	}

	// 2. Knights
	knightAttacks := attack.KnightAttacks[to] & b.Pieces[board.Knight] & b.Colors[side] & occ
	if knightAttacks != 0 {
		sq := board.Square(knightAttacks.LSB())
		return sq, board.Knight, true
	}

	// 3. Bishops
	bishopAttacks := magic.GetBishopAttacks(int(to), occ) & b.Pieces[board.Bishop] & b.Colors[side] & occ
	if bishopAttacks != 0 {
		sq := board.Square(bishopAttacks.LSB())
		return sq, board.Bishop, true
	}

	// 4. Rooks
	rookAttacks := magic.GetRookAttacks(int(to), occ) & b.Pieces[board.Rook] & b.Colors[side] & occ
	if rookAttacks != 0 {
		sq := board.Square(rookAttacks.LSB())
		return sq, board.Rook, true
	}

	// 5. Queens
	queenAttacks := magic.GetQueenAttacks(int(to), occ) & b.Pieces[board.Queen] & b.Colors[side] & occ
	if queenAttacks != 0 {
		sq := board.Square(queenAttacks.LSB())
		return sq, board.Queen, true
	}

	// 6. Kings
	kingAttacks := attack.KingAttacks[to] & b.Pieces[board.King] & b.Colors[side] & occ
	if kingAttacks != 0 {
		sq := board.Square(kingAttacks.LSB())
		return sq, board.King, true
	}

	return board.NoSquare, board.NoPieceType, false
}
