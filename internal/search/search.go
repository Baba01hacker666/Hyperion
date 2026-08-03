package search

import (
	"hyperion/internal/attack"
	"hyperion/internal/board"
	"hyperion/internal/evaluation"
	"hyperion/internal/move"
	"hyperion/internal/movegen"
	"hyperion/internal/tt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Infinity = 30000
	MateVal  = 20000
	MaxDepth = 64
)

var lmrTable [MaxDepth][64]int

func init() {
	for depth := 1; depth < MaxDepth; depth++ {
		for moves := 1; moves < 64; moves++ {
			lmrTable[depth][moves] = int(0.75 + math.Log(float64(depth))*math.Log(float64(moves))/2.25)
		}
	}
}

func applyHistoryBonus(entry *int, bonus int, maxVal int) {
	if bonus > maxVal {
		bonus = maxVal
	} else if bonus < -maxVal {
		bonus = -maxVal
	}
	absVal := *entry
	if absVal < 0 {
		absVal = -absVal
	}
	*entry += bonus - (bonus * absVal / maxVal)
}

type Limits struct {
	Depth     int
	Nodes     uint64
	MoveTime  time.Duration
	Time      time.Duration
	Inc       time.Duration
	MovesToGo int
}

// Searcher holds shared state for multi-threaded Lazy SMP search.
type Searcher struct {
	Threads   int
	TT        *tt.Table
	StartTime time.Time
	MaxTime   time.Duration
	Stopped   atomic.Bool
	Nodes     atomic.Uint64
}

// Worker holds per-thread state to avoid data races.
type Worker struct {
	id           int
	searcher     *Searcher
	board        *board.Board
	nodes        uint64
	KillerMoves  [MaxDepth][2]move.Move
	History      [2][64][64]int
	CounterMoves [64][64]move.Move
}

func NewSearcher(ttSizeMB int) *Searcher {
	return &Searcher{
		TT:      tt.NewTable(ttSizeMB),
		Threads: 1,
	}
}

// Search performs multi-threaded Iterative Deepening with Aspiration Windows.
func (s *Searcher) Search(b *board.Board, maxDepth int) (move.Move, int) {
	return s.SearchWithLimits(b, Limits{Depth: maxDepth})
}

// SearchWithLimits handles dynamic time allocation and search bounds.
func (s *Searcher) SearchWithLimits(b *board.Board, limits Limits) (move.Move, int) {
	s.Nodes.Store(0)
	s.Stopped.Store(false)
	s.StartTime = time.Now()

	if limits.MoveTime > 0 {
		s.MaxTime = limits.MoveTime
	} else if limits.Time > 0 {
		movesToGo := limits.MovesToGo
		if movesToGo <= 0 {
			movesToGo = 30
		}
		s.MaxTime = (limits.Time / time.Duration(movesToGo)) + (limits.Inc * 3 / 4)
		if s.MaxTime < 50*time.Millisecond {
			s.MaxTime = 50 * time.Millisecond
		}
	} else {
		s.MaxTime = 4000 * time.Millisecond
		if evaluation.CurrentStyle == evaluation.StyleEvil {
			s.MaxTime = 12000 * time.Millisecond
		} else if evaluation.CurrentStyle == evaluation.StyleBlitz {
			s.MaxTime = 1000 * time.Millisecond // Blitz Mode: 1 second per move!
		}
	}

	maxDepth := limits.Depth
	if maxDepth <= 0 {
		maxDepth = MaxDepth
	}

	numThreads := s.Threads
	if numThreads < 1 {
		numThreads = 1
	}

	mainWorker := &Worker{
		id:       0,
		searcher: s,
		board:    b.Clone(),
	}

	var wg sync.WaitGroup

	// Spawn helper worker threads for Lazy SMP
	for i := 1; i < numThreads; i++ {
		wg.Add(1)
		helperWorker := &Worker{
			id:       i,
			searcher: s,
			board:    b.Clone(),
		}
		go func(w *Worker) {
			defer wg.Done()
			w.runHelperSearch(maxDepth)
		}(helperWorker)
	}

	// Main worker drives search and returns best move and score
	bestMove, bestScore := mainWorker.runMainSearch(maxDepth)

	// Signal helper threads to stop and wait for completion
	s.Stopped.Store(true)
	wg.Wait()

	return bestMove, bestScore
}

