package persona

import (
	"encoding/json"
	"os"
)

// Persona defines the play style, search heuristics, and time management parameters of an opponent or style.
type Persona struct {
	Name             string  `json:"name"`
	EvalStyle        string  `json:"eval_style"`        // "Blitz", "Evil", "Balanced", "Defense", "Gamble"
	TimeAggression   float64 `json:"time_aggression"`   // Multiplier for SoftTime/HardTime. >1 = thinks longer.
	LMRBase          float64 `json:"lmr_base"`          // Base coefficient for LMR (e.g., 0.7844)
	LMRMultiplier    float64 `json:"lmr_multiplier"`    // Divisor/Multiplier for LMR log(d)*log(m) (e.g., 2.4696)
	AspirationWindow int     `json:"aspiration_window"` // Starting alpha-beta window size (e.g., 25)
	MobilityWeight   float64 `json:"mobility_weight"`   // Custom multiplier for mobility evaluation
	KingSafetyWeight float64 `json:"king_safety_weight"` // Custom multiplier for King Safety
}

var ActivePersona *Persona = nil

// LoadPersona reads a .persona JSON file and sets it as the active persona.
func LoadPersona(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	var p Persona
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	ActivePersona = &p
	return nil
}

// DefaultPersona returns the standard Hyperion 1.1 parameters.
func DefaultPersona() *Persona {
	return &Persona{
		Name:             "Hyperion-Default",
		EvalStyle:        "Blitz",
		TimeAggression:   1.0,
		LMRBase:          0.7844,
		LMRMultiplier:    2.4696,
		AspirationWindow: 25,
		MobilityWeight:   1.0,
		KingSafetyWeight: 1.0,
	}
}
