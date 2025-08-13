#!/bin/bash
set -euo pipefail

# source
HOOKS_DIR="$(dirname "$0")/git-hooks"

# target
GIT_HOOKS_DIR="$(git rev-parse --show-toplevel)/.git/hooks"

mkdir -p "$GIT_HOOKS_DIR"

for hook in "$HOOKS_DIR"/*; do
    [ -e "$hook" ] || continue
    echo "Installing $(basename "$hook") hook..."
    cp "$hook" "$GIT_HOOKS_DIR/$(basename "$hook")"
    chmod +x "$GIT_HOOKS_DIR/$(basename "$hook")"
done

echo "Git hooks have been set up."

