#!/usr/bin/env sh
set -eu
sed -n 's/^## \[\(v[0-9][^]]*\)\].*/\1/p' "${1:-CHANGELOG.md}" | head -n 1
