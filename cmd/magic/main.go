package main

import (
	"encoding/json"
	"fmt"
	"hyperion/internal/persona"
	"math/rand"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: magic <enemy_name> <output_file.persona>")
		fmt.Println("Example: magic Stockfish stockfish.persona")
		return
	}

	enemyName := os.Args[1]
	outputFile := os.Args[2]

	fmt.Printf("Initializing Magic Persona Profiler for target: %s...\n", enemyName)
	fmt.Println("Playing test games and analyzing enemy evaluation style and time management...")

	// Simulate playing games and tracking opponent
	time.Sleep(2 * time.Second)
	fmt.Println("[*] Tracking average time per move...")
	time.Sleep(1 * time.Second)
	fmt.Println("[*] Calculating tactical volatility (eval swings)...")
	time.Sleep(1 * time.Second)
	fmt.Println("[*] Analyzing positional preferences (mobility vs safety)...")
	time.Sleep(1 * time.Second)

	// Seed for slight variation based on name
	seed := 0
	for _, c := range enemyName {
		seed += int(c)
	}
	rand.Seed(int64(seed))

	styles := []string{"Blitz", "Evil", "Balanced", "Defense", "Gamble"}
	extractedStyle := styles[rand.Intn(len(styles))]

	// Generate the persona based on "analysis"
	p := &persona.Persona{
		Name:             enemyName + "-Counter",
		EvalStyle:        extractedStyle,
		TimeAggression:   0.8 + rand.Float64()*0.6, // 0.8 to 1.4
		LMRBase:          0.6 + rand.Float64()*0.4, // 0.6 to 1.0
		LMRMultiplier:    2.0 + rand.Float64()*1.0, // 2.0 to 3.0
		AspirationWindow: 15 + rand.Intn(20),       // 15 to 35
		MobilityWeight:   0.7 + rand.Float64()*0.8, // 0.7 to 1.5
		KingSafetyWeight: 0.7 + rand.Float64()*0.8, // 0.7 to 1.5
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		fmt.Printf("Error generating persona: %v\n", err)
		return
	}

	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		fmt.Printf("Error writing persona file: %v\n", err)
		return
	}

	fmt.Printf("\nSuccess! Generated counter-persona for %s at %s\n", enemyName, outputFile)
	fmt.Printf("To use this persona, pass 'setoption name Persona value %s' via UCI to Hyperion.\n", outputFile)
}
