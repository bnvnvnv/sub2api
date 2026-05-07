#!/usr/bin/env bash
#
# Pull upstream source, merge the current custom branch, validate, build, and
# switch the installed Sub2API binary only after all checks pass.
#
# Typical systemd usage:
#   install -m 0755 tools/sub2api-auto-upgrade.sh /usr/local/bin/sub2api-auto-upgrade
#   install -m 0644 deploy/sub2api-auto-upgrade.service /etc/systemd/system/
#   install -m 0644 deploy/sub2api-auto-upgrade.timer /etc/systemd/system/
#   install -m 0644 deploy/sub2api-auto-upgrade.env.example /etc/sub2api/auto-upgrade.env
#   systemctl daemon-reload
#   systemctl enable --now sub2api-auto-upgrade.timer

set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"

usage() {
  cat <<'EOF'
Usage:
  sub2api-auto-upgrade [--dry-run] [--force] [--no-restart]

Environment overrides:
  REPO_DIR                  Source repository path. Default: script parent dir
  UPSTREAM_REMOTE           Upstream git remote. Default: origin
  UPSTREAM_BRANCH           Upstream branch. Default: main
  LOCAL_BRANCH              Local branch to update. Default: current branch
  FORK_REMOTE               Optional fork remote for push. Default: fork
  PUSH_AFTER_UPGRADE        Push local branch after successful upgrade. Default: 0
  SERVICE_NAME              systemd service name. Default: sub2api
  INSTALL_DIR               Runtime install dir. Default: /opt/sub2api
  BINARY_PATH               Runtime binary path. Default: $INSTALL_DIR/sub2api
  HEALTH_URL                Health check URL. Default: http://127.0.0.1:${SERVER_PORT:-8080}/health
  FRONTEND_BUILD_CMD        Frontend build command. Default: npm run build
  RUN_FRONTEND_BUILD        Build frontend before Go binary. Default: 1
  RUN_GO_TESTS              Run focused Go tests. Default: 1
  RUN_GENERATE              Run go generate before tests/build. Default: 0
  RESTART_SERVICE           Restart service after binary switch. Default: 1
  FORCE_REBUILD             Build/restart even when upstream has no new commits. Default: 0
EOF
}

log() {
  printf '[%s] [%s] %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$SCRIPT_NAME" "$*"
}

die() {
  log "ERROR: $*"
  exit 1
}

as_bool() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
default_repo_dir="$(cd "$script_dir/.." && pwd)"

REPO_DIR="${REPO_DIR:-$default_repo_dir}"
UPSTREAM_REMOTE="${UPSTREAM_REMOTE:-origin}"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH:-main}"
LOCAL_BRANCH="${LOCAL_BRANCH:-}"
FORK_REMOTE="${FORK_REMOTE:-fork}"
PUSH_AFTER_UPGRADE="${PUSH_AFTER_UPGRADE:-0}"

SERVICE_NAME="${SERVICE_NAME:-sub2api}"
INSTALL_DIR="${INSTALL_DIR:-/opt/sub2api}"
BINARY_PATH="${BINARY_PATH:-$INSTALL_DIR/sub2api}"
BACKUP_DIR="${BACKUP_DIR:-$INSTALL_DIR/backups/auto-upgrade}"
LOCK_FILE="${LOCK_FILE:-/tmp/sub2api-auto-upgrade.lock}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:${SERVER_PORT:-8080}/health}"
HEALTH_TIMEOUT_SECONDS="${HEALTH_TIMEOUT_SECONDS:-60}"

RUN_FRONTEND_BUILD="${RUN_FRONTEND_BUILD:-1}"
FRONTEND_BUILD_CMD="${FRONTEND_BUILD_CMD:-npm run build}"
RUN_GO_TESTS="${RUN_GO_TESTS:-1}"
RUN_GENERATE="${RUN_GENERATE:-0}"
RESTART_SERVICE="${RESTART_SERVICE:-1}"
FORCE_REBUILD="${FORCE_REBUILD:-0}"
DRY_RUN="${DRY_RUN:-0}"

