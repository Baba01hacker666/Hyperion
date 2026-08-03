#!/usr/bin/env python3
"""
play_game.py — Play one UCI game: Hyperion vs Stockfish.
Prints result: "hyperion_win", "hyperion_loss", or "draw"
"""
import subprocess
import sys
import argparse
import time
import threading

def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("--hyperion", required=True)
    p.add_argument("--stockfish", required=True)
    p.add_argument("--hyperion-color", required=True, choices=["white", "black"])
    p.add_argument("--movetime", type=int, default=1000)
    p.add_argument("--skill", type=int, default=5)
    return p.parse_args()

class UCIEngine:
    def __init__(self, path, name="engine"):
        self.name = name
        self.proc = subprocess.Popen(
            [path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1
        )

    def send(self, cmd):
        self.proc.stdin.write(cmd + "\n")
        self.proc.stdin.flush()

    def readline(self, timeout=30):
        import select
        rlist, _, _ = select.select([self.proc.stdout], [], [], timeout)
        if rlist:
            return self.proc.stdout.readline().strip()
        return ""

    def read_until(self, keyword, timeout=60):
        lines = []
        deadline = time.time() + timeout
        while time.time() < deadline:
            line = self.readline(timeout=max(0.1, deadline - time.time()))
            if line:
                lines.append(line)
                if keyword in line:
                    return line, lines
        return "", lines

    def uci_init(self):
        self.send("uci")
        self.read_until("uciok", timeout=10)

    def new_game(self):
        self.send("ucinewgame")
        self.send("isready")
        self.read_until("readyok", timeout=10)

    def get_move(self, position_cmd, movetime):
        self.send(position_cmd)
        self.send(f"go movetime {movetime}")
        deadline = time.time() + movetime / 1000 + 10
        best_move = None
        while time.time() < deadline:
            line = self.readline(timeout=max(0.1, deadline - time.time()))
            if line.startswith("bestmove"):
                parts = line.split()
                best_move = parts[1] if len(parts) > 1 else None
                break
        return best_move

    def quit(self):
        try:
            self.send("quit")
            self.proc.wait(timeout=3)
        except:
            self.proc.kill()


def apply_move(board_moves, new_move):
    return board_moves + [new_move]

def is_checkmate_or_stalemate(moves_history, engine):
    """Check by asking engine to evaluate: if it returns no best move or 0000"""
    pos = "position startpos moves " + " ".join(moves_history)
    engine.send(pos)
    engine.send("go movetime 100")
    deadline = time.time() + 5
    while time.time() < deadline:
        line = engine.readline(timeout=2)
        if line.startswith("bestmove"):
            parts = line.split()
            mv = parts[1] if len(parts) > 1 else "0000"
            return mv in ("0000", "(none)", "")
    return False

def main():
    args = parse_args()

    hyperion = UCIEngine(args.hyperion, "Hyperion")
    stockfish = UCIEngine(args.stockfish, "Stockfish")

    hyperion.uci_init()
    # Enable blitz style for Hyperion
    hyperion.send("setoption name Style value Blitz")
    hyperion.send("setoption name Hash value 128")
    hyperion.new_game()

    stockfish.uci_init()
    stockfish.send(f"setoption name Skill Level value {args.skill}")
    stockfish.send("setoption name Hash value 128")
    stockfish.new_game()

    hyperion_is_white = (args.hyperion_color == "white")
    moves = []
    max_moves = 300  # Draw by move limit
    result = "draw"

    for move_num in range(max_moves):
        # Whose turn is it?
        white_to_move = (move_num % 2 == 0)
        hyperion_to_move = (white_to_move == hyperion_is_white)

        pos_cmd = "position startpos"
        if moves:
            pos_cmd += " moves " + " ".join(moves)

        active_engine = hyperion if hyperion_to_move else stockfish
        movetime = args.movetime if hyperion_to_move else 500

        mv = active_engine.get_move(pos_cmd, movetime)

        if mv is None or mv in ("0000", "(none)", ""):
            # No move available — the side to move is mated or stalemated
            if hyperion_to_move:
                result = "hyperion_loss"
            else:
                result = "hyperion_win"
            break

        moves.append(mv)

        # Check if position is terminal for next player
        # We detect this via the next engine returning no move
        # (handled at top of next iteration)

    else:
        result = "draw"

    hyperion.quit()
    stockfish.quit()
    print(result)

if __name__ == "__main__":
    main()
