package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hyperion/internal/board"
	"hyperion/internal/movegen"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func handleMove(w http.ResponseWriter, r *http.Request) {
	fen := r.URL.Query().Get("fen")
	if fen == "" {
		http.Error(w, "missing fen", http.StatusBadRequest)
		return
	}

	style := r.URL.Query().Get("style")
	if style == "" {
		style = "normal"
	}

	// Calculate absolute path to the binary to avoid working directory issues
	binaryPath, _ := filepath.Abs("bin/hyperion")

	// Command to execute
	input := fmt.Sprintf("setoption name Style value %s\nposition fen %s\ngo depth 8\nquit\n", style, fen)

	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader(input)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Printf("Hyperion error: %v", err)
		http.Error(w, "engine error", http.StatusInternalServerError)
		return
	}

	output := out.String()
	bestMove := ""
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bestmove": bestMove})
}

type MoveAnalysis struct {
	MoveNum        int    `json:"move_num"`
	Color          string `json:"color"`
	Played         string `json:"played"`
	Best           string `json:"best"`
	Eval           int    `json:"eval"`
	CPLoss         int    `json:"cp_loss"`
	Classification string `json:"classification"`
}

type AnalysisReport struct {
	Accuracy     int            `json:"accuracy"`
	AvgCPLoss    int            `json:"avg_cp_loss"`
	Inaccuracies int            `json:"inaccuracies"`
	Mistakes     int            `json:"mistakes"`
	Blunders     int            `json:"blunders"`
	Details      []MoveAnalysis `json:"details"`
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	movesStr := r.URL.Query().Get("moves")
	if movesStr == "" {
		http.Error(w, "missing moves parameter", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("stockfish")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		http.Error(w, "stockfish error", http.StatusInternalServerError)
		return
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "stockfish pipe error", http.StatusInternalServerError)
		return
	}
	if err := cmd.Start(); err != nil {
		http.Error(w, "stockfish start error", http.StatusInternalServerError)
		return
	}

	reader := bufio.NewReader(stdoutPipe)
	send := func(c string) { fmt.Fprintln(stdin, c) }

	send("uci")
	send("setoption name Skill Level value 20")
	send("isready")

	// Read until readyok
	for {
		line, _ := reader.ReadString('\n')
		if strings.Contains(line, "readyok") {
			break
		}
	}

	analyzePos := func(fen string) (string, int) {
		send(fmt.Sprintf("position fen %s", fen))
		send("go depth 10")
		bestMove := ""
		scoreCP := 0
		for {
			line, err := reader.ReadString('\n')
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

	b := board.New()
	b.SetFEN(board.StartFEN)
	moveList := strings.Fields(movesStr)

	report := AnalysisReport{
		Details: []MoveAnalysis{},
	}
	totalLoss := 0
	evaluatedCount := 0

	for i, playedMove := range moveList {
		fen := b.FEN()
		isWhite := (b.SideToMove == board.White)
		moveNum := (i / 2) + 1

		bestMove, bestScore := analyzePos(fen)
		cpLoss := 0
		classification := "🌟 Best Move"

		if playedMove != bestMove {
			tempBoard := board.New()
			tempBoard.SetFEN(fen)
			list := &movegen.MoveList{}
			movegen.GenerateLegalMoves(tempBoard, list)

			for mIdx := 0; mIdx < list.Count; mIdx++ {
				m := list.Moves[mIdx]
				if m.String() == playedMove {
					var undo board.Undo
					tempBoard.MakeMove(m, &undo)
					_, playedScore := analyzePos(tempBoard.FEN())
					cpLoss = bestScore - (-playedScore)
					if cpLoss < 0 {
						cpLoss = 0
					}
					break
				}
			}

			if cpLoss > 200 {
				classification = "🔴 BLUNDER"
				report.Blunders++
			} else if cpLoss > 90 {
				classification = "🟠 MISTAKE"
				report.Mistakes++
			} else if cpLoss > 30 {
				classification = "🟡 INACCURACY"
				report.Inaccuracies++
			} else {
				classification = "🟢 Good Move"
			}
		}

		turnStr := "White"
		if !isWhite {
			turnStr = "Black"
		}

		report.Details = append(report.Details, MoveAnalysis{
			MoveNum:        moveNum,
			Color:          turnStr,
			Played:         playedMove,
			Best:           bestMove,
			Eval:           bestScore,
			CPLoss:         cpLoss,
			Classification: classification,
		})

		totalLoss += cpLoss
		evaluatedCount++

		// Advance board
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

	send("quit")
	cmd.Wait()

	if evaluatedCount > 0 {
		report.AvgCPLoss = totalLoss / evaluatedCount
	}
	report.Accuracy = max(0, 100-(report.AvgCPLoss/2))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("cmd/web/static")))
	http.HandleFunc("/api/move", handleMove)
	http.HandleFunc("/api/analyze", handleAnalyze)

	fmt.Println("Hyperion Web Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
