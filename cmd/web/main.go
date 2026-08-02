package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
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

func main() {
	http.Handle("/", http.FileServer(http.Dir("cmd/web/static")))
	http.HandleFunc("/api/move", handleMove)

	fmt.Println("Hyperion Web Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
