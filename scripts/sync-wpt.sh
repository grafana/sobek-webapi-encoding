#!/usr/bin/env bash
#
# sync-wpt fetches the Web Platform Tests files listed in wpt.json from the
# pinned upstream commit and writes them under target_dir, applying any
# configured patch afterward.
#
# Usage:
#   scripts/sync-wpt.sh [-c wpt.json] [-r repo-root]
#
# Requires: curl, jq, and git (only needed for entries with a "patch" field).
# Fetches from raw.githubusercontent.com by default; override the base URL
# with SYNC_WPT_RAW_BASE_URL (mainly useful for testing against a local
# file:// or http:// mirror).

set -euo pipefail

config="wpt.json"
root="."

usage() {
  echo "Usage: $0 [-c config] [-r repo-root]" >&2
  exit 1
}

while getopts "c:r:h" opt; do
  case "$opt" in
    c) config="$OPTARG" ;;
    r) root="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

for bin in curl jq git; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "sync-wpt: '$bin' is required but was not found in PATH" >&2
    exit 1
  fi
done

config_path="$root/$config"
if [ ! -f "$config_path" ]; then
  echo "sync-wpt: config not found: $config_path" >&2
  exit 1
fi

commit=$(jq -r '.commit // ""' "$config_path")
target_dir=$(jq -r '.target_dir // ""' "$config_path")

if [ -z "$commit" ]; then
  echo "sync-wpt: $config_path: missing required \"commit\" field" >&2
  exit 1
fi
if [ -z "$target_dir" ]; then
  echo "sync-wpt: $config_path: missing required \"target_dir\" field" >&2
  exit 1
fi

raw_base_url="${SYNC_WPT_RAW_BASE_URL:-https://raw.githubusercontent.com/web-platform-tests/wpt}"

file_count=$(jq '.files | length' "$config_path")

for ((i = 0; i < file_count; i++)); do
  entry=$(jq -c ".files[$i]" "$config_path")
  src=$(echo "$entry" | jq -r '.src')
  # dst defaults to src -- only entries that rename the fetched file (e.g.
  # stripping ".any" from WPT's *.any.js test files) need to set it.
  dst=$(echo "$entry" | jq -r '.dst // .src')
  patch=$(echo "$entry" | jq -r '.patch // ""')

  dst_path="$root/$target_dir/$dst"
  mkdir -p "$(dirname "$dst_path")"

  url="$raw_base_url/$commit/$src"
  if ! curl -fsSL "$url" -o "$dst_path"; then
    echo "sync-wpt: failed to fetch $url" >&2
    exit 1
  fi

  if [ -n "$patch" ]; then
    if ! git -C "$root" apply "$patch" </dev/null; then
      echo "sync-wpt: failed to apply patch $patch to $dst_path" >&2
      exit 1
    fi
  fi

  echo "synced: $src -> $target_dir/$dst"
done
