#!/usr/bin/env bash
set -euo pipefail

SRC=".claude/claude.md"
DST=".github/copilot-instructions.md"

if [[ ! -f "$SRC" ]]; then
  echo "Missing source file: $SRC" >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

{
  echo "<!-- AUTO-GENERATED FROM .claude/claude.md. DO NOT EDIT DIRECTLY. -->"
  echo
  cat "$SRC"
} > "$tmp"

if [[ ! -f "$DST" ]] || ! cmp -s "$tmp" "$DST"; then
  mkdir -p "$(dirname "$DST")"
  mv "$tmp" "$DST"
  echo "Updated $DST from $SRC"
  echo "Please stage $DST and commit again."
  exit 1
fi

echo "$DST is already in sync with $SRC"