GO_BUILD_TAGS="${GO_BUILD_TAGS:-embed}"
GOCACHE="${GOCACHE:-/tmp/go-build-sub2api-auto-upgrade}"
GOMODCACHE="${GOMODCACHE:-/tmp/go-mod-sub2api-auto-upgrade}"
GO_TEST_PACKAGES="${GO_TEST_PACKAGES:-./cmd/server ./internal/handler/admin ./internal/service ./internal/pkg/proxyurl ./internal/pkg/proxyutil}"
GO_TEST_RUN="${GO_TEST_RUN:-Test(OpenAIWeb|ForwardOpenAIWeb|BuildOpenAIWeb|PrepareOpenAIWeb|AccountSupportsOpenAIWeb|ParseSubscription|ParseCPA|AccountHandler|BatchUpdateCredentials|ConfigureTransportProxy)}"

while (($#)); do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      ;;
    --force)
      FORCE_REBUILD=1
      ;;
    --no-restart)
      RESTART_SERVICE=0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
  shift
done

command -v git >/dev/null 2>&1 || die "git is required"
command -v go >/dev/null 2>&1 || die "go is required"
command -v curl >/dev/null 2>&1 || die "curl is required"

mkdir -p "$(dirname "$LOCK_FILE")"
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  die "another auto-upgrade is already running: $LOCK_FILE"
fi

tmp_dir=""
backup_binary=""
cleanup() {
  if [[ -n "$tmp_dir" && -d "$tmp_dir" ]]; then
    rm -rf "$tmp_dir"
  fi
}
trap cleanup EXIT

cd "$REPO_DIR"
git rev-parse --show-toplevel >/dev/null 2>&1 || die "REPO_DIR is not a git repository: $REPO_DIR"

if [[ -z "$LOCAL_BRANCH" ]]; then
  LOCAL_BRANCH="$(git symbolic-ref --quiet --short HEAD)" || die "detached HEAD; set LOCAL_BRANCH"
fi

current_branch="$(git symbolic-ref --quiet --short HEAD)" || die "detached HEAD; checkout $LOCAL_BRANCH first"
if [[ "$current_branch" != "$LOCAL_BRANCH" ]]; then
  die "current branch is $current_branch, expected $LOCAL_BRANCH"
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  die "working tree is dirty; commit or stash local changes before auto-upgrade"
fi

log "fetching $UPSTREAM_REMOTE/$UPSTREAM_BRANCH"
git fetch --prune "$UPSTREAM_REMOTE" "$UPSTREAM_BRANCH"

upstream_ref="$UPSTREAM_REMOTE/$UPSTREAM_BRANCH"
upstream_commit="$(git rev-parse "$upstream_ref")"
before_commit="$(git rev-parse HEAD)"

if git merge-base --is-ancestor "$upstream_commit" HEAD && ! as_bool "$FORCE_REBUILD"; then
  log "already includes $upstream_ref ($upstream_commit); nothing to do"
  exit 0
fi

if ! git merge-base --is-ancestor "$upstream_commit" HEAD; then
  log "merging $upstream_ref into $LOCAL_BRANCH"
  if ! git merge --no-edit "$upstream_commit"; then
    git merge --abort >/dev/null 2>&1 || true
    die "merge failed; upstream changes require manual conflict resolution"
  fi
else
  log "force rebuild requested without new upstream commits"
fi

after_merge_commit="$(git rev-parse HEAD)"
log "source ready: $before_commit -> $after_merge_commit"

if as_bool "$RUN_GENERATE"; then
  log "running go generate"
  (cd backend && GOCACHE="$GOCACHE" GOMODCACHE="$GOMODCACHE" go generate ./cmd/server)
fi

log "checking whitespace"
git diff --check

if as_bool "$RUN_GO_TESTS"; then
  log "running focused Go tests"
  # shellcheck disable=SC2086
  (cd backend && GOCACHE="$GOCACHE" GOMODCACHE="$GOMODCACHE" go test $GO_TEST_PACKAGES -run "$GO_TEST_RUN")
fi

