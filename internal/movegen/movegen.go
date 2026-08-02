package movegen

import (
	"hyperion/internal/attack"
	"hyperion/internal/bitboard"
	"hyperion/internal/board"
	"hyperion/internal/magic"
	"hyperion/internal/move"
)

// GeneratePseudoLegalMoves generates all pseudo-legal moves and appends them to the list.
func GeneratePseudoLegalMoves(b *board.Board, list *MoveList) {
	us := b.SideToMove
	them := us.Opposite()

	ourPieces := b.Colors[us]
	theirPieces := b.Colors[them]
	occ := b.AllPieces()
	empty := ^occ

	// 1. Pawns
	pawns := b.Pieces[board.Pawn] & ourPieces
	var up, up2, capLeft, capRight bitboard.Bitboard
	var promoRank board.Rank
	var capLeftOffset, capRightOffset int

	if us == board.White {
		up = (pawns << 8) & empty
		up2 = ((up & bitboard.Rank3) << 8) & empty // Rank 3
		capLeft = (pawns << 7) & theirPieces & ^bitboard.FileH
		capRight = (pawns << 9) & theirPieces & ^bitboard.FileA
		promoRank = board.Rank8
		capLeftOffset = 7
		capRightOffset = 9
	} else {
		up = (pawns >> 8) & empty
		up2 = ((up & bitboard.Rank6) >> 8) & empty // Rank 6
		capLeft = (pawns >> 9) & theirPieces & ^bitboard.FileH
		capRight = (pawns >> 7) & theirPieces & ^bitboard.FileA
		promoRank = board.Rank1
		capLeftOffset = 9
		capRightOffset = 7
	}

	addPawnMoves(list, up, 8, us, promoRank, move.Quiet)
	addPawnMoves(list, up2, 16, us, promoRank, move.DoublePawnPush)
	addPawnMoves(list, capLeft, capLeftOffset, us, promoRank, move.Capture)
	addPawnMoves(list, capRight, capRightOffset, us, promoRank, move.Capture)

	// En Passant
	if b.EnPassant != board.NoSquare {
		epBB := bitboard.Empty
		epBB.Set(int(b.EnPassant))

		var epCapLeft, epCapRight bitboard.Bitboard
		if us == board.White {
			epCapLeft = (pawns << 7) & epBB & ^bitboard.FileH
			epCapRight = (pawns << 9) & epBB & ^bitboard.FileA
		} else {
			epCapLeft = (pawns >> 9) & epBB & ^bitboard.FileH
			epCapRight = (pawns >> 7) & epBB & ^bitboard.FileA
		}

		if epCapLeft != 0 {
			to := int(b.EnPassant)
			from := to - 7
			if us == board.Black {
				from = to + 9
			}
			list.Add(move.New(from, to, move.EPCapture))
		}
		if epCapRight != 0 {
			to := int(b.EnPassant)
			from := to - 9
			if us == board.Black {
				from = to + 7
			}
			list.Add(move.New(from, to, move.EPCapture))
		}
	}

	// 2. Knights
	knights := b.Pieces[board.Knight] & ourPieces
	for knights != 0 {
		from := knights.PopLSB()
		attacks := attack.KnightAttacks[from] & ^ourPieces
		addPieceMoves(list, from, attacks, theirPieces)
	}

	// 3. Bishops
	bishops := b.Pieces[board.Bishop] & ourPieces
	for bishops != 0 {
		from := bishops.PopLSB()
		attacks := magic.GetBishopAttacks(from, occ) & ^ourPieces
		addPieceMoves(list, from, attacks, theirPieces)
	}

	// 4. Rooks
	rooks := b.Pieces[board.Rook] & ourPieces
	for rooks != 0 {
		from := rooks.PopLSB()
		attacks := magic.GetRookAttacks(from, occ) & ^ourPieces
		addPieceMoves(list, from, attacks, theirPieces)
	}

	// 5. Queens
	queens := b.Pieces[board.Queen] & ourPieces
	for queens != 0 {
		from := queens.PopLSB()
		attacks := magic.GetQueenAttacks(from, occ) & ^ourPieces
		addPieceMoves(list, from, attacks, theirPieces)
	}

	// 6. Kings
	kings := b.Pieces[board.King] & ourPieces
	if kings != 0 {
		from := kings.PopLSB()
		attacks := attack.KingAttacks[from] & ^ourPieces
		addPieceMoves(list, from, attacks, theirPieces)

		// Castling
		if us == board.White {
			if b.Castle&board.WhiteOO != 0 && empty.Has(int(board.F1)) && empty.Has(int(board.G1)) {
				list.Add(move.New(int(board.E1), int(board.G1), move.KingCastle))
			}
			if b.Castle&board.WhiteOOO != 0 && empty.Has(int(board.D1)) && empty.Has(int(board.C1)) && empty.Has(int(board.B1)) {
				list.Add(move.New(int(board.E1), int(board.C1), move.QueenCastle))
			}
		} else {
			if b.Castle&board.BlackOO != 0 && empty.Has(int(board.F8)) && empty.Has(int(board.G8)) {
				list.Add(move.New(int(board.E8), int(board.G8), move.KingCastle))
			}
			if b.Castle&board.BlackOOO != 0 && empty.Has(int(board.D8)) && empty.Has(int(board.C8)) && empty.Has(int(board.B8)) {
				list.Add(move.New(int(board.E8), int(board.C8), move.QueenCastle))
			}
		}
	}
}

