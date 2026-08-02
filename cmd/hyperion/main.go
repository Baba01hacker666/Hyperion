package main

import (
	"fmt"
	"hyperion/internal/uci"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		fmt.Println("Hyperion is running successfully.")
		return
	}

	// Start the UCI loop
	uci.Loop()
}
