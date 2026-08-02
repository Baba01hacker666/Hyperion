package evaluation

import (
	"hyperion/internal/board"
)

const (
	PawnValue   = 100
	KnightValue = 320
	BishopValue = 330
	RookValue   = 500
	QueenValue  = 900
)

var pieceValues = [6]int{PawnValue, KnightValue, BishopValue, RookValue, QueenValue, 0}

// Evaluate returns a static evaluation of the board from the perspective of the side to move.
// Positive values mean the side to move is winning.
func Evaluate(b *board.Board) int {
	whiteScore := 0
	blackScore := 0

	for sq := 0; sq < 64; sq++ {
		p := b.Squares[sq]
		if p == board.NoPiece {
			continue
		}

		color := p.Color()
		pt := p.Type()

		val := pieceValues[pt]

		// Add Piece-Square Table value
		if color == board.White {
			val += pstWhite[pt][sq^56]
			whiteScore += val
		} else {
			// For black, sq directly maps top-to-bottom (Rank 8 = index 56..63)
			val += pstWhite[pt][sq]
			blackScore += val
		}
	}

	eval := whiteScore - blackScore
	if b.SideToMove == board.Black {
		eval = -eval
	}

	return eval
}
