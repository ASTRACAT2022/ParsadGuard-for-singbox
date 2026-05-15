#!/usr/bin/env bash
set -euo pipefail

# ============================================================
#  PasarGuard-for-singbox — Unified Installer
#  Repo: https://github.com/ASTRACAT2022/ParsadGuard-for-singbox
#
#  Interactive:   curl -sLo install.sh https://... && bash install.sh
#  Non-interactive:
#    bash install.sh panel
#    bash install.sh node
#    bash install.sh sub [lang]
#    bash install.sh all
# ============================================================

REPO="https://github.com/ASTRACAT2022/ParsadGuard-for-singbox"
RAW="https://raw.githubusercontent.com/ASTRACAT2022/ParsadGuard-for-singbox/main"
DEFAULT_PANEL_DIR="/opt/pasarguard"
DEFAULT_NODE_DIR="/opt/pasarguard-node"
DEFAULT_SUB_DIR="/var/lib/pasarguard/templates/subscription"
DEFAULT_ENV_FILE="/opt/pasarguard/.env"

# --- Colors ---
RED=$'\033[0;31m'; GREEN=$'\033[0;32m'; YELLOW=$'\033[1;33m'; BLUE=$'\033[1;34m'; NC=$'\033[0m'
info()  { printf "%s[+]%s %s\n" "$GREEN" "$NC" "$*"; }
warn()  { printf "%s[!]%s %s\n" "$YELLOW" "$NC" "$*"; }
err()   { printf "%s[x]%s %s\n" "$RED" "$NC" "$*"; }
title() { printf "\n%s=== %s ===%s\n\n" "$BLUE" "$*" "$NC"; }

# ============================================================
#  Utility
# ============================================================
download() {
    if command -v curl &>/dev/null; then
        curl -fsSL "$1" -o "$2"
    elif command -v wget &>/dev/null; then
        wget -q "$1" -O "$2"
    else
        err "Neither curl nor wget found. Install one first."
        exit 1
    fi
}

detect_os() {
    if [ -f /etc/os-release ]; then . /etc/os-release; OS=$ID; else OS=$(uname -s | tr '[:upper:]' '[:lower:]'); fi
    case "$OS" in
        ubuntu|debian)            PKG_MGR="apt-get"; INSTALL_CMD="sudo apt-get update && sudo apt-get install -y" ;;
        centos|rhel|fedora)       PKG_MGR="dnf";    INSTALL_CMD="sudo $PKG_MGR install -y" ;;
        arch|manjaro)             PKG_MGR="pacman"; INSTALL_CMD="sudo pacman -Sy --noconfirm" ;;
        alpine)                   PKG_MGR="apk";    INSTALL_CMD="sudo apk add" ;;
        *) warn "Unknown OS: $OS. Install dependencies manually."; PKG_MGR=""; INSTALL_CMD="echo" ;;
    esac
}

