#!/usr/bin/env bash
#
# sync-wpt fetches a handful of Web Platform Tests directories from a
# pinned upstream commit and writes them under wpt/.
#
# It checks out whole directories (encoding, common, resources) from a
# throwaway git clone: a cone-mode sparse-checkout limits the working tree
# to those directories, and a blob-filtered, depth-1 fetch of the single
# pinned commit means only the trees and blobs under them are ever
# downloaded. Fetching is git-only -- no curl, no jq, no config file to
# parse. Any repo-local path changes happen here, as a copy step after the
# checkout, rather than by asking git to check out into a different layout.
#
# Usage:
#   scripts/sync-wpt.sh [-r repo-root]
#
# repo-root defaults to the current directory.
#
# Requires: git.

set -euo pipefail

commit="f5cac385941da42024176a1763f3941e33bb7bb0"
repo_url="https://github.com/web-platform-tests/wpt.git"
target_dir="wpt"
dirs=(encoding common resources)

root="."

usage() {
  echo "Usage: $0 [-r repo-root]" >&2
  exit 1
}

while getopts "r:h" opt; do
  case "$opt" in
    r) root="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

if ! command -v git >/dev/null 2>&1; then
  echo "sync-wpt: 'git' is required but was not found in PATH" >&2
  exit 1
fi

if [ ! -d "$root" ]; then
  echo "sync-wpt: repo root not found: $root" >&2
  exit 1
fi
root="$(cd "$root" && pwd)"

clone_dir="$(mktemp -d)"
trap 'rm -rf "$clone_dir"' EXIT

git init -q "$clone_dir"
git -C "$clone_dir" remote add origin "$repo_url"
git -C "$clone_dir" sparse-checkout init --cone >/dev/null
git -C "$clone_dir" sparse-checkout set "${dirs[@]}"
git -C "$clone_dir" fetch -q --depth 1 --filter=blob:none origin "$commit"
git -C "$clone_dir" checkout -q FETCH_HEAD

target_root="$root/$target_dir"

for dir in "${dirs[@]}"; do
  src="$clone_dir/$dir"
  dst="$target_root/$dir"

  if [ ! -d "$src" ]; then
    echo "sync-wpt: expected directory not found after checkout: $dir" >&2
    exit 1
  fi

  rm -rf "$dst"
  mkdir -p "$(dirname "$dst")"
  cp -R "$src" "$dst"
  echo "synced: $dir -> $target_dir/$dir"
done
