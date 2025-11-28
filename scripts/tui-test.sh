#!/bin/bash
# TUI testing helper for AI-driven development
# Usage: ./scripts/tui-test.sh [command]
#
# Commands:
#   start       - Start app in tmux session (default)
#   view        - Capture and display current state
#   keys "..."  - Send keystrokes (e.g., keys "jjjl")
#   type "..."  - Type text (e.g., type "/bug")
#   enter       - Press Enter
#   quit        - Send 'q' and clean up
#   clean       - Kill session without sending quit

SESSION="abacus-tui-test"
WIDTH="${TUI_WIDTH:-120}"
HEIGHT="${TUI_HEIGHT:-40}"

start() {
    tmux kill-session -t "$SESSION" 2>/dev/null
    make build >/dev/null 2>&1
    tmux new-session -d -s "$SESSION" -x "$WIDTH" -y "$HEIGHT" './abacus'
    sleep 2
    echo "Started abacus in tmux session '$SESSION' (${WIDTH}x${HEIGHT})"
    view
}

view() {
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    tmux capture-pane -t "$SESSION" -p -e
}

view_plain() {
    # Strip ANSI codes for cleaner output
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    tmux capture-pane -t "$SESSION" -p | sed 's/\x1b\[[0-9;]*m//g'
}

keys() {
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    for key in $(echo "$1" | grep -o .); do
        tmux send-keys -t "$SESSION" "$key"
        sleep 0.15
    done
    sleep 0.3
    view
}

type_text() {
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    tmux send-keys -t "$SESSION" "$1"
    sleep 0.3
    view
}

send_enter() {
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    tmux send-keys -t "$SESSION" Enter
    sleep 0.3
    view
}

send_escape() {
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No session running. Use: $0 start"
        exit 1
    fi
    tmux send-keys -t "$SESSION" Escape
    sleep 0.3
    view
}

quit_app() {
    if tmux has-session -t "$SESSION" 2>/dev/null; then
        tmux send-keys -t "$SESSION" 'q'
        sleep 0.5
        tmux kill-session -t "$SESSION" 2>/dev/null
        echo "Session cleaned up"
    else
        echo "No session running"
    fi
}

clean() {
    tmux kill-session -t "$SESSION" 2>/dev/null
    echo "Session killed"
}

case "${1:-start}" in
    start)  start ;;
    view)   view ;;
    plain)  view_plain ;;
    keys)   keys "$2" ;;
    type)   type_text "$2" ;;
    enter)  send_enter ;;
    escape) send_escape ;;
    quit)   quit_app ;;
    clean)  clean ;;
    *)
        echo "Usage: $0 {start|view|plain|keys|type|enter|escape|quit|clean}"
        echo ""
        echo "Examples:"
        echo "  $0 start           # Start app"
        echo "  $0 keys 'jjjl'     # Navigate down 3x, expand"
        echo "  $0 type '/bug'     # Type search text"
        echo "  $0 enter           # Press Enter"
        echo "  $0 view            # Capture current state"
        echo "  $0 quit            # Quit app and clean up"
        exit 1
        ;;
esac
