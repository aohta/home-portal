#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"
version="${VERSION:-dev}"
output="${OUTPUT:-dist/home-portal}"
if [[ ! "${version}" =~ ^[A-Za-z0-9.+_-]+$ ]]; then
  echo 'VERSION must contain only letters, numbers, dots, +, _ or -' >&2
  exit 1
fi

if [[ ! -f apps/web/node_modules/.package-lock.json || apps/web/package-lock.json -nt apps/web/node_modules/.package-lock.json || apps/web/package.json -nt apps/web/node_modules/.package-lock.json || apps/web/.npmrc -nt apps/web/node_modules/.package-lock.json ]]; then
  (cd apps/web && npm ci --ignore-scripts)
fi
(cd apps/web && npm run build)
node scripts/copy-assets.mjs
mkdir -p "$(dirname "${output}")"
CGO_ENABLED=0 go build -trimpath -tags embedded \
  -ldflags "-X github.com/aohta/home-portal/internal/command.Version=${version}" \
  -o "${output}" ./cmd/home-portal
echo "built ${output} (Vue frontend included)"