func (w *Worker) runMainSearch(maxDepth int) (move.Move, int) {
	var overallBestMove move.Move
	var overallBestScore int

	alpha := -Infinity
	beta := Infinity

	for depth := 1; depth <= maxDepth; depth++ {
		if w.searcher.Stopped.Load() {
			break
		}
		if depth > 1 && time.Since(w.searcher.StartTime) > w.searcher.MaxTime {
			w.searcher.Stopped.Store(true)
			break
		}

		window := 25
		if depth >= 4 {
			alpha = overallBestScore - window
			beta = overallBestScore + window
		} else {
			alpha = -Infinity
			beta = Infinity
		}

		researches := 0
		for {
			if w.searcher.Stopped.Load() {
				break
			}

			bestMove, score := w.searchRoot(w.board, depth, alpha, beta)
			if bestMove != move.NullMove && !w.searcher.Stopped.Load() {
				overallBestMove = bestMove
				overallBestScore = score
			}

			if score <= alpha {
				window *= 2
				alpha = overallBestScore - window
				if alpha < -Infinity/2 {
					alpha = -Infinity
				}
			} else if score >= beta {
				window *= 2
				beta = overallBestScore + window
				if beta > Infinity/2 {
					beta = Infinity
				}
			} else {
				break
			}

			researches++
			if researches >= 3 {
				alpha = -Infinity
				beta = Infinity
				bestMove, score = w.searchRoot(w.board, depth, alpha, beta)
				if bestMove != move.NullMove && !w.searcher.Stopped.Load() {
					overallBestMove = bestMove
					overallBestScore = score
				}
				break
			}
		}
	}

	return overallBestMove, overallBestScore
}

func (w *Worker) runHelperSearch(maxDepth int) {
	targetDepth := maxDepth
	if targetDepth < MaxDepth {
		targetDepth = maxDepth + 2
	}

	for depth := 1; depth <= targetDepth; depth++ {
		if w.searcher.Stopped.Load() {
			break
		}
		if time.Since(w.searcher.StartTime) > w.searcher.MaxTime {
			break
		}

		w.searchRoot(w.board, depth, -Infinity, Infinity)
	}
}

func (w *Worker) searchRoot(b *board.Board, depth, alpha, beta int) (move.Move, int) {
	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)

	if list.Count == 0 {
		return move.NullMove, 0
	}

	us := b.SideToMove
	them := us.Opposite()
	kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
	inCheck := attack.IsSquareAttacked(b, kingPos, them)

	searchDepth := depth
	if inCheck {
		searchDepth++
	}

	ttMove := move.NullMove
	if entry, ok := w.searcher.TT.Probe(b.Hash); ok {
		ttMove = entry.Move
	}

	w.orderMoves(b, list, ttMove, 0, move.NullMove)

	bestMove := list.Moves[0]
	bestScore := -Infinity

	var undo board.Undo

	for i := 0; i < list.Count; i++ {
		if w.searcher.Stopped.Load() {
			break
		}
		if i > 0 && time.Since(w.searcher.StartTime) > w.searcher.MaxTime {
			w.searcher.Stopped.Store(true)
			break
		}
		m := list.Moves[i]
		b.MakeMove(m, &undo)

		score := 0
		if i == 0 {
			score = -w.alphaBeta(b, searchDepth-1, -beta, -alpha, 1, m)
		} else {
			score = -w.alphaBeta(b, searchDepth-1, -alpha-1, -alpha, 1, m)
			if score > alpha && score < beta {
				score = -w.alphaBeta(b, searchDepth-1, -beta, -alpha, 1, m)
			}
		}

		b.UnmakeMove(&undo)

		if w.searcher.Stopped.Load() {
			break
		}

		if score > bestScore {
			bestScore = score
			bestMove = m
		}

		if score > alpha {
			alpha = score
		}
	}

	if !w.searcher.Stopped.Load() {
		w.searcher.TT.Store(b.Hash, depth, tt.Exact, bestScore, bestMove)
	}
	return bestMove, bestScore
}

