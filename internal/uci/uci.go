package uci

import (
	"bufio"
	"fmt"
	"hyperion/internal/board"
	"hyperion/internal/movegen"
	"hyperion/internal/search"
	"os"
	"strconv"
	"strings"
)

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
			fmt.Println("id author Antigravity")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
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
	depth := 5 // Default fixed depth for now

	for i := 0; i < len(args); i++ {
		if args[i] == "depth" && i+1 < len(args) {
			if d, err := strconv.Atoi(args[i+1]); err == nil {
				depth = d
			}
		}
	}

	searcher := search.NewSearcher(64)
	bestMove, score := searcher.Search(b, depth)

	fmt.Printf("info depth %d score cp %d nodes %d\n", depth, score, searcher.Nodes)
	fmt.Printf("bestmove %s\n", bestMove.String())
}
