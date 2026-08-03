package evaluation

import (
	"hyperion/internal/attack"
	"hyperion/internal/bitboard"
	"hyperion/internal/board"
	"hyperion/internal/magic"
)

// EvaluatePositional Terms adds pawn structure, king safety, piece mobility, bishop pair, and outpost bonuses.
func EvaluatePositional(b *board.Board) int {
	score := 0

	// 1. Pawn Structure & Passed Pawns
	score += evaluatePawns(b, board.White) - evaluatePawns(b, board.Black)

	// 2. Piece Mobility
	score += evaluateMobility(b, board.White) - evaluateMobility(b, board.Black)

	// 3. King Safety & Pawn Shields
	score += evaluateKingSafety(b, board.White) - evaluateKingSafety(b, board.Black)

	// 4. Bishop Pair Bonus
	whiteBishops := (b.Pieces[board.Bishop] & b.Colors[board.White]).PopCount()
	blackBishops := (b.Pieces[board.Bishop] & b.Colors[board.Black]).PopCount()
	if whiteBishops >= 2 {
		score += 35
	}
	if blackBishops >= 2 {
		score -= 35
	}

	// 5. Knight Outposts
	score += evaluateKnights(b, board.White) - evaluateKnights(b, board.Black)

	return score
}

func evaluateKnights(b *board.Board, color board.Color) int {
	score := 0
	knights := b.Pieces[board.Knight] & b.Colors[color]
	friendlyPawns := b.Pieces[board.Pawn] & b.Colors[color]
	enemyPawns := b.Pieces[board.Pawn] & b.Colors[color.Opposite()]

	for knights != 0 {
		sq := board.Square(knights.PopLSB())
		rank := sq.Rank()
		file := sq.File()

		// Outpost condition: Knight on ranks 4, 5, or 6
		isOutpostRank := false
		if color == board.White && rank >= 3 && rank <= 5 {
			isOutpostRank = true
		} else if color == board.Black && rank >= 2 && rank <= 4 {
			isOutpostRank = true
		}

		if isOutpostRank {
			// Protected by friendly pawn
			pawnAttacks := attack.PawnAttacks[color.Opposite()][sq]
			if (pawnAttacks & friendlyPawns) != 0 {
				// No enemy pawn can attack this square from adjacent files
				adjFiles := bitboard.Bitboard(0)
				if file > 0 {
					adjFiles |= bitboard.FileMasks[file-1]
				}
				if file < 7 {
					adjFiles |= bitboard.FileMasks[file+1]
				}
				if (enemyPawns & adjFiles) == 0 {
					score += 25 // Strong Knight Outpost!
				}
			}
		}
	}

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
			score += (r * r * 8)
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
	opp := color.Opposite()
	kingSq := board.Square((b.Colors[color] & b.Pieces[board.King]).LSB())
	friendlyPawns := b.Pieces[board.Pawn] & b.Colors[color]
	enemyHeavy := (b.Pieces[board.Queen] | b.Pieces[board.Rook]) & b.Colors[opp]

	file := kingSq.File()
	rank := kingSq.Rank()

	// Pawn shield around King
	if (color == board.White && rank <= 1) || (color == board.Black && rank >= 6) {
		for f := max(0, int(file)-1); f <= min(7, int(file)+1); f++ {
			shieldFile := bitboard.FileMasks[f]
			if (friendlyPawns & shieldFile) != 0 {
				score += 15
			} else {
				score -= 20 // Missing pawn shield penalty
			}
		}
	} else if enemyHeavy != 0 {
		// Exposed King in middle of board with enemy Queens/Rooks present
		var exposedDist int
		if color == board.White {
			exposedDist = int(rank)
		} else {
			exposedDist = 7 - int(rank)
		}
		if exposedDist >= 2 {
			score -= exposedDist * 30 // Heavy penalty for advancing king prematurely
		}
	}

	// King Ring Pressure
	kingZone := attack.KingAttacks[kingSq]
	attackersCount := 0

	for pType := board.Knight; pType <= board.Queen; pType++ {
		enemyPieces := b.Pieces[pType] & b.Colors[opp]
		for enemyPieces != 0 {
			eSq := board.Square(enemyPieces.PopLSB())
			var eAttacks bitboard.Bitboard
			switch pType {
			case board.Knight:
				eAttacks = attack.KnightAttacks[eSq]
			case board.Bishop:
				eAttacks = magic.GetBishopAttacks(int(eSq), b.AllPieces())
			case board.Rook:
				eAttacks = magic.GetRookAttacks(int(eSq), b.AllPieces())
			case board.Queen:
				eAttacks = magic.GetQueenAttacks(int(eSq), b.AllPieces())
			}
			if (eAttacks & kingZone) != 0 {
				attackersCount++
			}
		}
	}
	if attackersCount >= 2 {
		score -= attackersCount * attackersCount * 8 // Aggressive king attack penalty
	}

	return score
}

func getKingZone(sq board.Square) bitboard.Bitboard {
	return attack.KingAttacks[sq]
}

func getFileMask(f int) bitboard.Bitboard {
	return bitboard.FileMasks[f]
}
