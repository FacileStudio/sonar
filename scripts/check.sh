#!/usr/bin/env sh
#
# The repository quality gate. Reports, never rewrites (except --format).
#
#   sh scripts/check.sh             gofmt + vet + test, the web client, then filet
#   sh scripts/check.sh --go-only   the Go module only
#   sh scripts/check.sh --format    rewrite Go sources in place
#
# Deliberately depends on nothing but a `go` on PATH, plus `bun` for the client
# and `filet` for the style gate when they are installed. The server's database
# tests run when TEST_DATABASE_URL points at a throwaway PostgreSQL.
# It is NOT invoked through mise: `mise run` resolves every tool in the merged
# config before running any task body, so an unrelated broken tool in the
# user's global config would take this gate down with it.

set -eu

GO_MODULES="."

mode="all"
case "${1:-}" in
--go-only) mode="go" ;;
--format) mode="format" ;;
"") ;;
*)
  echo "usage: $0 [--go-only|--format]" >&2
  exit 2
  ;;
esac

cd "$(git rev-parse --show-toplevel)"

# Resolve the toolchain from GOROOT when it is set. mise exports GOROOT for the
# version this repo pins, but leaves an unrelated `go` earlier on PATH (Homebrew's,
# here), and a go binary driving a different GOROOT fails with
# `compile: version "X" does not match go tool version "Y"`.
if [ -z "${GO:-}" ]; then
  if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/go" ]; then GO="$GOROOT/bin/go"; else GO=go; fi
fi
if [ -z "${GOFMT:-}" ]; then
  if [ -n "${GOROOT:-}" ] && [ -x "$GOROOT/bin/gofmt" ]; then GOFMT="$GOROOT/bin/gofmt"; else GOFMT=gofmt; fi
fi

if ! command -v "$GO" >/dev/null 2>&1 && [ ! -x "$GO" ]; then
  echo "check: no usable go ('$GO')" >&2
  exit 1
fi

if [ "$mode" = "format" ]; then
  for dir in $GO_MODULES; do
    (cd "$dir" && "$GO" fmt ./...)
  done
  exit 0
fi

status=0

for dir in $GO_MODULES; do
  echo "==> $dir"
  (
    cd "$dir" || exit 1
    s=0

    unformatted="$("$GOFMT" -l . | grep -v '^vendor/' || true)"
    if [ -n "$unformatted" ]; then
      echo "gofmt: the following files are not formatted (run 'sh scripts/check.sh --format'):"
      echo "$unformatted"
      s=1
    fi

    "$GO" vet ./... || s=1
    "$GO" test ./... || s=1

    exit "$s"
  ) || status=1
done

if [ "$mode" = "all" ]; then
  echo "==> apps/client"
  (
    cd apps/client || exit 1
    if ! command -v bun >/dev/null 2>&1; then
      echo "check: no 'bun' on PATH, skipping the client" >&2
      exit 0
    fi
    [ -d node_modules ] || bun install --frozen-lockfile >/dev/null
    bun run check && bun run test
  ) || status=1

  echo "==> filet"
  if command -v filet >/dev/null 2>&1; then
    filet check . || status=1
  else
    echo "check: no 'filet' on PATH, skipping the style gate" >&2
  fi
fi

if [ "$status" -ne 0 ]; then
  echo "check failed"
fi
exit "$status"