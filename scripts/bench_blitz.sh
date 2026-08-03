#!/usr/bin/env bash
# Benchmark: Hyperion blitz (1s/move) vs Stockfish Skill 5 (~1600 Elo)
# Runs N games alternating colors, reports results.

HYPERION="/root/Hyperion/hyperion_bin"
STOCKFISH="/usr/games/stockfish"
GAMES=${1:-10}
MOVETIME=1000  # 1 second per move

W=0; L=0; D=0
game=0

run_game() {
    local hyperion_color=$1  # "white" or "black"
    local game_num=$2

    # We drive the game via UCI pipes
    # Returns: "hyperion_win", "hyperion_loss", "draw"
    python3 /root/Hyperion/scripts/play_game.py \
        --hyperion "$HYPERION" \
        --stockfish "$STOCKFISH" \
        --hyperion-color "$hyperion_color" \
        --movetime "$MOVETIME" \
        --skill 5
}

echo "========================================"
echo " Hyperion Blitz Benchmark vs SF Skill 5"
echo " (~1600 Elo) — $GAMES games, 1s/move"
echo "========================================"
echo ""

for ((i=1; i<=GAMES; i++)); do
    if (( i % 2 == 1 )); then
        COLOR="white"
    else
        COLOR="black"
    fi

    echo -n "Game $i/$GAMES (Hyperion=$COLOR): "
    RESULT=$(run_game "$COLOR" "$i")

    case "$RESULT" in
        "hyperion_win")
            echo "✅ WIN"
            ((W++))
            ;;
        "hyperion_loss")
            echo "❌ LOSS"
            ((L++))
            ;;
        "draw")
            echo "🤝 DRAW"
            ((D++))
            ;;
        *)
            echo "❓ UNKNOWN ($RESULT)"
            ;;
    esac
done

TOTAL=$((W + L + D))
SCORE=$(echo "scale=1; ($W + $D * 0.5) * 100 / $TOTAL" | bc)

echo ""
echo "========================================"
echo " RESULTS: +$W =$D -$L  (Score: $SCORE%)"
echo "========================================"