tmp_dir="$(mktemp -d /tmp/sub2api-auto-upgrade.XXXXXX)"
new_binary="$tmp_dir/sub2api"

if as_bool "$RUN_FRONTEND_BUILD"; then
  command -v npm >/dev/null 2>&1 || die "npm is required when RUN_FRONTEND_BUILD=1"
  log "building frontend: $FRONTEND_BUILD_CMD"
  (cd frontend && bash -lc "$FRONTEND_BUILD_CMD")
fi

version="$(tr -d '\r\n' < backend/cmd/server/VERSION 2>/dev/null || true)"
if [[ -z "$version" ]]; then
  version="0.0.0-dev"
fi
commit="$(git rev-parse --short=12 HEAD)"
built_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
ldflags="-s -w -X main.Version=$version -X main.Commit=$commit -X main.Date=$built_at -X main.BuildType=source"

log "building backend binary"
(cd backend && CGO_ENABLED=0 GOCACHE="$GOCACHE" GOMODCACHE="$GOMODCACHE" go build -tags "$GO_BUILD_TAGS" -ldflags "$ldflags" -trimpath -o "$new_binary" ./cmd/server)
"$new_binary" --version >/dev/null

if as_bool "$DRY_RUN"; then
  log "dry-run complete; binary built at $new_binary"
  exit 0
fi

if [[ ! -d "$INSTALL_DIR" ]]; then
  die "INSTALL_DIR does not exist: $INSTALL_DIR"
fi

mkdir -p "$BACKUP_DIR"
timestamp="$(date -u '+%Y%m%d%H%M%S')"
if [[ -f "$BINARY_PATH" ]]; then
  backup_binary="$BACKUP_DIR/sub2api.$timestamp"
  log "backing up current binary to $backup_binary"
  cp -a "$BINARY_PATH" "$backup_binary"
fi

owner="root"
group="root"
if [[ -e "$BINARY_PATH" ]]; then
  owner="$(stat -c '%U' "$BINARY_PATH" 2>/dev/null || echo root)"
  group="$(stat -c '%G' "$BINARY_PATH" 2>/dev/null || echo root)"
fi

log "installing new binary to $BINARY_PATH"
install -o "$owner" -g "$group" -m 0755 "$new_binary" "$BINARY_PATH.new"
mv -f "$BINARY_PATH.new" "$BINARY_PATH"

if [[ -d backend/resources ]]; then
  log "syncing backend resources"
  mkdir -p "$INSTALL_DIR/resources"
  cp -a backend/resources/. "$INSTALL_DIR/resources/"
fi

wait_health() {
  local deadline
  deadline=$((SECONDS + HEALTH_TIMEOUT_SECONDS))
  while ((SECONDS < deadline)); do
    if curl -fsS --max-time 3 "$HEALTH_URL" >/dev/null; then
      return 0
    fi
    sleep 2
  done
  return 1
}

rollback_binary() {
  if [[ -n "$backup_binary" && -f "$backup_binary" ]]; then
    log "rolling back binary from $backup_binary"
    cp -a "$backup_binary" "$BINARY_PATH"
    if as_bool "$RESTART_SERVICE"; then
      systemctl restart "$SERVICE_NAME" || true
    fi
  fi
}

if as_bool "$RESTART_SERVICE"; then
  command -v systemctl >/dev/null 2>&1 || die "systemctl is required when RESTART_SERVICE=1"
  log "restarting $SERVICE_NAME"
  if ! systemctl restart "$SERVICE_NAME"; then
    rollback_binary
    die "systemctl restart failed"
  fi

  log "waiting for health: $HEALTH_URL"
  if ! wait_health; then
    rollback_binary
    die "health check failed after restart"
  fi
else
  log "RESTART_SERVICE=0; binary installed but service not restarted"
fi

if as_bool "$PUSH_AFTER_UPGRADE"; then
  log "pushing $LOCAL_BRANCH to $FORK_REMOTE"
  git push "$FORK_REMOTE" "HEAD:$LOCAL_BRANCH"
fi

log "upgrade completed successfully at $after_merge_commit"
