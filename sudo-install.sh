#!/usr/bin/env bash
set -e
set -u
set -o pipefail

INSTALL_DIR="/usr/local/bin"

# Find the binary: use argument, or "grd" in cwd, or any grd-* match
if [ $# -ge 1 ] && [ -f "$1" ]; then
    BIN="$1"
elif [ -f grd ]; then
    BIN="grd"
else
    BIN=$(ls grd-* 2>/dev/null | head -1)
fi

if [ -z "${BIN:-}" ] || [ ! -f "$BIN" ]; then
    echo "Error: no grd binary found." >&2
    echo "Usage: ./sudo-install.sh [path-to-binary]" >&2
    exit 1
fi

echo "Installing $BIN to $INSTALL_DIR/grd ..."

sudo -v
sudo cp "$BIN" "$INSTALL_DIR/grd"
sudo chmod +x "$INSTALL_DIR/grd"

echo "Installed to $INSTALL_DIR/grd"
echo "Run 'grd integrate' to set up shell aliases."
