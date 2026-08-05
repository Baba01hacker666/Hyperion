package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("/usr/games/stockfish")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	fmt.Fprintln(stdin, "uci")
	fmt.Fprintln(stdin, "isready")

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if scanner.Text() == "readyok" {
			break
		}
	}
	fmt.Println("Stockfish ready")
	stdin.Close()
	cmd.Wait()
}