ensure_deps() {
    local deps=("$@")
    local missing=()
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &>/dev/null; then missing+=("$dep"); fi
    done
    if [ ${#missing[@]} -gt 0 ]; then
        info "Installing missing: ${missing[*]}"
        $INSTALL_CMD "${missing[@]}"
    fi
}

setup_service() {
    local name="$1" desc="$2" exec_cmd="$3" workdir="$4"
    local svc="/etc/systemd/system/${name}.service"
    sudo tee "$svc" >/dev/null <<EOF
[Unit]
Description=${desc}
After=network.target

[Service]
ExecStart=${exec_cmd}
Restart=on-failure
WorkingDirectory=${workdir}

[Install]
WantedBy=multi-user.target
EOF
    sudo systemctl daemon-reload
    info "Service ${name} created. Enable with: sudo systemctl enable --now ${name}"
}

# ============================================================
#  Panel install
# ============================================================
install_panel() {
    title "Installing PasarGuard Panel"
    detect_os

    PANEL_DIR="${PANEL_DIR:-$DEFAULT_PANEL_DIR}"
    info "Install directory: $PANEL_DIR"

    # Install python3 + venv (distro-specific name)
    ensure_deps python3 git curl
    # Debian splits ensurepip into python3.X-venv — install it unconditionally
    if [ -f /etc/debian_version ]; then
        PY_VER=$(python3 -c "import sys;print(f'{sys.version_info.major}.{sys.version_info.minor}')" 2>/dev/null || echo "3")
        info "Installing python${PY_VER}-venv..."
        sudo apt-get update -qq && sudo apt-get install -y -qq "python${PY_VER}-venv" || {
            warn "Failed to install python${PY_VER}-venv, trying python3-venv..."
            sudo apt-get install -y -qq python3-venv
        }
    elif ! python3 -m venv --help &>/dev/null; then
        $INSTALL_CMD python3-venv
    fi

    if [ -d "$PANEL_DIR/.git" ]; then
        info "Existing install found. Updating..."
        git -C "$PANEL_DIR" pull
    else
        info "Cloning repository..."
        sudo mkdir -p "$PANEL_DIR"
        sudo chown "$(whoami)" "$PANEL_DIR"
        git clone --depth 1 --filter=blob:none --sparse "$REPO" "$PANEL_DIR"
        git -C "$PANEL_DIR" sparse-checkout set panel-main
        mv "$PANEL_DIR/panel-main/"* "$PANEL_DIR/" 2>/dev/null || true
        mv "$PANEL_DIR/panel-main/".* "$PANEL_DIR/" 2>/dev/null || true
        rmdir "$PANEL_DIR/panel-main" 2>/dev/null || true
    fi

    cd "$PANEL_DIR"

    info "Setting up Python virtual environment..."
    python3 -m venv .venv || {
        err "Failed to create virtual environment. Install python3-venv and retry."
        exit 1
    }
    # shellcheck disable=SC1091
    source .venv/bin/activate

    info "Installing Python dependencies..."
    pip install --upgrade pip
    pip install --upgrade uv 2>/dev/null || true
    if command -v uv &>/dev/null; then
        uv pip install -e ".[dev]"
    else
        pip install -e ".[dev]"
    fi

    if [ ! -f .env ]; then
        info "Generating .env file..."
        cat > .env <<'ENVEOF'
# Database — defaults to SQLite
SQLALCHEMY_DATABASE_URL=sqlite+aiosqlite:///db.sqlite3

# Admin credentials (change immediately after first login)
SUDO_USERNAME=admin
SUDO_PASSWORD=admin

# Server
UVICORN_HOST=0.0.0.0
UVICORN_PORT=8000

# Subscription path
SUBSCRIPTION_PATH=sub

# WireGuard
WIREGUARD_ENABLED=true

# Roles: all-in-one | node | scheduler
ROLE=all-in-one

# Node API key (UUID)
# API_KEY=xxxxxxxx-yyyy-zzzz-mmmm-aaaaaaaaaaa
ENVEOF
        warn "Default .env created. Change SUDO_USERNAME/SUDO_PASSWORD immediately!"
    fi

    info "Running database migrations..."
    python -m alembic upgrade head

    setup_service "pasarguard" "PasarGuard Panel — sing-box Edition" \
        "$PANEL_DIR/.venv/bin/python $PANEL_DIR/main.py" \
        "$PANEL_DIR"

    echo ""
    info "Panel installed."
    printf "  Start:  sudo systemctl start pasarguard\n"
    printf "  Enable: sudo systemctl enable --now pasarguard\n"
    printf "  Logs:   sudo journalctl -u pasarguard -f\n"
    printf "  Web:    http://<server-ip>:8000/dashboard/\n"
}

# ============================================================
#  Node install
# ============================================================
install_node() {
    title "Installing PasarGuard Node (sing-box)"
    detect_os

    NODE_DIR="${NODE_DIR:-$DEFAULT_NODE_DIR}"
    info "Install directory: $NODE_DIR"

    ensure_deps curl tar git

    # --- Install sing-box binary ---
    if ! command -v sing-box &>/dev/null; then
        info "Installing sing-box..."
        SB_ARCH="amd64"
        case "$(uname -m)" in
            aarch64|arm64) SB_ARCH="arm64" ;;
            armv7l)         SB_ARCH="armv7" ;;
        esac
        SB_VER="v1.10.8"
        SB_URL="https://github.com/SagerNet/sing-box/releases/download/${SB_VER}/sing-box-${SB_VER#v}-linux-${SB_ARCH}.tar.gz"
        download "$SB_URL" /tmp/sing-box.tgz
        tar xzf /tmp/sing-box.tgz -C /tmp
        sudo install -m 0755 /tmp/sing-box-${SB_VER#v}-linux-${SB_ARCH}/sing-box /usr/local/bin/sing-box
        rm -rf /tmp/sing-box.tgz /tmp/sing-box-${SB_VER#v}-linux-${SB_ARCH}
        info "sing-box $(sing-box version 2>/dev/null | head -1 || echo 'installed')"
    else
        info "sing-box already installed: $(sing-box version 2>/dev/null | head -1)"
    fi

    # --- Build or download node binary ---
    if [ -d "$NODE_DIR/.git" ]; then
        info "Existing node found. Updating..."
        git -C "$NODE_DIR" pull
    else
        info "Cloning node repository..."
        sudo mkdir -p "$NODE_DIR"
        sudo chown "$(whoami)" "$NODE_DIR"
        git clone --depth 1 --filter=blob:none --sparse "$REPO" "$NODE_DIR"
        git -C "$NODE_DIR" sparse-checkout set node-main
        mv "$NODE_DIR/node-main/"* "$NODE_DIR/" 2>/dev/null || true
        mv "$NODE_DIR/node-main/".* "$NODE_DIR/" 2>/dev/null || true
        rmdir "$NODE_DIR/node-main" 2>/dev/null || true
    fi

    cd "$NODE_DIR"

    # Check for pre-built binary
    info "Checking for pre-built binary..."
    NODE_ARCH="amd64"; case "$(uname -m)" in aarch64|arm64) NODE_ARCH="arm64" ;; esac
    BIN_URL="${REPO}/releases/latest/download/pasarguard-node-linux-${NODE_ARCH}"
    if download "$BIN_URL" "$NODE_DIR/pasarguard-node" 2>/dev/null; then
        chmod +x "$NODE_DIR/pasarguard-node"
        info "Downloaded pre-built pasarguard-node"
    else
        if ! command -v go &>/dev/null; then
            info "Installing Go..."
            GO_VER="1.23.4"
            download "https://go.dev/dl/go${GO_VER}.linux-${NODE_ARCH}.tar.gz" /tmp/go.tgz
            sudo tar -C /usr/local -xzf /tmp/go.tgz
            export PATH="/usr/local/go/bin:$PATH"
            rm /tmp/go.tgz
        fi
        info "Building pasarguard-node from source..."
        go build -o pasarguard-node ./cmd/node
        info "Built pasarguard-node"
    fi

    # --- Generate default .env ---
    if [ ! -f .env ]; then
        API_KEY=$(python3 -c "import uuid; print(uuid.uuid4())" 2>/dev/null || uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid)
        info "Generating .env (API_KEY=$API_KEY)"
        cat > .env <<EOF
SERVICE_PORT=62050
SERVICE_PROTOCOL=grpc
SINGBOX_EXECUTABLE_PATH=/usr/local/bin/sing-box
GENERATED_CONFIG_PATH=/var/lib/pg-node/generated
SSL_CERT_FILE=/var/lib/pg-node/certs/ssl_cert.pem
SSL_KEY_FILE=/var/lib/pg-node/certs/ssl_key.pem
API_KEY=$API_KEY
EOF
        info "Generating self-signed TLS cert..."
        mkdir -p /var/lib/pg-node/certs
        openssl req -x509 -newkey rsa:4096 -keyout /var/lib/pg-node/certs/ssl_key.pem \
            -out /var/lib/pg-node/certs/ssl_cert.pem -days 36500 -nodes \
            -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" 2>/dev/null
    fi

    mkdir -p /var/lib/pg-node/generated

    setup_service "pasarguard-node" "PasarGuard Node — sing-box" \
        "$NODE_DIR/pasarguard-node" \
        "$NODE_DIR"

    echo ""
    info "Node installed."
    printf "  Start:  sudo systemctl start pasarguard-node\n"
    printf "  Enable: sudo systemctl enable --now pasarguard-node\n"
    printf "  Logs:   sudo journalctl -u pasarguard-node -f\n"
    echo ""
    warn "Copy the API_KEY from $NODE_DIR/.env into the panel's node settings."
}

# ============================================================
#  Subscription template install
# ============================================================
install_subscription() {
    title "Installing Subscription Template"
    detect_os

    LANG_CODE="${1:-fa}"
    case "$LANG_CODE" in en|fa|zh|ru) ;; *) err "Invalid language: $LANG_CODE. Use en|fa|zh|ru"; exit 1 ;; esac

    ensure_deps curl

    SUB_DIR="${SUB_DIR:-$DEFAULT_SUB_DIR}"
    sudo mkdir -p "$SUB_DIR"

    URL="${RAW}/subscription-template-main/dist/${LANG_CODE}.html"
    FA_URL="${RAW}/subscription-template-main/dist/index.html"
    if [ "$LANG_CODE" = "fa" ]; then URL="$FA_URL"; fi

    info "Downloading template (lang=$LANG_CODE)..."
    download "$URL" "$SUB_DIR/index.html"

    # Update panel .env
    ENV_FILE="${ENV_FILE:-$DEFAULT_ENV_FILE}"
    sudo mkdir -p "$(dirname "$ENV_FILE")"
    if [ -f "$ENV_FILE" ]; then
        if grep -q '^CUSTOM_TEMPLATES_DIRECTORY=' "$ENV_FILE"; then
            sudo sed -i 's|^CUSTOM_TEMPLATES_DIRECTORY=.*|CUSTOM_TEMPLATES_DIRECTORY="/var/lib/pasarguard/templates/"|' "$ENV_FILE"
        else
            echo 'CUSTOM_TEMPLATES_DIRECTORY="/var/lib/pasarguard/templates/"' | sudo tee -a "$ENV_FILE" >/dev/null
        fi
        if grep -q '^SUBSCRIPTION_PAGE_TEMPLATE=' "$ENV_FILE"; then
            sudo sed -i 's|^SUBSCRIPTION_PAGE_TEMPLATE=.*|SUBSCRIPTION_PAGE_TEMPLATE="subscription/index.html"|' "$ENV_FILE"
        else
            echo 'SUBSCRIPTION_PAGE_TEMPLATE="subscription/index.html"' | sudo tee -a "$ENV_FILE" >/dev/null
        fi
    fi

    echo ""
    info "Subscription template installed ($LANG_CODE) at $SUB_DIR/index.html"
    printf "  Restart panel: sudo systemctl restart pasarguard\n"
}

