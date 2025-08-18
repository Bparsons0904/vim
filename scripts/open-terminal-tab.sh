#!/bin/bash

# Open new terminal tab script
# Uses keystroke simulation to open a new tab in the current terminal

set -e

# Check if xdotool is available
if ! command -v xdotool &> /dev/null; then
    echo "Error: xdotool is required but not installed."
    echo "Install with: sudo apt install xdotool"
    exit 1
fi

# Get the currently focused window
WINDOW_ID=$(xdotool getwindowfocus)

# Check if the focused window is a terminal (by checking window class)
WINDOW_CLASS=$(xdotool getwindowname "$WINDOW_ID" 2>/dev/null || echo "unknown")

echo "Opening new terminal tab..."
echo "Current window: $WINDOW_CLASS"

# Send Ctrl+Shift+T to open new tab
xdotool key ctrl+shift+t

echo "✅ Sent Ctrl+Shift+T keystroke to open new tab"