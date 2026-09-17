#!/usr/bin/env bash
# GitHub に private リポジトリを作成して push する（初回のみ）。
# 前提: gh CLI に OWNER のアカウントでログイン済み（アクティブでなくてよい）、
#       ~/.ssh/config に OWNER の鍵を使う Host エイリアス（SSH_HOST）がある。
# gh のアクティブアカウントと git のグローバル設定は変更しない。
set -euo pipefail

OWNER="${OWNER:-kom3da}"
REPO="${REPO:-offgrid}"
SSH_HOST="${SSH_HOST:-github-kom3da}"
GIT_NAME="${GIT_NAME:-kom3da}"
GIT_EMAIL="${GIT_EMAIL:-me@kom3da.dev}"
VERSION="${VERSION:-v2026.09}"

cd "$(dirname "$0")/.."

if [ ! -d .git ]; then
  git init -b main
  # Commit identity for this repository only
  git config --local user.name "$GIT_NAME"
  git config --local user.email "$GIT_EMAIL"
  git add -A
  git commit -m "[ai] Initialize learning package ${VERSION}"
  git tag "$VERSION"
fi

# Use OWNER's token for this single command instead of `gh auth switch`
GH_TOKEN="$(gh auth token --user "$OWNER")" gh repo create "${OWNER}/${REPO}" \
  --private \
  --description "ゼロから学び直す：Linux・Git・ネットワーク・Go・TypeScript・PostgreSQL（AIなしデー・閉域開発・監視ダッシュボード）"

git remote add origin "git@${SSH_HOST}:${OWNER}/${REPO}.git"
git push -u origin main --tags

echo "Done: https://github.com/${OWNER}/${REPO}"
