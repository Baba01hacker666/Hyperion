#!/usr/bin/env python3
"""
Robust UCI match runner for Hyperion vs Stockfish benchmarks.
Usage: python3 scripts/run_match.py [num_games] [skill_level]
"""
import subprocess, time, sys, os

def make_engine(path, options=None):
    proc = subprocess.Popen(
        [path],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        text=True,
        bufsize=1
    )

    def send(cmd):
        proc.stdin.write(cmd + "\n")
        proc.stdin.flush()

    def recv(timeout=15):
        import select
        r, _, _ = select.select([proc.stdout], [], [], timeout)
        if r:
            return proc.stdout.readline().strip()
        return ""

    def wait_for(keyword, timeout=15):
        deadline = time.time() + timeout
        while time.time() < deadline:
            line = recv(max(0.05, deadline - time.time()))
            if keyword in line:
                return True
        return False

    # Init UCI
    send("uci")
    wait_for("uciok", timeout=10)

    if options:
        for k, v in options.items():
            send(f"setoption name {k} value {v}")

    send("ucinewgame")
    send("isready")
    wait_for("readyok", timeout=10)

    def get_move(moves_list, movetime_ms):
        pos = "position startpos"
        if moves_list:
            pos += " moves " + " ".join(moves_list)
        send(pos)
        send(f"go movetime {movetime_ms}")
        deadline = time.time() + movetime_ms / 1000.0 + 15
        while time.time() < deadline:
            line = recv(max(0.05, deadline - time.time()))
            if line.startswith("bestmove"):
                parts = line.split()
                mv = parts[1] if len(parts) > 1 else ""
                if mv and mv not in ("0000", "(none)"):
                    return mv
                return None
        return None

    def new_game():
        send("ucinewgame")
        send("isready")
        wait_for("readyok", timeout=10)

    def quit():
        try:
            send("quit")
            proc.wait(timeout=3)
        except Exception:
            proc.kill()

    return {"send": send, "recv": recv, "get_move": get_move, "new_game": new_game, "quit": quit, "proc": proc}


def play_one_game(hyperion_path, sf_path, hyperion_is_white, movetime_ms, sf_skill):
    h = make_engine(hyperion_path, {"Style": "Blitz", "Hash": "128"})
    sf = make_engine(sf_path, {"Skill Level": str(sf_skill), "Hash": "64"})

    moves = []
    result = "draw"

    for half_move in range(400):  # 200 moves max
        white_to_move = (half_move % 2 == 0)
        hyperion_turn = (white_to_move == hyperion_is_white)

        engine = h if hyperion_turn else sf
        mt = movetime_ms if hyperion_turn else 200  # SF plays fast

        mv = engine["get_move"](moves, mt)

        if mv is None:
            # No legal move → mated or stalemate
            result = "hyperion_loss" if hyperion_turn else "hyperion_win"
            break

        moves.append(mv)

    h["quit"]()
    sf["quit"]()
    return result, len(moves)


def main():
    num_games = int(sys.argv[1]) if len(sys.argv) > 1 else 10
    sf_skill  = int(sys.argv[2]) if len(sys.argv) > 2 else 5   # skill 5 ≈ 1600 Elo
    movetime  = int(sys.argv[3]) if len(sys.argv) > 3 else 1000  # 1s/move

    hyperion = "/root/Hyperion/hyperion_bin"
    stockfish = "/usr/games/stockfish"

    print(f"{'='*50}")
    print(f"  Hyperion Blitz vs Stockfish (Skill {sf_skill} ≈ 1600 Elo)")
    print(f"  {num_games} games | {movetime}ms/move for Hyperion")
    print(f"{'='*50}")

    W = L = D = 0
    for i in range(1, num_games + 1):
        h_white = (i % 2 == 1)
        color_str = "White" if h_white else "Black"
        print(f"  Game {i:2d}/{num_games} (Hyperion={color_str}): ", end="", flush=True)

        result, nmoves = play_one_game(hyperion, stockfish, h_white, movetime, sf_skill)

        if result == "hyperion_win":
            W += 1; sym = "✅ WIN"
        elif result == "hyperion_loss":
            L += 1; sym = "❌ LOSS"
        else:
            D += 1; sym = "🤝 DRAW"

        print(f"{sym}  ({nmoves} half-moves)", flush=True)

    total = W + L + D
    score = (W + D * 0.5) * 100 / total if total else 0
    print(f"\n{'='*50}")
    print(f"  FINAL: +{W} ={D} -{L}   Score: {score:.1f}%")
    if score >= 60:
        print(f"  🔥 Hyperion is dominating ~1600 Elo!")
    elif score >= 50:
        print(f"  ✅ Hyperion is competitive with ~1600 Elo!")
    elif score >= 40:
        print(f"  ⚠️  Hyperion is close but below 1600 Elo target.")
    else:
        print(f"  ❌ Hyperion needs more work to reach 1600 Elo.")
    print(f"{'='*50}")


if __name__ == "__main__":
    main()
