#!/bin/bash

# Open multiple terminal tabs script
# Opens 5 additional tabs assuming 1 BillyWu tab is already open

set -e

# Configuration
# Default project paths (can be overridden with environment variables)
DEFAULT_BILLYWU_PATH="${HOME}/Development/billywu/BillyWu"
DEFAULT_CLAUDE_PATH="${HOME}/Development/billywu/claude"
DEFAULT_GEMINI_PATH="${HOME}/Development/billywu/gemini"

# Use environment variables if set, otherwise use defaults
PROJECTS=(
    "${BILLYWU_PATH:-$DEFAULT_BILLYWU_PATH}"
    "${CLAUDE_PATH:-$DEFAULT_CLAUDE_PATH}"
    "${GEMINI_PATH:-$DEFAULT_GEMINI_PATH}"
)

# Parse arguments
DELAY=${1:-1}

# Check if xdotool is available
if ! command -v xdotool &> /dev/null; then
    echo "Error: xdotool is required but not installed."
    echo "Install with: sudo apt install xdotool"
    exit 1
fi

echo "Opening 5 additional terminal tabs..."
echo "Delay between tabs: ${DELAY}s"
echo "Assuming current tab is already in BillyWu directory"

# Function to open a tab and run a command
open_tab_with_command() {
    local tab_name="$1"
    local project_dir="$2"
    
    echo "Opening $tab_name..."
    
    # Send Ctrl+Shift+T to open new tab
    xdotool key ctrl+shift+t
    sleep $DELAY
    
    # Navigate to project directory
    if [[ -n "$project_dir" && -d "$project_dir" ]]; then
        echo "  → Navigating to: $project_dir"
        xdotool type "\\cd $project_dir"
        xdotool key Return
        sleep 0.5
    fi
}

# Open 5 additional tabs to complete the 6-tab setup
echo "Tab 1: Already open (current BillyWu tab)"

# Tab 2: BillyWu-2 (stay in current directory - duplicate current tab)
echo "Opening BillyWu-2 (staying in current directory)..."
xdotool key ctrl+shift+t
sleep $DELAY

# Tabs 3-4: claude
open_tab_with_command "claude-1" "${PROJECTS[1]}"
open_tab_with_command "claude-2" "${PROJECTS[1]}"

# Tabs 5-6: gemini  
open_tab_with_command "gemini-1" "${PROJECTS[2]}"
open_tab_with_command "gemini-2" "${PROJECTS[2]}"

echo "✅ Opened 5 additional terminal tabs successfully!"
echo ""
echo "Final tab layout:"
echo "  Tab 1: BillyWu (original tab)"
echo "  Tab 2: BillyWu-2"
echo "  Tab 3: claude-1"
echo "  Tab 4: claude-2"
echo "  Tab 5: gemini-1"
echo "  Tab 6: gemini-2"
echo ""
echo "Usage:"
echo "  ./scripts/open-multiple-tabs.sh     # Default 1s delay"
echo "  ./scripts/open-multiple-tabs.sh 2   # 2s delay between tabs"