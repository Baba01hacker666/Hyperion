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
	"strings"
	"time"
)

type Engine struct {
	cmd    *exec.Cmd
	stdin  io.Writer
	stdout *bufio.Reader
}

func newEngine(path string) (*Engine, error) {
	cmd := exec.Command(path)
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

	e := &Engine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}

	e.send("uci")
	e.expect("uciok")
	return e, nil
}

func (e *Engine) send(cmd string) {
	fmt.Fprintln(e.stdin, cmd)
}

func (e *Engine) expect(target string) string {
	for {
		line, err := e.stdout.ReadString('\n')
		if err != nil {
			return ""
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, target) || strings.Contains(line, target) {
			return line
		}
	}
}

func (e *Engine) getBestMove(fen string, depth int, movetime int) (string, string) {
	e.send(fmt.Sprintf("position fen %s", fen))
	if movetime > 0 {
		e.send(fmt.Sprintf("go movetime %d", movetime))
	} else {
		e.send(fmt.Sprintf("go depth %d", depth))
	}

	infoLine := ""
	bestMove := ""

	for {
		line, err := e.stdout.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "info") {
			infoLine = line
		}
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
			break
		}
	}
	return bestMove, infoLine
}

func (e *Engine) close() {
	e.send("quit")
	e.cmd.Wait()
}

func main() {
	numGames := flag.Int("games", 2, "Number of games to play")
	sfSkill := flag.Int("skill", 5, "Stockfish Skill Level (0-20), skill 5 ≈ 1600 Elo")
	hyperionDepth := flag.Int("hdepth", 0, "Hyperion search depth (0 = use movetime)")
	sfDepth := flag.Int("sfdepth", 0, "Stockfish search depth (0 = use movetime)")
	movetime := flag.Int("movetime", 1000, "Time per move in ms (used when depth=0, for blitz mode)")
	sfMovetime := flag.Int("sfmovetime", 300, "Stockfish time per move in ms")
	style := flag.String("style", "Blitz", "Hyperion play style: Blitz, Balanced, Gamble, Defense, Evil")
	flag.Parse()

	blitzMode := (*hyperionDepth == 0)

	fmt.Println("=========================================================")
	fmt.Println("         HYPERION CLI BENCHMARK & ARENA RUNNER           ")
	fmt.Println("=========================================================")
	if blitzMode {
		fmt.Printf("Hyperion BLITZ (%dms/move, Style=%s) vs Stockfish (Skill %d, %dms/move)\n",
			*movetime, *style, *sfSkill, *sfMovetime)
	} else {
		fmt.Printf("Hyperion (Depth %d) vs Stockfish (Skill %d, Depth %d)\n", *hyperionDepth, *sfSkill, *sfDepth)
	}
	fmt.Printf("Total Match Games: %d\n", *numGames)
	fmt.Println("---------------------------------------------------------")

	hyperionWins := 0
	sfWins := 0
	draws := 0

	for gameNum := 1; gameNum <= *numGames; gameNum++ {
		fmt.Printf("\n--- GAME %d of %d ---\n", gameNum, *numGames)

		hyperion, err := newEngine("./bin/hyperion")
		if err != nil {
			log.Fatalf("Failed to launch Hyperion: %v", err)
		}
		hyperion.send(fmt.Sprintf("setoption name Style value %s", *style))
		hyperion.send("setoption name Hash value 128")
		hyperion.send("setoption name Threads value 8")
		hyperion.send("ucinewgame")
		hyperion.send("isready")
		hyperion.expect("readyok")

		stockfish, err := newEngine("/usr/games/stockfish")
		if err != nil {
			log.Fatalf("Failed to launch Stockfish: %v", err)
		}
		stockfish.send(fmt.Sprintf("setoption name Skill Level value %d", *sfSkill))
		stockfish.send("setoption name Hash value 64")
		stockfish.send("ucinewgame")
		stockfish.send("isready")
		stockfish.expect("readyok")

		hyperionIsWhite := (gameNum%2 != 0)
		whiteName := "Hyperion"
		blackName := "Stockfish"
		if !hyperionIsWhite {
			whiteName = "Stockfish"
			blackName = "Hyperion"
		}
		fmt.Printf("White: %s | Black: %s\n", whiteName, blackName)

		b := board.New()
		b.SetFEN(board.StartFEN)
		moveCount := 1

		for moveCount <= 120 {
			currentFEN := b.FEN()
			isWhiteTurn := (b.SideToMove == board.White)
			isHyperionTurn := (isWhiteTurn && hyperionIsWhite) || (!isWhiteTurn && !hyperionIsWhite)

			engineName := "Stockfish"
			if isHyperionTurn {
				engineName = "Hyperion"
			}

			start := time.Now()
			var moveStr, info string
			if isHyperionTurn {
				mt := 0
				if blitzMode {
					mt = *movetime
				}
				moveStr, info = hyperion.getBestMove(currentFEN, *hyperionDepth, mt)
			} else {
				mt := 0
				if blitzMode {
					mt = *sfMovetime
				}
				moveStr, info = stockfish.getBestMove(currentFEN, *sfDepth, mt)
			}
			elapsed := time.Since(start)

			if moveStr == "" || moveStr == "(none)" || moveStr == "0000" {
				fmt.Printf("Game Over! %s has no legal moves (Checkmate/Stalemate).\n", engineName)
				if isHyperionTurn {
					sfWins++
				} else {
					hyperionWins++
				}
				break
			}

			// Apply move to Hyperion's board representation to get exact next FEN
			list := &movegen.MoveList{}
			movegen.GenerateLegalMoves(b, list)
			applied := false
			for i := 0; i < list.Count; i++ {
				m := list.Moves[i]
				if m.String() == moveStr {
					var undo board.Undo
					b.MakeMove(m, &undo)
					applied = true
					break
				}
			}

			if !applied {
				fmt.Printf("Illegal move %s attempted by %s! Game over.\n", moveStr, engineName)
				if isHyperionTurn {
					sfWins++
				} else {
					hyperionWins++
				}
				break
			}

			fmt.Printf("Move %d (%s): %s | Time: %v | Info: %s\n", moveCount, engineName, moveStr, elapsed.Round(time.Millisecond), info)
			fmt.Printf("  FEN: %s\n", b.FEN())

			moveCount++
		}

		if moveCount > 120 {
			fmt.Println("Game drawn by move limit.")
			draws++
		}

		hyperion.close()
		stockfish.close()
	}

	fmt.Println("\n=========================================================")
	fmt.Println("                    FINAL SCORECARD                      ")
	fmt.Println("=========================================================")
	fmt.Printf("Hyperion Wins : %d\n", hyperionWins)
	fmt.Printf("Stockfish Wins: %d\n", sfWins)
	fmt.Printf("Draws         : %d\n", draws)
	fmt.Println("=========================================================")
}
