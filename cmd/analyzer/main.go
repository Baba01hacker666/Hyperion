package main

import (
	"bufio"
	"flag"
	"fmt"
	"hyperion/internal/board"
	"hyperion/internal/movegen"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type StockfishAnalyzer struct {
	cmd    *exec.Cmd
	stdin  io.Writer
	stdout *bufio.Reader
}

func newAnalyzer() (*StockfishAnalyzer, error) {
	cmd := exec.Command("stockfish")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	sa := &StockfishAnalyzer{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}

	sa.send("uci")
	sa.expect("uciok")
	sa.send("setoption name Skill Level value 20") // Full grandmaster power
	sa.send("isready")
	sa.expect("readyok")

	return sa, nil
}

func (sa *StockfishAnalyzer) send(cmd string) {
	fmt.Fprintln(sa.stdin, cmd)
}

func (sa *StockfishAnalyzer) expect(target string) string {
	for {
		line, err := sa.stdout.ReadString('\n')
		if err != nil {
			return ""
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, target) || strings.Contains(line, target) {
			return line
		}
	}
}

func (sa *StockfishAnalyzer) analyzePosition(fen string, depth int) (string, int) {
	sa.send(fmt.Sprintf("position fen %s", fen))
	sa.send(fmt.Sprintf("go depth %d", depth))

	bestMove := ""
	scoreCP := 0

	for {
		line, err := sa.stdout.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.Contains(line, "score cp") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "cp" && i+1 < len(parts) {
					if sc, err := strconv.Atoi(parts[i+1]); err == nil {
						scoreCP = sc
					}
				}
			}
		}
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
			break
		}
	}

	return bestMove, scoreCP
}

func (sa *StockfishAnalyzer) close() {
	sa.send("quit")
	sa.cmd.Wait()
}

func main() {
	depth := flag.Int("depth", 12, "Stockfish analysis depth")
	movesInput := flag.String("moves", "e2e4 c7c5 g1f3 d7d6 d2d4 c5d4 f3d4 g8f6 b1c3 a7a6", "Space-separated UCI move sequence")
	flag.Parse()

	analyzer, err := newAnalyzer()
	if err != nil {
		log.Fatalf("Failed to launch Stockfish analyzer: %v", err)
	}
	defer analyzer.close()

	fmt.Println("=========================================================")
	fmt.Println("       STOCKFISH GRANDMASTER POST-GAME ANALYZER          ")
	fmt.Println("=========================================================")
	fmt.Printf("Analysis Depth: %d | Analyzing PGN Move Sequence...\n", *depth)
	fmt.Println("---------------------------------------------------------")

	b := board.New()
	b.SetFEN(board.StartFEN)

	moveList := strings.Fields(*movesInput)
	totalCPLoss := 0
	hyperionMoves := 0
	inaccuracies := 0
	mistakes := 0
	blunders := 0

	for i, playedMove := range moveList {
		fen := b.FEN()
		isWhite := (b.SideToMove == board.White)
		moveNum := (i / 2) + 1

		bestMove, bestScore := analyzer.analyzePosition(fen, *depth)

		// Calculate Centipawn Loss if played move differs
		cpLoss := 0
		classification := "🌟 Best Move"

		if playedMove != bestMove {
			// Find score after played move
			tempBoard := board.New()
			tempBoard.SetFEN(fen)
			list := &movegen.MoveList{}
			movegen.GenerateLegalMoves(tempBoard, list)

			for mIdx := 0; mIdx < list.Count; mIdx++ {
				m := list.Moves[mIdx]
				if m.String() == playedMove {
					var undo board.Undo
					tempBoard.MakeMove(m, &undo)
					_, playedScore := analyzer.analyzePosition(tempBoard.FEN(), *depth)
					cpLoss = bestScore - (-playedScore)
					if cpLoss < 0 {
						cpLoss = 0
					}
					break
				}
			}

			if cpLoss > 200 {
				classification = "🔴 BLUNDER"
				blunders++
			} else if cpLoss > 90 {
				classification = "🟠 MISTAKE"
				mistakes++
			} else if cpLoss > 30 {
				classification = "🟡 INACCURACY"
				inaccuracies++
			} else {
				classification = "🟢 Good Move"
			}
		}

		if isWhite {
			hyperionMoves++
			totalCPLoss += cpLoss
		}

		turnStr := "White"
		if !isWhite {
			turnStr = "Black"
		}

		fmt.Printf("Move %d (%s): Played %s | Best %s | Eval: %+d cp | Loss: %d cp [%s]\n",
			moveNum, turnStr, playedMove, bestMove, bestScore, cpLoss, classification)

		// Apply move to board
		list := &movegen.MoveList{}
		movegen.GenerateLegalMoves(b, list)
		for mIdx := 0; mIdx < list.Count; mIdx++ {
			m := list.Moves[mIdx]
			if m.String() == playedMove {
				var undo board.Undo
				b.MakeMove(m, &undo)
				break
			}
		}
	}

	avgCPLoss := 0
	if hyperionMoves > 0 {
		avgCPLoss = totalCPLoss / hyperionMoves
	}

	accuracy := max(0, 100-(avgCPLoss/2))

	fmt.Println("\n=========================================================")
	fmt.Println("                MATCH ANALYSIS REPORT                    ")
	fmt.Println("=========================================================")
	fmt.Printf("Hyperion Accuracy Score  : %d%%\n", accuracy)
	fmt.Printf("Average Centipawn Loss   : %d cp\n", avgCPLoss)
	fmt.Printf("Inaccuracies (30-90 cp)  : %d\n", inaccuracies)
	fmt.Printf("Mistakes (90-200 cp)     : %d\n", mistakes)
	fmt.Printf("Blunders (>200 cp)       : %d\n", blunders)
	fmt.Println("=========================================================")
}
