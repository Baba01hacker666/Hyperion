package search

import (
	"hyperion/internal/attack"
	"hyperion/internal/board"
	"hyperion/internal/evaluation"
	"hyperion/internal/move"
	"hyperion/internal/movegen"
	"hyperion/internal/tt"
	"sort"
)

const (
	Infinity = 30000
	MateVal  = 20000
	MaxDepth = 64
)

// Searcher holds state for search.
type Searcher struct {
	Nodes       uint64
	TT          *tt.Table
	KillerMoves [MaxDepth][2]move.Move
	History     [2][64][64]int
}

func NewSearcher(ttSizeMB int) *Searcher {
	return &Searcher{
		TT: tt.NewTable(ttSizeMB),
	}
}

// Search performs Iterative Deepening with Aspiration Windows up to maxDepth.
func (s *Searcher) Search(b *board.Board, maxDepth int) (move.Move, int) {
	s.Nodes = 0
	var overallBestMove move.Move
	var overallBestScore int

	alpha := -Infinity
	beta := Infinity

	for depth := 1; depth <= maxDepth; depth++ {
		// Aspiration Window
		window := 50
		if depth >= 4 {
			alpha = overallBestScore - window
			beta = overallBestScore + window
		} else {
			alpha = -Infinity
			beta = Infinity
		}

		researches := 0
		for {
			bestMove, score := s.searchRoot(b, depth, alpha, beta)
			if bestMove != move.NullMove {
				overallBestMove = bestMove
				overallBestScore = score
			}

			if score <= alpha {
				alpha -= window * 2
				window *= 2
			} else if score >= beta {
				beta += window * 2
				window *= 2
			} else {
				break
			}

			researches++
			if researches >= 2 {
				// Fallback to full window if failing repeatedly
				alpha = -Infinity
				beta = Infinity
				bestMove, score = s.searchRoot(b, depth, alpha, beta)
				if bestMove != move.NullMove {
					overallBestMove = bestMove
					overallBestScore = score
				}
				break
			}
		}
	}

	return overallBestMove, overallBestScore
}

func (s *Searcher) searchRoot(b *board.Board, depth, alpha, beta int) (move.Move, int) {
	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)

	if list.Count == 0 {
		return move.NullMove, 0
	}

	ttMove := move.NullMove
	if entry, ok := s.TT.Probe(b.Hash); ok {
		ttMove = entry.Move
	}

	s.orderMoves(b, list, ttMove, 0)

	bestMove := list.Moves[0]
	bestScore := -Infinity

	var undo board.Undo

	for i := 0; i < list.Count; i++ {
		m := list.Moves[i]
		b.MakeMove(m, &undo)

		score := 0
		if i == 0 {
			score = -s.alphaBeta(b, depth-1, -beta, -alpha, 1)
		} else {
			score = -s.alphaBeta(b, depth-1, -alpha-1, -alpha, 1)
			if score > alpha && score < beta {
				score = -s.alphaBeta(b, depth-1, -beta, -alpha, 1)
			}
		}

		b.UnmakeMove(&undo)

		if score > bestScore {
			bestScore = score
			bestMove = m
		}

		if score > alpha {
			alpha = score
		}
	}

	s.TT.Store(b.Hash, depth, tt.Exact, bestScore, bestMove)
	return bestMove, bestScore
}

