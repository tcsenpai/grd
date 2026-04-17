#!/usr/bin/env bash
set -e
set -u
set -o pipefail

INSTALL_DIR="$HOME/.local/bin"

echo "Installing grd to $INSTALL_DIR ..."

if [ ! -f grd ]; then
    echo "Error: 'grd' binary not found in current directory." >&2
    echo "Run build.sh first, then copy the appropriate binary here as 'grd'." >&2
    exit 1
fi

mkdir -p "$INSTALL_DIR"
cp grd "$INSTALL_DIR/grd"
chmod +x "$INSTALL_DIR/grd"

echo "Installed to $INSTALL_DIR/grd"
echo ""
echo "Make sure $INSTALL_DIR is in your PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "Then run 'grd integrate' to set up shell aliases."

