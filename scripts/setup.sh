#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# netmon-go Setup Script
# ==============================================================================
# Interactive setup & installation script for netmon-go.
# Usage:
#   curl -sSL https://raw.githubusercontent.com/vernak2539/netmon-go/main-go/scripts/setup.sh | bash
# ==============================================================================

BOLD="\033[1m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
BLUE="\033[0;34m"
RED="\033[0;31m"
RESET="\033[0m"

info() {
    echo -e "${BLUE}${BOLD}[INFO]${RESET} $*"
}

success() {
    echo -e "${GREEN}${BOLD}[SUCCESS]${RESET} $*"
}

warn() {
    echo -e "${YELLOW}${BOLD}[WARNING]${RESET} $*"
}

error() {
    echo -e "${RED}${BOLD}[ERROR]${RESET} $*"
}

echo -e "${BOLD}"
echo "=================================================================="
echo "                   netmon-go Setup & Installer                   "
echo "=================================================================="
echo -e "${RESET}"

# ------------------------------------------------------------------------------
# 1. Detect OS & Architecture
# ------------------------------------------------------------------------------
info "Detecting operating system and CPU architecture..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    linux)
        TARGET_OS="linux"
        ;;
    darwin)
        TARGET_OS="darwin"
        ;;
    *)
        error "Unsupported operating system: $OS. netmon-go requires Linux or macOS."
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        TARGET_ARCH="amd64"
        ;;
    aarch64|arm64)
        TARGET_ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH."
        exit 1
        ;;
esac

ASSET_NAME="netmon-go-${TARGET_OS}-${TARGET_ARCH}"
info "Detected platform: ${TARGET_OS}/${TARGET_ARCH} (Binary asset: ${ASSET_NAME})"

# ------------------------------------------------------------------------------
# 2. Check & Install System Prerequisites (nmap, curl, etc.)
# ------------------------------------------------------------------------------
info "Checking system prerequisites..."

if ! command -v curl >/dev/null 2>&1; then
    error "'curl' is required but not installed."
    exit 1
fi

if ! command -v nmap >/dev/null 2>&1; then
    warn "'nmap' is not installed (required for LAN device scanning)."
    if [ "$TARGET_OS" = "linux" ]; then
        if command -v apt-get >/dev/null 2>&1; then
            info "Attempting to install nmap via apt-get..."
            sudo apt-get update && sudo apt-get install -y nmap
        else
            warn "Please install nmap using your package manager (e.g. yum, pacman, etc.)."
        fi
    elif [ "$TARGET_OS" = "darwin" ]; then
        if command -v brew >/dev/null 2>&1; then
            info "Attempting to install nmap via Homebrew..."
            brew install nmap
        else
            warn "Please install nmap via Homebrew ('brew install nmap')."
        fi
    fi
fi

if command -v nmap >/dev/null 2>&1; then
    NMAP_PATH="$(command -v nmap)"
    success "Found nmap at: $NMAP_PATH"

    # Configure passwordless sudo for nmap if on Linux and sudo is available
    if [ "$TARGET_OS" = "linux" ] && command -v sudo >/dev/null 2>&1; then
        SUDOERS_FILE="/etc/sudoers.d/netmon-nmap"
        if [ ! -f "$SUDOERS_FILE" ]; then
            info "Configuring passwordless sudo access for nmap ($SUDOERS_FILE)..."
            USER_NAME="$(whoami)"
            echo "$USER_NAME ALL=(root) NOPASSWD: $NMAP_PATH" | sudo tee "$SUDOERS_FILE" >/dev/null
            sudo chmod 440 "$SUDOERS_FILE"
            success "Passwordless sudo configured for nmap."
        fi
    fi
else
    error "nmap installation could not be completed. LAN device discovery will fall back to TCP."
fi

# ------------------------------------------------------------------------------
# 3. Download Latest Binary Release
# ------------------------------------------------------------------------------
info "Fetching latest netmon-go release from GitHub..."