// Principal Variation Search (PVS) with Alpha-Beta Pruning.
func (s *Searcher) alphaBeta(b *board.Board, depth, alpha, beta, ply int) int {
	if depth <= 0 {
		return s.quiescence(b, alpha, beta, ply)
	}

	s.Nodes++

	us := b.SideToMove
	them := us.Opposite()
	kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
	inCheck := attack.IsSquareAttacked(b, kingPos, them)

	// Check extension
	if inCheck {
		depth++
	}

	// 1. Transposition Table Lookup
	ttMove := move.NullMove
	if entry, ok := s.TT.Probe(b.Hash); ok && int(entry.Depth) >= depth {
		if entry.Flag == tt.Exact {
			return int(entry.Score)
		} else if entry.Flag == tt.LowerBound && int(entry.Score) >= beta {
			return int(entry.Score)
		} else if entry.Flag == tt.UpperBound && int(entry.Score) <= alpha {
			return int(entry.Score)
		}
		ttMove = entry.Move
	}

	// 2. Null-Move Pruning (NMP)
	if depth >= 3 && !inCheck && ply > 0 {
		nonPawnPieces := b.Colors[us] &^ b.Pieces[board.Pawn] &^ b.Pieces[board.King]
		if nonPawnPieces != 0 {
			b.SideToMove = them
			oldEP := b.EnPassant
			b.EnPassant = board.NoSquare

			nullScore := -s.alphaBeta(b, depth-1-2, -beta, -beta+1, ply+1)

			b.SideToMove = us
			b.EnPassant = oldEP

			if nullScore >= beta {
				return beta
			}
		}
	}

	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)

	if list.Count == 0 {
		if inCheck {
			return -MateVal + ply
		}
		return 0
	}

	s.orderMoves(b, list, ttMove, ply)

	bestMove := move.NullMove
	bestScore := -Infinity
	origAlpha := alpha

	var undo board.Undo

	for i := 0; i < list.Count; i++ {
		m := list.Moves[i]
		isQuiet := !m.IsCapture() && !m.IsPromotion()

		b.MakeMove(m, &undo)

		score := 0
		if i == 0 {
			score = -s.alphaBeta(b, depth-1, -beta, -alpha, ply+1)
		} else {
			// 3. Late Move Reduction (LMR) & Zero Window Search
			reduction := 0
			if i >= 3 && depth >= 3 && isQuiet && !inCheck {
				reduction = 1
			}

			score = -s.alphaBeta(b, depth-1-reduction, -alpha-1, -alpha, ply+1)

			if score > alpha && score < beta {
				score = -s.alphaBeta(b, depth-1, -beta, -alpha, ply+1)
			}
		}

		b.UnmakeMove(&undo)

		if score >= beta {
			if isQuiet && ply < MaxDepth {
				s.KillerMoves[ply][1] = s.KillerMoves[ply][0]
				s.KillerMoves[ply][0] = m
				s.History[us][m.From()][m.To()] += depth * depth
			}

			s.TT.Store(b.Hash, depth, tt.LowerBound, beta, m)
			return beta
		}

		if score > bestScore {
			bestScore = score
			bestMove = m
		}

		if score > alpha {
			alpha = score
		}
	}

	flag := tt.Exact
	if bestScore <= origAlpha {
		flag = tt.UpperBound
	}
	s.TT.Store(b.Hash, depth, flag, bestScore, bestMove)

	return bestScore
}

// Quiescence Search with Delta Pruning.
func (s *Searcher) quiescence(b *board.Board, alpha, beta, ply int) int {
	s.Nodes++
	standPat := evaluation.Evaluate(b)

	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}

	// Cap quiescence depth to 12 plies to prevent midgame explosion
	if ply >= MaxDepth || ply >= 20 {
		return standPat
	}

	pseudo := &movegen.MoveList{}
	movegen.GeneratePseudoLegalMoves(b, pseudo)

	us := b.SideToMove
	them := us.Opposite()
	var undo board.Undo

	for i := 0; i < pseudo.Count; i++ {
		m := pseudo.Moves[i]
		if !m.IsCapture() {
			continue
		}

		// Delta Pruning: If standPat + value of captured piece + 200 < alpha, prune!
		capturedSq := board.Square(m.To())
		capturedPiece := b.PieceAt(capturedSq).Type()
		capturedVal := pieceValue(capturedPiece)

		if standPat+capturedVal+200 < alpha && !m.IsPromotion() {
			continue
		}

		b.MakeMove(m, &undo)

		kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
		if attack.IsSquareAttacked(b, kingPos, them) {
			b.UnmakeMove(&undo)
			continue
		}

		score := -s.quiescence(b, -beta, -alpha, ply+1)
		b.UnmakeMove(&undo)

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

func pieceValue(pt board.PieceType) int {
	switch pt {
	case board.Pawn:
		return 100
	case board.Knight:
		return 320
	case board.Bishop:
		return 330
	case board.Rook:
		return 500
	case board.Queen:
		return 900
	default:
		return 0
	}
}

type scoredMove struct {
	m     move.Move
	score int
}

func (s *Searcher) orderMoves(b *board.Board, list *movegen.MoveList, ttMove move.Move, ply int) {
	us := b.SideToMove
	scored := make([]scoredMove, list.Count)

	for i := 0; i < list.Count; i++ {
		m := list.Moves[i]
		score := 0

		if m == ttMove {
			score = 1000000
		} else if m.IsCapture() {
			victimSq := board.Square(m.To())
			victim := b.PieceAt(victimSq).Type()
			attackerSq := board.Square(m.From())
			attacker := b.PieceAt(attackerSq).Type()

			score = 10000 + (int(victim)*100 - int(attacker))
		} else {
			if ply < MaxDepth {
				if m == s.KillerMoves[ply][0] {
					score = 9000
				} else if m == s.KillerMoves[ply][1] {
					score = 8000
				} else {
					score = s.History[us][m.From()][m.To()]
				}
			}
		}

		scored[i] = scoredMove{m: m, score: score}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	for i := 0; i < list.Count; i++ {
		list.Moves[i] = scored[i].m
	}
}
