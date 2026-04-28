#!/usr/bin/env bash
# Mirrors GitHub Actions + GitLab CI: build, lint (golangci-lint v1.64.8), test.
set -euo pipefail
ROOT="${ROOT:-/workspace}"
cd "$ROOT"
export GOPATH="${GOPATH:-$ROOT/.go}"
export GOCACHE="${GOCACHE:-$ROOT/.cache/go-build}"
mkdir -p "$GOPATH" "$GOCACHE"

echo "=== go version ==="
go version

echo "=== BUILD ==="
cd "$ROOT/go-init-common" && go build ./...
cd "$ROOT/go-init-manager" && go build -o /dev/null ./cmd
cd "$ROOT/go-init-generator" && go build ./...
cd "$ROOT/go-init-publisher" && go build ./...
echo "BUILD_OK"

echo "=== INSTALL golangci-lint v1.64.8 ==="
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.64.8
export PATH="$PATH:$(go env GOPATH)/bin"
golangci-lint version

echo "=== LINT ==="
for d in go-init-common go-init-manager go-init-generator go-init-publisher; do
  echo "--- lint $d ---"
  (cd "$ROOT/$d" && golangci-lint run ./...)
done
echo "LINT_OK"

echo "=== TEST ==="
for d in go-init-common go-init-manager go-init-generator go-init-publisher; do
  echo "--- test $d ---"
  (cd "$ROOT/$d" && go test ./... -count=1 -short)
done
echo "TEST_OK"
echo "=== ALL CI STEPS PASSED ==="
