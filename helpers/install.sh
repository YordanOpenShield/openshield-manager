#!/usr/bin/env bash

set -e

AGENT_BIN="openshield-manager"
SYSTEMD_UNIT="openshield-manager.service"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Detect OS
OS="$(uname | tr '[:upper:]' '[:lower:]')"

echo "Detected OS: $OS"

# Require VERSION argument
if [[ -z "$1" ]]; then
    echo "Usage: $0 <VERSION|latest>"
    echo "Example: $0 v1.0.1"
    echo "         $0 latest"
    exit 1
fi
VERSION="$1"

# If version is "latest", fetch the latest release tag from GitHub API
if [[ "$VERSION" == "latest" ]]; then
    echo "Fetching latest version from GitHub..."
    VERSION=$(curl -s https://api.github.com/repos/YordanOpenShield/openshield-manager/releases/latest | grep -oP '"tag_name":\s*"\K(.*)(?=")')
    if [[ -z "$VERSION" ]]; then
        echo "Failed to fetch latest version."
        exit 1
    fi
    echo "Latest version is $VERSION"
fi

if [[ "$OS" == "linux" ]]; then
    # Download manager binary if not present
    if [[ ! -f "$AGENT_BIN" ]]; then
        AGENT_URL="https://github.com/YordanOpenShield/openshield-manager/releases/download/${VERSION}/openshield-manager-linux-amd64-${VERSION}"
        echo "Downloading manager binary from $AGENT_URL ..."
        curl -L -o "$AGENT_BIN" "$AGENT_URL"
        chmod +x "$AGENT_BIN"
    fi

    echo "Copying $AGENT_BIN to $INSTALL_DIR (requires sudo)..."
    sudo cp "$AGENT_BIN" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$AGENT_BIN"

    # Download systemd unit file if not present
    if [[ ! -f "$SYSTEMD_UNIT" ]]; then
        UNIT_URL="https://raw.githubusercontent.com/YordanOpenShield/openshield-manager/refs/heads/main/helpers/openshield-manager.service"
        echo "Downloading systemd unit file from $UNIT_URL ..."
        curl -L -o "$SYSTEMD_UNIT" "$UNIT_URL"
    fi
    # Copy systemd unit file
    echo "Copying $SYSTEMD_UNIT to $SYSTEMD_DIR (requires sudo)..."
    sudo cp "$SYSTEMD_UNIT" "$SYSTEMD_DIR/"
    sudo systemctl daemon-reload
    sudo systemctl enable openshield-manager
    sudo systemctl start openshield-manager
    echo "OpenShield Manager installed and started as a systemd service."

elif [[ "$OS" == "darwin" ]]; then
    echo "Detected macOS. Installing manager binary..."
    if [[ ! -f "$AGENT_BIN" ]]; then
        AGENT_URL="https://github.com/YordanOpenShield/openshield-manager/releases/download/${VERSION}/openshield-manager-darwin-amd64-${VERSION}"
        echo "Downloading manager binary from $AGENT_URL ..."
        curl -L -o "$AGENT_BIN" "$AGENT_URL"
        chmod +x "$AGENT_BIN"
    fi
    cp "$AGENT_BIN" /usr/local/bin/
    chmod +x /usr/local/bin/$AGENT_BIN
    echo "Manager installed to /usr/local/bin."
    echo "Note: macOS uses launchd, not systemd. Please create a launchd plist if you want to run as a service."

else
    echo "Unsupported OS: $OS"
    exit 1
fi