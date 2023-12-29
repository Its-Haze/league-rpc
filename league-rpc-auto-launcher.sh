#!/bin/bash

# Configurations
REPO="Its-Haze/league-rpc-linux"
EXECUTABLE="league_rpc_linux"
INSTALL_DIR="$HOME/.league-rpc-linux"
EXECUTABLE_PATH="$INSTALL_DIR/$EXECUTABLE"
LOG_FILE="$INSTALL_DIR/update_log.txt"
AUTO_INSTALL=true # Set to (false) to not automatically install the latest version.

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check for required programs: curl and grep
for required_command in curl grep; do
    if ! command -v $required_command &>/dev/null; then
        log_message "Error: $required_command is not installed. Please install it and try again."
        exit 1
    fi
done

# Create the installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Function to download and update the executable
update_executable() {
    log_message "Installing the executable..."
    DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" |
        grep "\"browser_download_url\": \".*$EXECUTABLE\"" |
        cut -d '"' -f 4)

    if [ -z "$DOWNLOAD_URL" ]; then
        log_message "Failed to fetch release info."
        exit 1
    fi

    log_message "Downloading from $DOWNLOAD_URL"
    if curl -Ls "$DOWNLOAD_URL" -o "$EXECUTABLE_PATH"; then
        chmod +x "$EXECUTABLE_PATH" || log_message "Failed to set executable permission."
        log_message "Update successful."
    else
        log_message "Failed to download the executable."
        exit 1
    fi
}

# Process command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
    --auto-install)
        AUTO_INSTALL=true
        shift
        ;;
    *)
        log_message "Unknown option: $1"
        exit 1
        ;;
    esac
done

# Perform auto-install if enabled
if [[ "$AUTO_INSTALL" == true ]]; then
    update_executable
fi

# Check if the executable exists and is executable
if [[ ! -x "$EXECUTABLE_PATH" ]]; then
    log_message "ERROR: Executable not found at $EXECUTABLE_PATH"
    log_message "INFO: If you want to quickly fetch the latest version, use --auto-install."
    exit 1
fi

# Execute the program
log_message "Executing the program..."
exec "$EXECUTABLE_PATH" --wait-for-league -1 --wait-for-discord -1
