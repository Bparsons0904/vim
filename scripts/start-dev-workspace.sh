#!/bin/bash

# BillyWu Development Workspace Startup Script
# Creates detached zellij session with tilt services and opens terminal tabs

set -e

# Configuration
SESSION_NAME="billywu-tilt-services"
LAYOUT_NAME="billywu-workspace"

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

# Check if required directories exist
echo "Checking project directories..."
for project in "${PROJECTS[@]}"; do
    if [[ ! -d "$project" ]]; then
        echo "Error: Project directory does not exist: $project"
        exit 1
    fi
done

# Check if xdotool is available for tab automation
if ! command -v xdotool &> /dev/null; then
    echo "Warning: xdotool not found. Terminal tabs won't be opened automatically."
    echo "Install with: sudo apt install xdotool"
    SKIP_TABS=true
else
    SKIP_TABS=false
fi

# Clean up any existing session (including dead ones)
if zellij list-sessions 2>/dev/null | grep -q "$SESSION_NAME"; then
    echo "Killing existing session: $SESSION_NAME"
    zellij kill-session "$SESSION_NAME" 2>/dev/null || true
fi

# Also delete any dead sessions with the same name
echo "Cleaning up any dead sessions..."
zellij delete-session "$SESSION_NAME" 2>/dev/null || true

echo "Starting detached zellij session with tilt services..."

# Start detached zellij session with our layout (force detached mode)
zellij -s "$SESSION_NAME" --layout "$LAYOUT_NAME" -d &

sleep 2

# Wait for session to be ready
while ! zellij list-sessions 2>/dev/null | grep -q "$SESSION_NAME"; do
    echo "Waiting for session to be ready..."
    sleep 1
done

echo "✅ Detached zellij session created successfully!"

# Open terminal tabs if xdotool is available
if [[ "$SKIP_TABS" == "false" ]]; then
    echo ""
    echo "Opening 5 additional terminal tabs..."
    echo "Delay between tabs: ${DELAY}s"
    echo "Assuming current tab is already in BillyWu directory"
    
    # Function to open a tab and navigate to directory
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
    
    # Tab 2: BillyWu-2 (stay in current directory)
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
fi

echo ""
echo "🚀 Development workspace ready!"
echo "Session name: $SESSION_NAME (detached)"
echo "Layout: $LAYOUT_NAME"
echo ""
echo "Services running in background:"
echo "  - BillyWu: tilt up (default port)"
echo "  - claude: tilt up --port 10351"
echo "  - gemini: tilt up --port 10352"
echo ""
echo "Useful commands:"
echo "  - View tilt logs: zellij attach $SESSION_NAME"
echo "  - Kill session: zellij kill-session $SESSION_NAME"
echo "  - List sessions: zellij list-sessions"
echo ""
echo "Usage:"
echo "  ./scripts/start-dev-workspace.sh     # Default 1s delay"
echo "  ./scripts/start-dev-workspace.sh 2   # 2s delay between tabs"
echo ""
echo "Environment variables (optional):"
echo "  BILLYWU_PATH - Path to BillyWu project (default: \$HOME/Development/billywu/BillyWu)"
echo "  CLAUDE_PATH  - Path to claude project (default: \$HOME/Development/billywu/claude)"
echo "  GEMINI_PATH  - Path to gemini project (default: \$HOME/Development/billywu/gemini)"