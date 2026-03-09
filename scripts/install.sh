#!/usr/bin/env bash

set -euo pipefail

APP_NAME="mkBlog"
DEFAULT_REPO_URL="https://github.com/mkitsdts/mkBlog.git"
INSTALL_DIR="${MKBLOG_INSTALL_DIR:-$HOME/.local/share/mkblog}"
REPO_URL="${MKBLOG_REPO_URL:-$DEFAULT_REPO_URL}"
REPO_REF="${MKBLOG_REPO_REF:-main}"
SERVICE_NAME="${MKBLOG_SERVICE_NAME:-mkblog}"
LAUNCHD_LABEL="${MKBLOG_LAUNCHD_LABEL:-com.mkblog.app}"
OS="$(uname -s)"

log() {
  printf '[mkBlog] %s\n' "$*"
}

fail() {
  printf '[mkBlog] ERROR: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

version_ge() {
  [ "$1" = "$2" ] && return 0
  local first
  first="$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n1)"
  [ "$first" = "$2" ]
}

check_go_version() {
  local raw version
  raw="$(go version)"
  version="$(printf '%s\n' "$raw" | sed -E 's/.* go([0-9]+\.[0-9]+(\.[0-9]+)?).*/\1/')"
  [ -n "$version" ] || fail "unable to parse Go version from: $raw"
  version_ge "$version" "1.24.0" || fail "Go 1.24.0 or newer is required, found $version"
}

check_node_version() {
  local raw version
  raw="$(node -v)"
  version="${raw#v}"
  if version_ge "$version" "22.12.0"; then
    return
  fi
  version_ge "$version" "20.19.0" || fail "Node.js 20.19.0+ or 22.12.0+ is required, found $version"
}

ensure_clean_worktree() {
  local status
  status="$(git -C "$INSTALL_DIR" status --porcelain)"
  [ -z "$status" ] || fail "repository has uncommitted changes in $INSTALL_DIR; commit or stash them before update"
}

usage() {
  cat <<EOF
Usage: $0 <command>

Commands:
  install    Clone or update source, build the app, install the service, and start it
  update     Pull latest code, rebuild, reinstall service definition, and restart it
  start      Start the registered service
  stop       Stop the registered service
  restart    Restart the registered service
  status     Show service status
  uninstall  Unregister the service but keep files on disk

Environment variables:
  MKBLOG_INSTALL_DIR     Install directory (default: $HOME/.local/share/mkblog)
  MKBLOG_REPO_URL        Git repository URL (default: $DEFAULT_REPO_URL)
  MKBLOG_REPO_REF        Git branch or tag to install (default: main)
  MKBLOG_SERVICE_NAME    systemd service name on Linux (default: mkblog)
  MKBLOG_LAUNCHD_LABEL   launchd label on macOS (default: com.mkblog.app)
EOF
}

ensure_common_prerequisites() {
  need_cmd git
  need_cmd make
  need_cmd go
  need_cmd node
  need_cmd npm
  check_go_version
  check_node_version

  case "$OS" in
    Linux)
      need_cmd systemctl
      need_cmd sudo
      ;;
    Darwin)
      need_cmd launchctl
      ;;
    *)
      fail "unsupported operating system: $OS"
      ;;
  esac
}

ensure_repo_ready() {
  local mode="${1:-install}"

  if [ -d "$INSTALL_DIR/.git" ]; then
    if [ "$mode" = "update" ]; then
      ensure_clean_worktree
    fi
    log "Updating existing source in $INSTALL_DIR"
    git -C "$INSTALL_DIR" fetch --tags origin
    git -C "$INSTALL_DIR" checkout "$REPO_REF"
    if git -C "$INSTALL_DIR" rev-parse --verify "origin/$REPO_REF" >/dev/null 2>&1; then
      git -C "$INSTALL_DIR" pull --ff-only origin "$REPO_REF"
    fi
    return
  fi

  if [ -e "$INSTALL_DIR" ] && [ ! -d "$INSTALL_DIR/.git" ]; then
    fail "install directory exists but is not a git repository: $INSTALL_DIR"
  fi

  log "Cloning source into $INSTALL_DIR"
  mkdir -p "$(dirname "$INSTALL_DIR")"
  git clone --branch "$REPO_REF" "$REPO_URL" "$INSTALL_DIR"
}

run_make() {
  make \
    -C "$INSTALL_DIR" \
    SERVICE_NAME="$SERVICE_NAME" \
    LAUNCHD_LABEL="$LAUNCHD_LABEL" \
    "$@"
}

install_or_update() {
  local mode="${1:-install}"
  ensure_common_prerequisites
  ensure_repo_ready "$mode"
  log "Building and installing service"
  run_make release
}

start_service() {
  ensure_common_prerequisites
  [ -d "$INSTALL_DIR/.git" ] || fail "mkBlog is not installed in $INSTALL_DIR"
  run_make release-start
}

stop_service() {
  ensure_common_prerequisites
  [ -d "$INSTALL_DIR/.git" ] || fail "mkBlog is not installed in $INSTALL_DIR"
  run_make release-stop
}

restart_service() {
  ensure_common_prerequisites
  [ -d "$INSTALL_DIR/.git" ] || fail "mkBlog is not installed in $INSTALL_DIR"
  run_make release-restart
}

status_service() {
  case "$OS" in
    Linux)
      sudo systemctl status "$SERVICE_NAME" --no-pager
      ;;
    Darwin)
      launchctl print "gui/$(id -u)/$LAUNCHD_LABEL"
      ;;
    *)
      fail "unsupported operating system: $OS"
      ;;
  esac
}

uninstall_service() {
  ensure_common_prerequisites
  [ -d "$INSTALL_DIR/.git" ] || fail "mkBlog is not installed in $INSTALL_DIR"
  run_make release-uninstall
  log "Service uninstalled. Files remain in $INSTALL_DIR"
}

main() {
  local cmd="${1:-}"

  case "$cmd" in
    install)
      install_or_update install
      ;;
    update)
      install_or_update update
      ;;
    start)
      start_service
      ;;
    stop)
      stop_service
      ;;
    restart)
      restart_service
      ;;
    status)
      status_service
      ;;
    uninstall)
      uninstall_service
      ;;
    -h|--help|help|"")
      usage
      ;;
    *)
      fail "unknown command: $cmd"
      ;;
  esac
}

main "$@"
