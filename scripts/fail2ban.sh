#!/bin/bash
# filepath: e:\OpenShield\openshield-agent\scripts\fail2ban.sh

set -e

ACTION="$1"
shift

detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    else
        echo "unknown"
    fi
}

install_fail2ban() {
    DISTRO=$(detect_distro)
    case "$DISTRO" in
    ubuntu | debian | kali)
        sudo apt-get update
        sudo apt-get install -y fail2ban
        ;;
    centos | rhel | fedora)
        sudo yum install -y epel-release
        sudo yum install -y fail2ban
        ;;
    *)
        echo "Unsupported Linux distribution: $DISTRO"
        exit 1
        ;;
    esac
}

configure_fail2ban() {
    for opt in "$@"; do
        case "$opt" in
        ssh)
            echo -e "[sshd]\nenabled = true" | sudo tee /etc/fail2ban/jail.d/sshd.conf
            ;;
        web)
            cat <<EOF | sudo tee /etc/fail2ban/jail.d/web.conf
[nginx-http-auth]
enabled = true
[nginx-botsearch]
enabled = true
[nginx-limit-req]
enabled = true
[nginx-req-limit]
enabled = true
[nginx-noscript]
enabled = true
[nginx-nohome]
enabled = true
[nginx-badbots]
enabled = true
[apache-auth]
enabled = true
[apache-badbots]
enabled = true
[apache-noscript]
enabled = true
[apache-overflows]
enabled = true
[apache-nohome]
enabled = true
[apache-shellshock]
enabled = true
EOF
            ;;
        mail)
            echo -e "[dovecot]\nenabled = true\n[postfix]\nenabled = true" | sudo tee /etc/fail2ban/jail.d/mail.conf
            ;;
        *)
            echo "Unsupported module: $opt"
            exit 1
            ;;
        esac
    done
}

start_fail2ban() {
    if systemctl is-active --quiet fail2ban; then
        sudo systemctl restart fail2ban
    else
        sudo systemctl start fail2ban
    fi
}

stop_fail2ban() {
    if systemctl is-active --quiet fail2ban; then
        sudo systemctl stop fail2ban
    fi
}

uninstall_fail2ban() {
    DISTRO=$(detect_distro)
    case "$DISTRO" in
    ubuntu | debian | kali)
        sudo apt-get remove -y fail2ban
        sudo apt-get autoremove -y
        ;;
    centos | rhel | fedora)
        sudo yum remove -y fail2ban
        ;;
    *)
        echo "Unsupported Linux distribution: $DISTRO"
        exit 1
        ;;
    esac
}

case "$ACTION" in
install)
    install_fail2ban
    ;;
configure)
    configure_fail2ban "$@"
    ;;
start)
    start_fail2ban
    ;;
stop)
    stop_fail2ban
    ;;
uninstall)
    uninstall_fail2ban
    ;;
*)
    echo "Usage: $0 {install|configure|start|stop|uninstall} [options]"
    exit 1
    ;;
esac
