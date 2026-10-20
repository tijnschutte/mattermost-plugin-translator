#!/usr/bin/env bash
# Wait for the local Mattermost, then create a team and two users once:
# `admin` (system admin, the Dutch user) and `colleague` (the Portuguese user).
set -euo pipefail

cd "$(dirname "$0")"
mmctl() { docker compose exec -T mattermost mmctl --local "$@"; }

echo "Waiting for Mattermost on http://localhost:8065"
for _ in $(seq 1 90); do
  curl -fs http://localhost:8065/api/v4/system/ping >/dev/null && break
  sleep 2
done
curl -fs http://localhost:8065/api/v4/system/ping >/dev/null

if mmctl user search admin 2>/dev/null | grep -q '^id:'; then
  echo "Already seeded"
  exit 0
fi

mmctl team create --name dev --display-name Dev
mmctl user create --username admin --email admin@example.com --password 'Admin-dev-1!' --system-admin
mmctl user create --username colleague --email colleague@example.com --password 'Colleague-dev-1!'
mmctl team users add dev admin colleague
echo "Seeded: log in at http://localhost:8065 as admin / Admin-dev-1! or colleague / Colleague-dev-1!"
