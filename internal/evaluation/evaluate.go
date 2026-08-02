package evaluation

import (
	"hyperion/internal/board"
)

// Evaluate returns a tapered evaluation of the board from the perspective of the side to move.
func Evaluate(b *board.Board) int {
	whiteMg := 0
	whiteEg := 0
	blackMg := 0
	blackEg := 0
	gamePhase := 0

	for sq := 0; sq < 64; sq++ {
		p := b.Squares[sq]
		if p == board.NoPiece {
			continue
		}

		color := p.Color()
		pt := p.Type()

		// Accumulate game phase
		switch pt {
		case board.Knight:
			gamePhase += KnightPhase
		case board.Bishop:
			gamePhase += BishopPhase
		case board.Rook:
			gamePhase += RookPhase
		case board.Queen:
			gamePhase += QueenPhase
		}

		if color == board.White {
			sqIdx := sq ^ 56
			whiteMg += mgValue[pt] + mgPst[pt][sqIdx]
			whiteEg += egValue[pt] + egPst[pt][sqIdx]
		} else {
			blackMg += mgValue[pt] + mgPst[pt][sq]
			blackEg += egValue[pt] + egPst[pt][sq]
		}
	}

	if gamePhase > TotalPhase {
		gamePhase = TotalPhase
	}

	mgScore := whiteMg - blackMg
	egScore := whiteEg - blackEg

	// Interpolate between middlegame and endgame scores
	mgWeight := gamePhase
	egWeight := TotalPhase - gamePhase

	eval := (mgScore*mgWeight+egScore*egWeight)/TotalPhase + EvaluatePositional(b)

	if b.SideToMove == board.Black {
		eval = -eval
	}

	return eval
}
