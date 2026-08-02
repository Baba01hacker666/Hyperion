package evaluation

import (
	"hyperion/internal/attack"
	"hyperion/internal/bitboard"
	"hyperion/internal/board"
	"hyperion/internal/magic"
)

// EvaluatePositional Terms adds pawn structure, king safety, and piece mobility bonuses.
func EvaluatePositional(b *board.Board) int {
	score := 0

	// 1. Pawn Structure & Passed Pawns
	score += evaluatePawns(b, board.White) - evaluatePawns(b, board.Black)

	// 2. Piece Mobility & Outposts
	score += evaluateMobility(b, board.White) - evaluateMobility(b, board.Black)

	// 3. King Safety & Pawn Shields
	score += evaluateKingSafety(b, board.White) - evaluateKingSafety(b, board.Black)

	return score
}

func evaluatePawns(b *board.Board, color board.Color) int {
	score := 0
	pawns := b.Pieces[board.Pawn] & b.Colors[color]
	enemyPawns := b.Pieces[board.Pawn] & b.Colors[color.Opposite()]

	for pawns != 0 {
		sq := board.Square(pawns.PopLSB())
		file := sq.File()
		rank := sq.Rank()

		// Passed Pawn Bonus
		passedMask := getPassedPawnMask(color, sq)
		if (passedMask & enemyPawns) == 0 {
			var r int
			if color == board.White {
				r = int(rank)
			} else {
				r = 7 - int(rank)
			}
			// Passed pawn bonus scales quadratically with rank advance
			score += (r * r * 5)
		}

		// Doubled Pawn Penalty
		fileMask := bitboard.FileMasks[file]
		if (pawns & fileMask) != 0 {
			score -= 15
		}

		// Isolated Pawn Penalty
		adjFiles := bitboard.Bitboard(0)
		if file > 0 {
			adjFiles |= bitboard.FileMasks[file-1]
		}
		if file < 7 {
			adjFiles |= bitboard.FileMasks[file+1]
		}
		if (b.Pieces[board.Pawn]&b.Colors[color]&adjFiles) == 0 {
			score -= 12
		}
	}

	return score
}

func getPassedPawnMask(color board.Color, sq board.Square) bitboard.Bitboard {
	file := int(sq.File())
	rank := int(sq.Rank())
	mask := bitboard.Bitboard(0)

	files := bitboard.FileMasks[file]
	if file > 0 {
		files |= bitboard.FileMasks[file-1]
	}
	if file < 7 {
		files |= bitboard.FileMasks[file+1]
	}

	if color == board.White {
		for r := rank + 1; r <= 7; r++ {
			mask |= (files & bitboard.RankMasks[r])
		}
	} else {
		for r := rank - 1; r >= 0; r-- {
			mask |= (files & bitboard.RankMasks[r])
		}
	}

	return mask
}

func evaluateMobility(b *board.Board, color board.Color) int {
	score := 0
	occ := b.AllPieces()

	// Rook on Open/Semi-Open File
	rooks := b.Pieces[board.Rook] & b.Colors[color]
	friendlyPawns := b.Pieces[board.Pawn] & b.Colors[color]
	enemyPawns := b.Pieces[board.Pawn] & b.Colors[color.Opposite()]

	for rooks != 0 {
		sq := board.Square(rooks.PopLSB())
		file := sq.File()
		fileMask := bitboard.FileMasks[file]

		if (friendlyPawns & fileMask) == 0 {
			if (enemyPawns & fileMask) == 0 {
				score += 25 // Fully open file
			} else {
				score += 12 // Semi-open file
			}
		}

		// Mobility count
		attacks := magic.GetRookAttacks(int(sq), occ)
		score += attacks.PopCount() * 2
	}

	// Bishop Mobility
	bishops := b.Pieces[board.Bishop] & b.Colors[color]
	for bishops != 0 {
		sq := board.Square(bishops.PopLSB())
		attacks := magic.GetBishopAttacks(int(sq), occ)
		score += attacks.PopCount() * 3
	}

	// Queen Mobility
	queens := b.Pieces[board.Queen] & b.Colors[color]
	for queens != 0 {
		sq := board.Square(queens.PopLSB())
		attacks := magic.GetQueenAttacks(int(sq), occ)
		score += attacks.PopCount() * 1
	}

	return score
}

func evaluateKingSafety(b *board.Board, color board.Color) int {
	score := 0
	kingSq := board.Square((b.Colors[color] & b.Pieces[board.King]).LSB())
	friendlyPawns := b.Pieces[board.Pawn] & b.Colors[color]

	// Pawn shield around King
	file := kingSq.File()
	rank := kingSq.Rank()

	if (color == board.White && rank <= 1) || (color == board.Black && rank >= 6) {
		for f := max(0, int(file)-1); f <= min(7, int(file)+1); f++ {
			shieldFile := bitboard.FileMasks[f]
			if (friendlyPawns & shieldFile) != 0 {
				score += 15
			} else {
				score -= 20 // Missing pawn shield penalty
			}
		}
	}

	return score
}

func getKingZone(sq board.Square) bitboard.Bitboard {
	return attack.KingAttacks[sq]
}

func getFileMask(f int) bitboard.Bitboard {
	return bitboard.FileMasks[f]
}
