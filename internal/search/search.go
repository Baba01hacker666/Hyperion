package search

import (
	"hyperion/internal/attack"
	"hyperion/internal/board"
	"hyperion/internal/evaluation"
	"hyperion/internal/move"
	"hyperion/internal/movegen"
	"hyperion/internal/tt"
	"hyperion/internal/zobrist"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
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

type stackEntry struct {
	eval  int
	move  move.Move
	piece board.PieceType
}

type pawnEntry struct {
	hash  uint64
	score int
}

// Worker holds per-thread state to avoid data races.
type Worker struct {
	id             int
	searcher       *Searcher
	board          *board.Board
	nodes          uint64
	stack          [MaxDepth + 4]stackEntry
	pawnTT         [4096]pawnEntry
	KillerMoves    [MaxDepth][2]move.Move
	History        [2][64][64]int
	CaptureHistory [6][64][6]int
	ContHist1      [6][64][6][64]int
	ContHist2      [6][64][6][64]int
	CounterMoves   [64][64]move.Move
}

type rootSearchResult struct {
	bestMove move.Move
	score    int
	complete bool
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
			s.MaxTime = 1000 * time.Millisecond
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
	overallBestMove := w.rootFallbackMove(w.board)
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

			result := w.searchRoot(w.board, depth, alpha, beta)
			if result.complete && result.bestMove != move.NullMove {
				overallBestMove = result.bestMove
				overallBestScore = result.score
			}

			if !result.complete {
				break
			}

			if result.score <= alpha {
				window *= 2
				alpha = overallBestScore - window
				if alpha < -Infinity/2 {
					alpha = -Infinity
				}
			} else if result.score >= beta {
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
				result = w.searchRoot(w.board, depth, alpha, beta)
				if result.complete && result.bestMove != move.NullMove {
					overallBestMove = result.bestMove
					overallBestScore = result.score
				}
				break
			}
		}

		if !w.searcher.Stopped.Load() {
			elapsed := time.Since(w.searcher.StartTime).Milliseconds()
			nodes := w.searcher.Nodes.Load()
			var nps int64
			if elapsed > 0 {
				nps = int64(nodes) * 1000 / elapsed
			}
			fmt.Printf("info depth %d score cp %d time %d nodes %d nps %d pv %s\n",
				depth, overallBestScore, elapsed, nodes, nps, overallBestMove.String())
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

func (w *Worker) rootFallbackMove(b *board.Board) move.Move {
	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)
	if list.Count == 0 {
		return move.NullMove
	}

	ttMove := move.NullMove
	if entry, ok := w.searcher.TT.Probe(b.Hash); ok {
		ttMove = entry.Move
	}
	w.orderMoves(b, list, ttMove, 0, move.NullMove, board.NoPieceType, move.NullMove, board.NoPieceType)
	return list.Moves[0]
}

func (w *Worker) searchRoot(b *board.Board, depth, alpha, beta int) rootSearchResult {
	list := &movegen.MoveList{}
	movegen.GenerateLegalMoves(b, list)

	if list.Count == 0 {
		return rootSearchResult{bestMove: move.NullMove, score: 0, complete: true}
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

	w.orderMoves(b, list, ttMove, 0, move.NullMove, board.NoPieceType, move.NullMove, board.NoPieceType)

	bestMove := move.NullMove
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
		pt := b.PieceAt(board.Square(m.From())).Type()

		b.MakeMove(m, &undo)

		w.stack[1].move = m
		w.stack[1].piece = pt

		score := 0
		if i == 0 {
			score = -w.alphaBeta(b, searchDepth-1, -beta, -alpha, 1, m, pt)
		} else {
			score = -w.alphaBeta(b, searchDepth-1, -alpha-1, -alpha, 1, m, pt)
			if score > alpha && score < beta {
				score = -w.alphaBeta(b, searchDepth-1, -beta, -alpha, 1, m, pt)
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

	complete := !w.searcher.Stopped.Load()
	if complete && bestMove != move.NullMove {
		w.searcher.TT.Store(b.Hash, depth, tt.Exact, bestScore, bestMove)
	}
	return rootSearchResult{bestMove: bestMove, score: bestScore, complete: complete}
}

// Principal Variation Search (PVS) with Alpha-Beta Pruning.
func (w *Worker) alphaBeta(b *board.Board, depth, alpha, beta, ply int, prevMove move.Move, prevPiece board.PieceType) int {
	if w.searcher.Stopped.Load() {
		return 0
	}

	w.nodes++
	w.searcher.Nodes.Add(1)

	if w.nodes%2048 == 0 && time.Since(w.searcher.StartTime) > w.searcher.MaxTime {
		w.searcher.Stopped.Store(true)
		return 0
	}

	if ply > 0 && b.Is2FoldRepetition() {
		eval := evaluation.Evaluate(b)
		if eval > 100 {
			return -50 // Penalize draws when we're winning
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

	// Static evaluation
	staticEval := evaluation.Evaluate(b)
	w.stack[ply+2].eval = staticEval

	// Accurate improving calculation comparing against 2 plies ago
	improving := false
	if !inCheck && ply >= 2 {
		improving = staticEval > w.stack[ply].eval
	} else if !inCheck {
		improving = staticEval > 0
	}

	// Internal Iterative Deepening (IID): if no TT move and depth is large,
	// do a shallow search to find a good move to try first.
	if ttMove == move.NullMove && depth >= 6 && !inCheck {
		w.alphaBeta(b, depth-3, alpha, beta, ply, prevMove, prevPiece)
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
			margin -= 20
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
			b.Hash ^= zobrist.SideToMove
			oldEP := b.EnPassant
			if oldEP != board.NoSquare {
				b.Hash ^= zobrist.EnPassant[oldEP]
			}
			b.EnPassant = board.NoSquare

			R := 3 + depth/4
			nullScore := -w.alphaBeta(b, depth-1-R, -beta, -beta+1, ply+1, move.NullMove, board.NoPieceType)

			b.SideToMove = us
			b.Hash ^= zobrist.SideToMove
			if oldEP != board.NoSquare {
				b.Hash ^= zobrist.EnPassant[oldEP]
			}
			b.EnPassant = oldEP

			if w.searcher.Stopped.Load() {
				return 0
			}

			if nullScore >= beta && abs(nullScore) < MateVal-100 {
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

	prevPrevMove := move.NullMove
	prevPrevPiece := board.NoPieceType
	if ply >= 2 {
		prevPrevMove = w.stack[ply].move
		prevPrevPiece = w.stack[ply].piece
	}

	w.orderMoves(b, list, ttMove, ply, prevMove, prevPiece, prevPrevMove, prevPrevPiece)

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
				maxQuiets += depth
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

		pt := b.PieceAt(board.Square(m.From())).Type()
		var victimType board.PieceType
		if isCapture {
			victimType = b.PieceAt(board.Square(m.To())).Type()
		}

		b.MakeMove(m, &undo)

		w.stack[ply+1].move = m
		w.stack[ply+1].piece = pt

		score := 0
		if i == 0 {
			score = -w.alphaBeta(b, depth-1, -beta, -alpha, ply+1, m, pt)
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
					if !improving {
						reduction++
					}
					histScore := w.History[us][m.From()][m.To()]
					if histScore > 3000 {
						reduction--
					} else if histScore < -1000 {
						reduction++
					}
					if prevPiece < 6 && pt < 6 {
						contScore := w.ContHist1[prevPiece][prevMove.To()][pt][m.To()]
						if contScore > 2000 {
							reduction--
						} else if contScore < -1000 {
							reduction++
						}
					}
				} else {
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

			score = -w.alphaBeta(b, depth-1-reduction, -alpha-1, -alpha, ply+1, m, pt)

			if score > alpha && reduction > 0 {
				score = -w.alphaBeta(b, depth-1, -alpha-1, -alpha, ply+1, m, pt)
			}

			if score > alpha && score < beta {
				score = -w.alphaBeta(b, depth-1, -beta, -alpha, ply+1, m, pt)
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

				if prevPiece < 6 && pt < 6 {
					applyHistoryBonus(&w.ContHist1[prevPiece][prevMove.To()][pt][m.To()], bonus, 10000)
				}
				if prevPrevPiece < 6 && pt < 6 {
					applyHistoryBonus(&w.ContHist2[prevPrevPiece][prevPrevMove.To()][pt][m.To()], bonus, 10000)
				}

				malus := -(depth * depth / 2)
				for j := 0; j < i; j++ {
					prev := list.Moves[j]
					if !prev.IsCapture() && !prev.IsPromotion() && prev != m {
						prevPt := b.PieceAt(board.Square(prev.From())).Type()
						applyHistoryBonus(&w.History[us][prev.From()][prev.To()], malus, 10000)
						if prevPiece < 6 && prevPt < 6 {
							applyHistoryBonus(&w.ContHist1[prevPiece][prevMove.To()][prevPt][prev.To()], malus, 10000)
						}
					}
				}

				if prevMove != move.NullMove {
					w.CounterMoves[prevMove.From()][prevMove.To()] = m
				}
			} else if isCapture && !isPromotion {
				bonus := depth * depth
				vt := victimType
				if vt >= 6 {
					vt = board.Pawn
				}
				ptAtt := pt
				if ptAtt >= 6 {
					ptAtt = board.Pawn
				}
				applyHistoryBonus(&w.CaptureHistory[ptAtt][m.To()][vt], bonus, 10000)
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

// Quiescence Search with SEE and Check Evasions.
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

var mvvLvaValues = [7]int{0, 100, 200, 300, 400, 500, 600}

func getMvvLvaScore(victim, attacker board.PieceType) int {
	return mvvLvaValues[victim]*10 - mvvLvaValues[attacker]
}

func (w *Worker) scoreMove(b *board.Board, m move.Move, ttMove move.Move, ply int,
	prevMove move.Move, prevPiece board.PieceType,
	prevPrevMove move.Move, prevPrevPiece board.PieceType) int {
	us := b.SideToMove

	if m == ttMove {
		return 10_000_000
	}

	if m.IsCapture() {
		victimSq := board.Square(m.To())
		victim := b.PieceAt(victimSq).Type()
		if victim >= 6 {
			victim = board.Pawn
		}
		attackerSq := board.Square(m.From())
		attacker := b.PieceAt(attackerSq).Type()
		if attacker >= 6 {
			attacker = board.Pawn
		}
		mvvLva := getMvvLvaScore(victim, attacker)

		if evaluation.SEE(b, m, 0) {
			captHist := w.CaptureHistory[attacker][m.To()][victim]
			return 100_000 + mvvLva + captHist/100
		}
		return -50_000 + mvvLva
	}

	pt := b.PieceAt(board.Square(m.From())).Type()
	if ply < MaxDepth {
		if m == w.KillerMoves[ply][0] {
			return 9000
		}
		if m == w.KillerMoves[ply][1] {
			return 8000
		}
		if prevMove != move.NullMove && m == w.CounterMoves[prevMove.From()][prevMove.To()] {
			return 7000
		}
	}

	score := w.History[us][m.From()][m.To()]

	if prevPiece < 6 && pt < 6 {
		score += w.ContHist1[prevPiece][prevMove.To()][pt][m.To()] * 2
	}
	if prevPrevPiece < 6 && pt < 6 {
		score += w.ContHist2[prevPrevPiece][prevPrevMove.To()][pt][m.To()]
	}

	return score
}

// orderMoves uses partial selection sort (pick-best) for high performance.
func (w *Worker) orderMoves(b *board.Board, list *movegen.MoveList, ttMove move.Move, ply int,
	prevMove move.Move, prevPiece board.PieceType,
	prevPrevMove move.Move, prevPrevPiece board.PieceType) {
	scores := make([]int, list.Count)
	for i := 0; i < list.Count; i++ {
		scores[i] = w.scoreMove(b, list.Moves[i], ttMove, ply, prevMove, prevPiece, prevPrevMove, prevPrevPiece)
	}

	for i := 0; i < list.Count; i++ {
		bestIdx := i
		for j := i + 1; j < list.Count; j++ {
			if scores[j] > scores[bestIdx] {
				bestIdx = j
			}
		}
		if bestIdx != i {
			list.Moves[i], list.Moves[bestIdx] = list.Moves[bestIdx], list.Moves[i]
			scores[i], scores[bestIdx] = scores[bestIdx], scores[i]
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
