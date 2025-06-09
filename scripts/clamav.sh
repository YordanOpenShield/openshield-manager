#!/bin/bash
# filepath: e:\OpenShield\openshield-agent\scripts\clamav.sh

set -e

ACTION="$1"

install_clamav() {
    # Detect distro
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case "$ID" in
        ubuntu | debian | kali)
            sudo apt-get update
            sudo apt-get install -y clamav
            ;;
        centos | rhel | fedora)
            sudo yum install -y epel-release
            sudo yum install -y clamav
            ;;
        *)
            echo "Unsupported Linux distribution: $ID"
            exit 1
            ;;
        esac
    else
        echo "/etc/os-release not found. Cannot detect Linux distribution."
        exit 1
    fi
}

scan_clamav() {
    # Check if the freshclam service is active
    if systemctl is-active --quiet clamav-freshclam; then
        echo "clamav-freshclam service is running; skipping manual freshclam update."
    else
        echo "clamav-freshclam service is not running; running freshclam manually."
        sudo freshclam
    fi
    sudo clamscan -r / --bell -i
}

uninstall_clamav() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case "$ID" in
        ubuntu | debian | kali)
            sudo apt-get remove -y clamav
            ;;
        centos | rhel | fedora)
            sudo yum remove -y clamav
            ;;
        *)
            echo "Unsupported Linux distribution: $ID"
            exit 1
            ;;
        esac
    else
        echo "/etc/os-release not found. Cannot detect Linux distribution."
        exit 1
    fi
}

case "$ACTION" in
install)
    install_clamav
    ;;
scan)
    scan_clamav
    ;;
uninstall)
    uninstall_clamav
    ;;
*)
    echo "Usage: $0 {install|scan|uninstall}"
    exit 1
    ;;
esac
