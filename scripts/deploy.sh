#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# 服务器地址不写进仓库，部署时传入：DEPLOY_HOST=<服务器 IP 或域名> scripts/deploy.sh
DEPLOY_HOST="${DEPLOY_HOST:-}"
if [[ -z "$DEPLOY_HOST" ]]; then
  echo "DEPLOY_HOST is required, e.g. DEPLOY_HOST=example.com scripts/deploy.sh" >&2
  exit 1
fi
DEPLOY_USER="${DEPLOY_USER:-ubuntu}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_ed25519}"
SSH_PORT="${SSH_PORT:-22}"
APP_DIR="${APP_DIR:-/opt/starcup}"
GOARCH="${GOARCH:-amd64}"
RUN_TESTS="${RUN_TESTS:-1}"
REMOTE="$DEPLOY_USER@$DEPLOY_HOST"
SSH_OPTIONS=(-o BatchMode=yes -o IdentitiesOnly=yes -i "$SSH_KEY" -p "$SSH_PORT")
SCP_OPTIONS=(-o BatchMode=yes -o IdentitiesOnly=yes -i "$SSH_KEY" -P "$SSH_PORT")

for command in go node npm tar ssh scp; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Missing required command: $command" >&2
    exit 1
  }
done

if [[ "$RUN_TESTS" == "1" ]]; then
  make check-generated-types
  go test ./internal/server/... ./internal/engine/...
fi

npm --prefix web ci
npm --prefix web run build
if [[ "$RUN_TESTS" == "1" ]]; then
  npm --prefix web test
fi

BUILD_DIR="$(mktemp -d "${TMPDIR:-/tmp}/starcup-deploy.XXXXXX")"
trap 'rm -rf "$BUILD_DIR"' EXIT
mkdir -p "$BUILD_DIR/package/web"
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build -trimpath -ldflags="-s -w" -o "$BUILD_DIR/package/starcup-server" ./cmd/server
cp -R web/dist "$BUILD_DIR/package/web/dist"

RELEASE="$(date +%Y%m%d%H%M%S)-$$"
PACKAGE="$BUILD_DIR/starcup-$RELEASE.tar.gz"
REMOTE_PACKAGE="/tmp/starcup-$RELEASE.tar.gz"
tar -C "$BUILD_DIR/package" -czf "$PACKAGE" starcup-server web

ssh "${SSH_OPTIONS[@]}" "$REMOTE" 'set -e; sudo -n true; if ! command -v nginx >/dev/null 2>&1; then sudo apt-get update; sudo DEBIAN_FRONTEND=noninteractive apt-get install -y nginx; fi'
ssh "${SSH_OPTIONS[@]}" "$REMOTE" "sudo mkdir -p '$APP_DIR/releases' && sudo chown -R '$DEPLOY_USER':'$DEPLOY_USER' '$APP_DIR'"

cat <<EOF | ssh "${SSH_OPTIONS[@]}" "$REMOTE" "sudo tee /etc/systemd/system/starcup.service >/dev/null"
[Unit]
Description=StarCup Game Server
After=network.target

[Service]
Type=simple
User=$DEPLOY_USER
Group=$DEPLOY_USER
WorkingDirectory=$APP_DIR/current
Environment=PORT=8080
Environment=STARCUP_TEST_MODE=0
ExecStart=$APP_DIR/current/starcup-server
Restart=always
RestartSec=3
LimitNOFILE=65535
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF

cat <<'EOF' | ssh "${SSH_OPTIONS[@]}" "$REMOTE" "sudo tee /etc/nginx/sites-available/starcup >/dev/null"
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

scp "${SCP_OPTIONS[@]}" "$PACKAGE" "$REMOTE:$REMOTE_PACKAGE"
ssh "${SSH_OPTIONS[@]}" "$REMOTE" "bash -s -- '$RELEASE' '$APP_DIR' '$REMOTE_PACKAGE'" <<'REMOTE'
set -euo pipefail
release="$1"
app_dir="$2"
remote_package="$3"
release_dir="$app_dir/releases/$release"

mkdir -p "$release_dir"
tar -xzf "$remote_package" -C "$release_dir"
chmod 755 "$release_dir/starcup-server"
previous="$(readlink -f "$app_dir/current" 2>/dev/null || true)"
ln -sfn "$release_dir" "$app_dir/current"

sudo systemctl daemon-reload
sudo systemctl enable starcup.service
sudo systemctl restart starcup.service
sudo ln -sfn /etc/nginx/sites-available/starcup /etc/nginx/sites-enabled/starcup
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl enable --now nginx.service
sudo systemctl reload nginx.service

healthy=0
for _ in {1..20}; do
  if curl -fsS http://127.0.0.1:8080/api/health >/dev/null; then
    healthy=1
    break
  fi
  sleep 1
done

if [[ "$healthy" != "1" ]]; then
  echo "Health check failed; restoring the previous release." >&2
  if [[ -n "$previous" ]]; then
    ln -sfn "$previous" "$app_dir/current"
    sudo systemctl restart starcup.service
  else
    sudo systemctl stop starcup.service
  fi
  sudo journalctl -u starcup.service -n 50 --no-pager >&2
  exit 1
fi

rm -f "$remote_package"
for old_release in $(ls -1dt "$app_dir"/releases/* | tail -n +6); do
  rm -rf -- "$old_release"
done
REMOTE

external_healthy=0
for _ in {1..20}; do
  if curl -fsS "http://$DEPLOY_HOST/api/health" >/dev/null; then
    external_healthy=1
    break
  fi
  sleep 1
done
if [[ "$external_healthy" != "1" ]]; then
  echo "Public health check failed: http://$DEPLOY_HOST/api/health" >&2
  exit 1
fi
echo "Deployed successfully: http://$DEPLOY_HOST"