// Principal Variation Search (PVS) with Alpha-Beta Pruning.
func (w *Worker) alphaBeta(b *board.Board, depth, alpha, beta, ply int, prevMove move.Move) int {
	if w.searcher.Stopped.Load() {
		return 0
	}

	w.nodes++
	w.searcher.Nodes.Add(1)

	if w.nodes%2048 == 0 && time.Since(w.searcher.StartTime) > w.searcher.MaxTime {
		w.searcher.Stopped.Store(true)
		return 0
	}

	if ply > 0 && b.IsRepetition() {
		eval := evaluation.Evaluate(b)
		if eval > 100 {
			return -50 // Penalize draws when winning to avoid 3-fold repetition
		}
		return 0
	}

	if depth <= 0 {
		return w.quiescence(b, alpha, beta, ply)
	}

	us := b.SideToMove
	them := us.Opposite()
	kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
	inCheck := attack.IsSquareAttacked(b, kingPos, them)

	if inCheck {
		depth++
	}

	// 1. Transposition Table Lookup
	ttMove := move.NullMove
	ttDepth := 0
	ttScore := 0
	ttFlag := tt.Exact
	if entry, ok := w.searcher.TT.Probe(b.Hash); ok {
		ttMove = entry.Move
		ttDepth = int(entry.Depth)
		ttScore = int(entry.Score)
		ttFlag = entry.Flag
		// Only use TT score for cutoff if depth is sufficient
		if ttDepth >= depth {
			if ttFlag == tt.Exact {
				return ttScore
			} else if ttFlag == tt.LowerBound && ttScore >= beta {
				return ttScore
			} else if ttFlag == tt.UpperBound && ttScore <= alpha {
				return ttScore
			}
		}
	}

	// Static evaluation (computed once, used by multiple pruning steps)
	staticEval := evaluation.Evaluate(b)

	// Improving: is our position better than 2 plies ago?
	// (We approximate by just using current eval vs zero for simplicity)
	improving := !inCheck && staticEval > 0

	// Internal Iterative Deepening (IID): if no TT move and depth is large,
	// do a shallow search to find a good move to try first.
	if ttMove == move.NullMove && depth >= 5 && !inCheck {
		w.alphaBeta(b, depth-4, alpha, beta, ply, prevMove)
		if entry, ok := w.searcher.TT.Probe(b.Hash); ok {
			ttMove = entry.Move
		}
	}

	// 2. Razoring
	if depth <= 3 && !inCheck && ply > 0 {
		if staticEval < alpha-300-150*depth {
			qScore := w.quiescence(b, alpha, beta, ply)
			if qScore < alpha {
				return qScore
			}
		}
	}

	// 3. Reverse Futility Pruning (RFP / Static Null Move Pruning)
	if depth <= 7 && !inCheck && ply > 0 {
		margin := 70 * depth
		if improving {
			margin -= 20 // Less aggressive pruning when improving
		}
		if staticEval-margin >= beta {
			return staticEval - margin
		}
	}

	// 4. Null-Move Pruning (NMP)
	if depth >= 3 && !inCheck && ply > 0 {
		nonPawnPieces := b.Colors[us] &^ b.Pieces[board.Pawn] &^ b.Pieces[board.King]
		if nonPawnPieces != 0 {
			b.SideToMove = them
			oldEP := b.EnPassant
			b.EnPassant = board.NoSquare

			R := 3 + depth/4
			nullScore := -w.alphaBeta(b, depth-1-R, -beta, -beta+1, ply+1, move.NullMove)

			b.SideToMove = us
			b.EnPassant = oldEP

			if w.searcher.Stopped.Load() {
				return 0
			}

			if nullScore >= beta {
				return beta
			}
		}
	}

	// 5. Extended Futility Pruning (depth <= 3)
	futilityPruning := false
	if depth <= 3 && !inCheck && ply > 0 {
		futilityMargin := 100 + 120*depth
		if improving {
			futilityMargin += 50
		}
		if staticEval+futilityMargin <= alpha {
			futilityPruning = true
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

	w.orderMoves(b, list, ttMove, ply, prevMove)

	bestMove := move.NullMove
	bestScore := -Infinity
	origAlpha := alpha

	var undo board.Undo
	quietCount := 0

	for i := 0; i < list.Count; i++ {
		if w.searcher.Stopped.Load() {
			return 0
		}

		m := list.Moves[i]
		isCapture := m.IsCapture()
		isPromotion := m.IsPromotion()
		isQuiet := !isCapture && !isPromotion

		if isQuiet {
			quietCount++

			// Late Move Pruning (LMP)
			maxQuiets := 3 + depth*depth
			if improving {
				maxQuiets += depth // Allow more moves when improving
			}
			if depth <= 5 && !inCheck && ply > 0 && quietCount > maxQuiets {
				continue
			}

			// Skip quiet moves if futility pruning condition holds
			if futilityPruning && i > 0 {
				continue
			}
		}

		// SEE-based capture pruning: skip bad captures at shallow depths
		if isCapture && !isPromotion && depth <= 3 && !inCheck && !evaluation.SEE(b, m, -50*depth) {
			continue
		}

		b.MakeMove(m, &undo)

		score := 0
		if i == 0 {
			score = -w.alphaBeta(b, depth-1, -beta, -alpha, ply+1, m)
		} else {
			reduction := 0
			if i >= 3 && depth >= 3 && !inCheck {
				mIdx := i
				if mIdx >= 64 {
					mIdx = 63
				}
				dIdx := depth
				if dIdx >= MaxDepth {
					dIdx = MaxDepth - 1
				}
				if isQuiet {
					reduction = lmrTable[dIdx][mIdx]
					// Reduce more for non-improving positions
					if !improving {
						reduction++
					}
					// History-based LMR adjustment
					histScore := w.History[us][m.From()][m.To()]
					if histScore > 3000 {
						reduction-- // Reduce less for historically good moves
					} else if histScore < -1000 {
						reduction++ // Reduce more for historically bad moves
					}
				} else {
					// Apply small LMR on captures too at high move indices
					if i >= 6 {
						reduction = 1
					}
				}
				if reduction < 0 {
					reduction = 0
				}
				if reduction >= depth {
					reduction = depth - 1
				}
			}

			score = -w.alphaBeta(b, depth-1-reduction, -alpha-1, -alpha, ply+1, m)

			// Re-search if LMR failed high
			if score > alpha && reduction > 0 {
				score = -w.alphaBeta(b, depth-1, -alpha-1, -alpha, ply+1, m)
			}

			if score > alpha && score < beta {
				score = -w.alphaBeta(b, depth-1, -beta, -alpha, ply+1, m)
			}
		}

		b.UnmakeMove(&undo)

		if w.searcher.Stopped.Load() {
			return 0
		}

		if score >= beta {
			if isQuiet && ply < MaxDepth {
				w.KillerMoves[ply][1] = w.KillerMoves[ply][0]
				w.KillerMoves[ply][0] = m

				bonus := depth * depth
				applyHistoryBonus(&w.History[us][m.From()][m.To()], bonus, 10000)

				// History malus: penalize all other quiet moves searched before this
				malus := -(depth * depth / 2)
				for j := 0; j < i; j++ {
					prev := list.Moves[j]
					if !prev.IsCapture() && !prev.IsPromotion() && prev != m {
						applyHistoryBonus(&w.History[us][prev.From()][prev.To()], malus, 10000)
					}
				}

				if prevMove != move.NullMove {
					w.CounterMoves[prevMove.From()][prevMove.To()] = m
				}
			}

			w.searcher.TT.Store(b.Hash, depth, tt.LowerBound, beta, m)
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

	if !w.searcher.Stopped.Load() {
		flag := tt.Exact
		if bestScore <= origAlpha {
			flag = tt.UpperBound
		}
		w.searcher.TT.Store(b.Hash, depth, flag, bestScore, bestMove)
	}

	return bestScore
}

// Quiescence Search with SEE, Delta Pruning, and Check Evasions.
func (w *Worker) quiescence(b *board.Board, alpha, beta, ply int) int {
	if w.searcher.Stopped.Load() {
		return 0
	}

	w.nodes++
	w.searcher.Nodes.Add(1)

	us := b.SideToMove
	them := us.Opposite()
	kingPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
	inCheck := attack.IsSquareAttacked(b, kingPos, them)

	if !inCheck {
		standPat := evaluation.Evaluate(b)
		if standPat >= beta {
			return beta
		}
		if standPat > alpha {
			alpha = standPat
		}
	}

	if ply >= MaxDepth || ply >= 20 {
		return evaluation.Evaluate(b)
	}

	pseudo := &movegen.MoveList{}
	if inCheck {
		movegen.GenerateLegalMoves(b, pseudo)
		if pseudo.Count == 0 {
			return -MateVal + ply
		}
	} else {
		movegen.GeneratePseudoLegalMoves(b, pseudo)
	}

	var undo board.Undo

	for i := 0; i < pseudo.Count; i++ {
		if w.searcher.Stopped.Load() {
			return 0
		}

		m := pseudo.Moves[i]
		if !inCheck {
			if !m.IsCapture() {
				continue
			}
			if !evaluation.SEE(b, m, 0) {
				continue
			}
		}

		b.MakeMove(m, &undo)

		if !inCheck {
			kPos := board.Square((b.Colors[us] & b.Pieces[board.King]).LSB())
			if attack.IsSquareAttacked(b, kPos, them) {
				b.UnmakeMove(&undo)
				continue
			}
		}

		score := -w.quiescence(b, -beta, -alpha, ply+1)
		b.UnmakeMove(&undo)

		if w.searcher.Stopped.Load() {
			return 0
		}

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

type scoredMove struct {
	m     move.Move
	score int
}

var mvvLvaValues = [7]int{0, 100, 200, 300, 400, 500, 600}

func getMvvLvaScore(victim, attacker board.PieceType) int {
	return mvvLvaValues[victim]*10 - mvvLvaValues[attacker]
}

func (w *Worker) orderMoves(b *board.Board, list *movegen.MoveList, ttMove move.Move, ply int, prevMove move.Move) {
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

			mvvLva := getMvvLvaScore(victim, attacker)

			if evaluation.SEE(b, m, 0) {
				score = 100000 + mvvLva
			} else {
				score = -50000 + mvvLva
			}
		} else {
			if ply < MaxDepth {
				if m == w.KillerMoves[ply][0] {
					score = 9000
				} else if m == w.KillerMoves[ply][1] {
					score = 8000
				} else if prevMove != move.NullMove && m == w.CounterMoves[prevMove.From()][prevMove.To()] {
					score = 7000
				} else {
					score = w.History[us][m.From()][m.To()]
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
