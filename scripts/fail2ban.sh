#!/bin/bash
# Cross-platform (Linux/Mac) fail2ban installer with minimal config
set -e

# Detect OS
if [ -f /etc/debian_version ]; then
    OS=debian
elif [ -f /etc/redhat-release ]; then
    OS=redhat
elif [ "$(uname)" = "Darwin" ]; then
    OS=mac
else
    echo "Unsupported OS. This script supports Debian/Ubuntu, RHEL/CentOS, and macOS (brew)."
    exit 1
fi

# Install fail2ban
if [ "$OS" = "debian" ]; then
    sudo apt-get update
    sudo apt-get install -y fail2ban
elif [ "$OS" = "redhat" ]; then
    sudo yum install -y epel-release
    sudo yum install -y fail2ban
elif [ "$OS" = "mac" ]; then
    brew install fail2ban
fi

# Minimal fail2ban config
sudo tee /etc/fail2ban/jail.local >/dev/null <<EOF
[sshd]
enabled = true
EOF

# Enable and start fail2ban
if [ "$OS" = "mac" ]; then
    sudo fail2ban-client start
else
    sudo systemctl enable fail2ban
    sudo systemctl restart fail2ban
fi

echo "fail2ban installed and started with minimal config."
