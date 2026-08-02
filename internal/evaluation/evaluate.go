package evaluation

import (
	"hyperion/internal/board"
)

type Style int

const (
	StyleBalanced Style = iota
	StyleGamble
	StyleDefense
	StyleEvil
)

var CurrentStyle = StyleBalanced

// SetStyle configures the engine's play style.
func SetStyle(s Style) {
	CurrentStyle = s
}

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

	// Apply Play Style Modifiers
	if CurrentStyle == StyleGamble {
		eval += evaluateGambleBonus(b)
	} else if CurrentStyle == StyleDefense {
		eval += evaluateDefenseBonus(b)
	} else if CurrentStyle == StyleEvil {
		eval += evaluateEvilBonus(b)
	}

	if b.SideToMove == board.Black {
		eval = -eval
	}

	return eval
}

func evaluateGambleBonus(b *board.Board) int {
	score := 0
	us := b.SideToMove
	them := us.Opposite()

	enemyKingSq := board.Square((b.Colors[them] & b.Pieces[board.King]).LSB())
	kingZone := getKingZone(enemyKingSq)

	friendlyAttackers := (b.Colors[us] &^ b.Pieces[board.Pawn] &^ b.Pieces[board.King])
	if (friendlyAttackers & kingZone) != 0 {
		score += 80
	}

	return score
}

func evaluateDefenseBonus(b *board.Board) int {
	score := 0
	us := b.SideToMove
	kingSq := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())

	file := kingSq.File()
	rank := kingSq.Rank()
	friendlyPawns := b.Pieces[board.Pawn] & b.Colors[us]

	if (us == board.White && rank <= 1) || (us == board.Black && rank >= 6) {
		for f := max(0, int(file)-1); f <= min(7, int(file)+1); f++ {
			shieldFile := getFileMask(f)
			if (friendlyPawns & shieldFile) != 0 {
				score += 25
			}
		}
	}

	return score
}

func evaluateEvilBonus(b *board.Board) int {
	score := 0
	us := b.SideToMove
	them := us.Opposite()

	// Evil Mode: Maximum aggression, king constriction, and passed pawn dominance
	enemyKingSq := board.Square((b.Colors[them] & b.Pieces[board.King]).LSB())
	kingZone := getKingZone(enemyKingSq)

	// Heavily penalize enemy king safety
	friendlyAttackers := (b.Colors[us] &^ b.Pieces[board.King])
	if (friendlyAttackers & kingZone) != 0 {
		score += 140 // Ruthless king attack bonus!
	}

	return score
}