LATEST_RELEASE_JSON="$(curl -sSL https://api.github.com/repos/vernak2539/netmon-go/releases/latest)"
LATEST_TAG="$(echo "$LATEST_RELEASE_JSON" | grep '"tag_name":' | head -n1 | sed -E 's/.*"([^"]+)".*/\1/')"

if [ -z "$LATEST_TAG" ]; then
    warn "Could not fetch latest release tag via GitHub API. Defaulting to 'latest' download URL."
    DOWNLOAD_URL="https://github.com/vernak2539/netmon-go/releases/latest/download/${ASSET_NAME}"
else
    info "Latest release version: ${LATEST_TAG}"
    DOWNLOAD_URL="https://github.com/vernak2539/netmon-go/releases/download/${LATEST_TAG}/${ASSET_NAME}"
fi

INSTALL_DIR="/usr/local/bin"
TARGET_BIN="${INSTALL_DIR}/netmon"

info "Downloading binary from ${DOWNLOAD_URL}..."

TMP_BIN="$(mktemp)"
if curl -sSL -o "$TMP_BIN" "$DOWNLOAD_URL"; then
    chmod +x "$TMP_BIN"
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_BIN" "$TARGET_BIN"
    else
        info "Elevated permissions required to install to ${INSTALL_DIR}."
        sudo mv "$TMP_BIN" "$TARGET_BIN"
    fi
    success "netmon-go installed successfully to ${TARGET_BIN}"
else
    error "Failed to download release binary from ${DOWNLOAD_URL}."
    exit 1
fi

# ------------------------------------------------------------------------------
# 4. Interactive Configuration (.env) Setup
# ------------------------------------------------------------------------------
CONFIG_DIR="/etc/netmon"
ENV_FILE="${CONFIG_DIR}/.env"

if [ ! -w "/etc" ]; then
    ENV_FILE="./.env"
    CONFIG_DIR="."
fi

info "Configuring environment settings (${ENV_FILE})..."

if [ -t 0 ]; then

    echo -e "\n${BOLD}Telegram Setup Instructions:${RESET}"
    echo "  - How to create a Telegram Bot:"
    echo "    https://medium.com/@a.kotchnev/how-to-create-a-telegram-bot-in-5-minutes-botfather-2fdbf1da2627"
    echo "  - How to get your Telegram Chat ID:"
    echo "    https://dev.to/marcotwzrd/how-to-get-a-telegram-chatid-in-2026-3-methods-that-actually-work-36g5"
    echo ""

    read -r -p "Enter Telegram Bot Token (TG_BOT_TOKEN): " INPUT_TG_BOT_TOKEN
    read -r -p "Enter Telegram Chat ID (TG_CHAT_ID): " INPUT_TG_CHAT_ID
    read -r -p "Enter OpenAI API Key [optional, press Enter to skip]: " INPUT_AI_API_KEY
    read -r -p "Enter SQLite Database Path [default: /var/lib/netmon/metrics.sql]: " INPUT_DB_PATH

    DB_PATH="${INPUT_DB_PATH:-/var/lib/netmon/metrics.sql}"
    mkdir -p "$(dirname "$DB_PATH")" 2>/dev/null || true

    if [ -n "$CONFIG_DIR" ] && [ "$CONFIG_DIR" != "." ]; then
        sudo mkdir -p "$CONFIG_DIR"
        sudo tee "$ENV_FILE" >/dev/null <<EOF
TG_BOT_TOKEN=${INPUT_TG_BOT_TOKEN}
TG_CHAT_ID=${INPUT_TG_CHAT_ID}
AI_API_KEY=${INPUT_AI_API_KEY}
AI_MODEL=gpt-4o-mini
AI_BASE_URL=https://api.openai.com/v1
DB_PATH=${DB_PATH}
EOF
    else
        cat <<EOF > "$ENV_FILE"
TG_BOT_TOKEN=${INPUT_TG_BOT_TOKEN}
TG_CHAT_ID=${INPUT_TG_CHAT_ID}
AI_API_KEY=${INPUT_AI_API_KEY}
AI_MODEL=gpt-4o-mini
AI_BASE_URL=https://api.openai.com/v1
DB_PATH=${DB_PATH}
EOF
    fi

    success "Environment configuration saved to ${ENV_FILE}"
else
    warn "Non-interactive shell detected. Skipping interactive configuration."
    warn "Please ensure environment variables (TG_BOT_TOKEN, TG_CHAT_ID, etc.) or a .env file are configured."
fi

# ------------------------------------------------------------------------------
# 5. systemd Service Setup (Linux)
# ------------------------------------------------------------------------------
if [ "$TARGET_OS" = "linux" ] && command -v systemctl >/dev/null 2>&1; then
    info "Setting up systemd service (netmon.service)..."
    SERVICE_FILE="/etc/systemd/system/netmon.service"

    SYSTEMD_UNIT="[Unit]
Description=netmon-go Network Monitoring Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${TARGET_BIN} --env ${ENV_FILE}
Restart=always
RestartSec=10
User=root

[Install]
WantedBy=multi-user.target"

    echo "$SYSTEMD_UNIT" | sudo tee "$SERVICE_FILE" >/dev/null
    sudo systemctl daemon-reload
    sudo systemctl enable --now netmon.service
    success "netmon.service successfully enabled and started!"
    info "Check service status: sudo systemctl status netmon"
else
    info "Setup complete! Run netmon manually via:"
    echo "  netmon --env ${ENV_FILE}"
fi

echo -e "\n${GREEN}${BOLD}🎉 netmon-go installation & setup complete!${RESET}\n"