func addPieceMoves(list *MoveList, from int, attacks bitboard.Bitboard, enemies bitboard.Bitboard) {
	for attacks != 0 {
		to := attacks.PopLSB()
		flag := move.Quiet
		if enemies.Has(to) {
			flag = move.Capture
		}
		list.Add(move.New(from, to, flag))
	}
}

func addPawnMoves(list *MoveList, bb bitboard.Bitboard, offset int, us board.Color, promoRank board.Rank, baseFlag move.Flag) {
	for bb != 0 {
		to := bb.PopLSB()
		from := to - offset
		if us == board.Black {
			from = to + offset
		}

		if board.Square(to).Rank() == promoRank {
			if baseFlag == move.Capture {
				list.Add(move.New(from, to, move.QueenPromoCap))
				list.Add(move.New(from, to, move.RookPromoCap))
				list.Add(move.New(from, to, move.BishopPromoCap))
				list.Add(move.New(from, to, move.KnightPromoCap))
			} else {
				list.Add(move.New(from, to, move.QueenPromotion))
				list.Add(move.New(from, to, move.RookPromotion))
				list.Add(move.New(from, to, move.BishopPromotion))
				list.Add(move.New(from, to, move.KnightPromotion))
			}
		} else {
			list.Add(move.New(from, to, baseFlag))
		}
	}
}

// GenerateLegalMoves generates all strictly legal moves.
func GenerateLegalMoves(b *board.Board, list *MoveList) {
	pseudo := &MoveList{}
	GeneratePseudoLegalMoves(b, pseudo)

	us := b.SideToMove
	them := us.Opposite()
	var undo board.Undo

	for i := 0; i < pseudo.Count; i++ {
		m := pseudo.Moves[i]

		// Additional castling checks (cannot castle out of, through, or into check)
		flag := m.Flag()
		if flag == move.KingCastle || flag == move.QueenCastle {
			kingSq := board.E1
			if us == board.Black {
				kingSq = board.E8
			}

			if attack.IsSquareAttacked(b, kingSq, them) {
				continue // Can't castle out of check
			}

			transitSq := kingSq + 1
			if flag == move.QueenCastle {
				transitSq = kingSq - 1
			}

			if attack.IsSquareAttacked(b, transitSq, them) {
				continue // Can't castle through check
			}
		}

		b.MakeMove(m, &undo)

		kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
		inCheck := attack.IsSquareAttacked(b, kingPos, them)

		b.UnmakeMove(&undo)

		if !inCheck {
			list.Add(m)
		}
	}
}
