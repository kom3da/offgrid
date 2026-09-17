#!/usr/bin/env bash
# テンプレート（kom3da/offgrid）の指定した版から、カリキュラム部分だけを取り込む。
# 使い方: ./scripts/update-curriculum.sh v2026.10.1
# PROGRESS.md、logs/、answers/、capstone/ と、drills/ に自分で置いたファイルには触れない。
set -euo pipefail

# Wrapped in a function so that bash reads the whole file before scripts/ is replaced
main() {
  TEMPLATE="${TEMPLATE:-kom3da/offgrid}"
  tag="${1:?usage: update-curriculum.sh <tag>  (e.g. v2026.10.1)}"

  cd "$(dirname "$0")/.."

  if [ -n "$(git status --porcelain)" ]; then
    echo "error: commit or stash your changes first (git status shows pending changes)" >&2
    exit 1
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  curl -fsSL "https://github.com/${TEMPLATE}/archive/refs/tags/${tag}.tar.gz" |
    tar -xz -C "$tmp" --strip-components=1

  # Directories owned by the curriculum are replaced as a whole
  for dir in docs prompts templates scripts; do
    rm -rf "$dir"
    cp -R "$tmp/$dir" "$dir"
  done
  cp "$tmp/CLAUDE.md" "$tmp/README.md" "$tmp/CHANGELOG.md" .

  # Task sheets: only README.md files under drills/
  (cd "$tmp" && find drills -name README.md -print) | while IFS= read -r sheet; do
    mkdir -p "$(dirname "$sheet")"
    cp "$tmp/$sheet" "$sheet"
  done

  echo "Updated to ${tag}. Review the changes, then commit:"
  echo "  git status"
  echo "  git add -A && git commit -m \"[docs] Update curriculum to ${tag}\""
}

main "$@"