# ============================================================
#  Menu
# ============================================================
show_menu() {
    printf "\n"
    printf "  %sPasarGuard-for-singbox — Unified Installer%s\n" "$BLUE" "$NC"
    printf "  %sRepo:%s %s\n" "$BLUE" "$NC" "$REPO"
    printf "\n"
    printf "  1) Install Panel\n"
    printf "  2) Install Node\n"
    printf "  3) Install Subscription Template\n"
    printf "  4) Install All (Panel + Node)\n"
    printf "  5) Exit\n"
    printf "\n"
}

show_quick_help() {
    printf "\n"
    printf "%s PasarGuard-for-singbox Installer %s\n" "$BLUE" "$NC"
    printf "%s Repo: %s %s\n" "$BLUE" "$NC" "$REPO"
    printf "\n"
    printf "  %s Interactive mode: %s\n" "$GREEN" "$NC"
    printf "    curl -sLo install.sh %s/raw/main/install.sh\n" "$REPO"
    printf "    bash install.sh\n"
    printf "\n"
    printf "  %s Direct mode: %s\n" "$GREEN" "$NC"
    printf "    curl -sLo install.sh %s/raw/main/install.sh && bash install.sh panel\n" "$REPO"
    printf "    curl -sLo install.sh %s/raw/main/install.sh && bash install.sh node\n" "$REPO"
    printf "    curl -sLo install.sh %s/raw/main/install.sh && bash install.sh sub\n" "$REPO"
    printf "    curl -sLo install.sh %s/raw/main/install.sh && bash install.sh all\n" "$REPO"
    printf "\n"
    exit 0
}

# ============================================================
#  Main
# ============================================================
main() {
    # Direct mode — first arg is the component
    if [ $# -ge 1 ]; then
        case "$1" in
            panel)   install_panel; exit 0 ;;
            node)    install_node; exit 0 ;;
            sub)     install_subscription "${2:-fa}"; exit 0 ;;
            all)     install_panel; install_node; exit 0 ;;
            -h|--help|help) show_quick_help; exit 0 ;;
            *)       err "Unknown component: $1"; show_quick_help; exit 1 ;;
        esac
    fi

    # If stdin is NOT a terminal, we can't show interactive menu
    if [ ! -t 0 ]; then
        show_quick_help
    fi

    while true; do
        show_menu
        read -rp "  Select [1-5]: " choice
        case "$choice" in
            1) install_panel ;;
            2) install_node ;;
            3)
                read -rp "  Language (en/fa/zh/ru) [fa]: " lang
                lang="${lang:-fa}"
                install_subscription "$lang"
                ;;
            4) install_panel; install_node ;;
            5) info "Done."; exit 0 ;;
            *) err "Invalid choice" ;;
        esac
        echo ""
        read -rp "  Press Enter to return to menu..."
    done
}

main "$@"
