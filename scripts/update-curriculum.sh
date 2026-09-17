#!/usr/bin/env bash
# テンプレート（kom3da/offgrid）の指定した版から、カリキュラム部分だけを取り込む。
# 使い方: ./scripts/update-curriculum.sh <版>   （版の一覧は CHANGELOG.md）
# 置き換えるのは docs/ prompts/ templates/ scripts/ tools/ CLAUDE.md README.md CHANGELOG.md と、
# drills/ の中の TASKS.md（課題シート）だけ。それ以外の自分のファイルには触れない。
set -euo pipefail

# Wrapped in a function so that bash reads the whole file before scripts/ is replaced
main() {
  TEMPLATE="${TEMPLATE:-kom3da/offgrid}"
  tag="${1:?usage: update-curriculum.sh <tag>  (see CHANGELOG.md for the list)}"
  dirs=(docs prompts templates scripts tools)
  files=(CLAUDE.md README.md CHANGELOG.md)

  cd "$(dirname "$0")/.."

  if [ -n "$(git status --porcelain)" ]; then
    echo "error: commit or stash your changes first (git status shows pending changes)" >&2
    exit 1
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  if ! curl -fsSL "https://github.com/${TEMPLATE}/archive/refs/tags/${tag}.tar.gz" |
    tar -xz -C "$tmp" --strip-components=1; then
    echo "error: could not download ${tag} from ${TEMPLATE}. Check the tag name in CHANGELOG.md" >&2
    exit 1
  fi

  # Check the download before touching anything, so that a bad or very old tag changes nothing
  for path in "${dirs[@]}" "${files[@]}" scripts/update-curriculum.sh; do
    if [ ! -e "$tmp/$path" ]; then
      echo "error: ${tag} does not contain ${path}; nothing was changed" >&2
      exit 1
    fi
  done

  for dir in "${dirs[@]}"; do
    rm -rf "$dir"
    cp -R "$tmp/$dir" "$dir"
  done
  for file in "${files[@]}"; do
    cp "$tmp/$file" "$file"
  done

  # Task sheets: only TASKS.md files under drills/
  (cd "$tmp" && find drills -name TASKS.md -print) | while IFS= read -r sheet; do
    mkdir -p "$(dirname "$sheet")"
    cp "$tmp/$sheet" "$sheet"
  done

  # Task sheets that no longer exist in the template are reported, not deleted
  find drills -name TASKS.md -print | while IFS= read -r sheet; do
    [ -e "$tmp/$sheet" ] || echo "note: ${sheet} is no longer part of the template (kept as is)"
  done

  echo "Updated to ${tag}. Review the changes, then commit:"
  echo "  git diff --stat"
  echo "  git add -A && git commit -m \"[docs] Update curriculum to ${tag}\""
  echo "To undo instead: git restore . && git clean -fd docs prompts templates scripts tools"

  # The binary is installed once on Day 0, so a MAJOR bump needs an explicit swap:
  # the sheet format changes and the old tool cannot read it.
  new_major="$(printf '%s' "${tag#v}" | cut -d. -f1)"
  old_major="$(offgrid version 2>/dev/null | awk '{print $2}' | cut -d. -f1 || true)"
  if [ -n "$old_major" ] && [ "$old_major" != "dev" ] && [ "$new_major" != "$old_major" ]; then
    echo
    echo "The major version changed (${old_major} -> ${new_major}): replace the tool as well."
    echo "  case \"\$(uname -m)\" in aarch64|arm64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; esac"
    echo "  curl -fsSL -o ~/.local/bin/offgrid \\"
    echo "    \"https://github.com/${TEMPLATE}/releases/download/${tag}/offgrid-linux-\$arch\""
    echo "  chmod +x ~/.local/bin/offgrid && offgrid selftest"
  fi
}

main "$@"
