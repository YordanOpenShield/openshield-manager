#!/usr/bin/env bash

set -e

OPENSHIELD_DIR="/etc/openshield"
CERT_DIR="$OPENSHIELD_DIR/certs"
AGENT_BIN="openshield-manager"
SYSTEMD_UNIT="openshield-manager.service"
INSTALL_DIR="/usr/local/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Create necessary directories if they don't exist
mkdir -p "$OPENSHIELD_DIR"
mkdir -p "$CERT_DIR"

echo "Generating certificates in $CERT_DIR"
# Generate CA if not exists
if [ ! -f "$CERT_DIR/ca.key" ]; then
  openssl genrsa -out "$CERT_DIR/ca.key" 4096
  openssl req -x509 -new -key "$CERT_DIR/ca.key" -sha256 -days 3650 -out "$CERT_DIR/ca.crt" -subj "/CN=OpenShieldCA"
fi
# Generate manager cert if not exists
if [ ! -f "$CERT_DIR/manager.key" ]; then
  openssl genrsa -out "$CERT_DIR/manager.key" 4096
  openssl req -new -key "$CERT_DIR/manager.key" -out "$CERT_DIR/manager.csr" -subj "/CN=manager"
  openssl x509 -req -in "$CERT_DIR/manager.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/manager.crt" -days 3650 -sha256
fi

echo "Manager and CA certificates generated in $CERT_DIR"

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
    cp "$AGENT_BIN" "$INSTALL_DIR/"
    chmod +x $INSTALL_DIR/$AGENT_BIN
    echo "Manager installed to $INSTALL_DIR."
    echo "Note: macOS uses launchd, not systemd. Please create a launchd plist if you want to run as a service."

else
    echo "Unsupported OS: $OS"
    exit 1
fi