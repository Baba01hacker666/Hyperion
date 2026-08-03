package uci

import (
	"bufio"
	"fmt"
	"hyperion/internal/board"
	"hyperion/internal/evaluation"
	"hyperion/internal/move"
	"hyperion/internal/movegen"
	"hyperion/internal/opening"
	"hyperion/internal/search"
	"os"
	"strconv"
	"strings"
	"time"
)

var hashSizeMB = 64
var numThreads = 1

// Loop starts the UCI protocol listening loop.
func Loop() {
	reader := bufio.NewReader(os.Stdin)
	b := board.New()
	b.SetFEN(board.StartFEN)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "uci":
			fmt.Println("id name Hyperion")
			fmt.Println("id author BABA01HACKER")
			fmt.Println("option name Style type combo default normal var normal var gamble var defense var evil var blitz")
			fmt.Println("option name Hash type spin default 64 min 1 max 1024")
			fmt.Println("option name Threads type spin default 1 min 1 max 128")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "setoption":
			parseSetOption(parts[1:])
		case "ucinewgame":
			b = board.New()
			b.SetFEN(board.StartFEN)
		case "position":
			parsePosition(b, parts[1:])
		case "go":
			parseGo(b, parts[1:])
		case "quit":
			return
		}
	}
}

func parseSetOption(args []string) {
	name := ""
	value := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "name" && i+1 < len(args) {
			name = strings.ToLower(args[i+1])
		}
		if args[i] == "value" && i+1 < len(args) {
			value = strings.ToLower(args[i+1])
		}
	}

	if name == "style" {
		switch value {
		case "gamble":
			evaluation.SetStyle(evaluation.StyleGamble)
		case "defense":
			evaluation.SetStyle(evaluation.StyleDefense)
		case "evil":
			evaluation.SetStyle(evaluation.StyleEvil)
		case "blitz":
			evaluation.SetStyle(evaluation.StyleBlitz)
		default:
			evaluation.SetStyle(evaluation.StyleBalanced)
		}
	} else if name == "hash" {
		if sz, err := strconv.Atoi(value); err == nil && sz >= 1 {
			hashSizeMB = sz
		}
	} else if name == "threads" {
		if t, err := strconv.Atoi(value); err == nil && t >= 1 {
			numThreads = t
		}
	}
}

func parsePosition(b *board.Board, args []string) {
	if len(args) == 0 {
		return
	}

	idx := 0
	if args[0] == "startpos" {
		b.SetFEN(board.StartFEN)
		idx = 1
	} else if args[0] == "fen" {
		var fenParts []string
		idx = 1
		for idx < len(args) && args[idx] != "moves" {
			fenParts = append(fenParts, args[idx])
			idx++
		}
		for len(fenParts) < 6 {
			if len(fenParts) == 4 {
				fenParts = append(fenParts, "0")
			} else if len(fenParts) == 5 {
				fenParts = append(fenParts, "1")
			} else {
				break
			}
		}
		fen := strings.Join(fenParts, " ")
		b.SetFEN(fen)
	}

	if idx < len(args) && args[idx] == "moves" {
		idx++
		for ; idx < len(args); idx++ {
			moveStr := args[idx]
			list := &movegen.MoveList{}
			movegen.GenerateLegalMoves(b, list)
			for i := 0; i < list.Count; i++ {
				m := list.Moves[i]
				if m.String() == moveStr {
					var undo board.Undo
					b.MakeMove(m, &undo)
					break
				}
			}
		}
	}
}

func parseGo(b *board.Board, args []string) {
	// Check opening book first
	if bookMove := opening.GetBookMove(b); bookMove != move.NullMove {
		fmt.Printf("info string opening book move\n")
		fmt.Printf("bestmove %s\n", bookMove.String())
		return
	}

	var limits search.Limits

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "depth":
			if i+1 < len(args) {
				if d, err := strconv.Atoi(args[i+1]); err == nil {
					limits.Depth = d
				}
			}
		case "wtime":
			if i+1 < len(args) && b.SideToMove == board.White {
				if t, err := strconv.Atoi(args[i+1]); err == nil {
					limits.Time = time.Duration(t) * time.Millisecond
				}
			}
		case "btime":
			if i+1 < len(args) && b.SideToMove == board.Black {
				if t, err := strconv.Atoi(args[i+1]); err == nil {
					limits.Time = time.Duration(t) * time.Millisecond
				}
			}
		case "winc":
			if i+1 < len(args) && b.SideToMove == board.White {
				if inc, err := strconv.Atoi(args[i+1]); err == nil {
					limits.Inc = time.Duration(inc) * time.Millisecond
				}
			}
		case "binc":
			if i+1 < len(args) && b.SideToMove == board.Black {
				if inc, err := strconv.Atoi(args[i+1]); err == nil {
					limits.Inc = time.Duration(inc) * time.Millisecond
				}
			}
		case "movestogo":
			if i+1 < len(args) {
				if mtg, err := strconv.Atoi(args[i+1]); err == nil {
					limits.MovesToGo = mtg
				}
			}
		case "movetime":
			if i+1 < len(args) {
				if mt, err := strconv.Atoi(args[i+1]); err == nil {
					limits.MoveTime = time.Duration(mt) * time.Millisecond
				}
			}
		}
	}

	searcher := search.NewSearcher(hashSizeMB)
	searcher.Threads = numThreads
	bestMove, score := searcher.SearchWithLimits(b, limits)

	fmt.Printf("info depth %d score cp %d nodes %d\n", limits.Depth, score, searcher.Nodes.Load())
	fmt.Printf("bestmove %s\n", bestMove.String())
}
