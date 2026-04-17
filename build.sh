#!/bin/sh
set -eu

mkdir -p dist

echo "Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o dist/grd-linux-amd64 .

echo "Building linux/arm64..."
GOOS=linux GOARCH=arm64 go build -o dist/grd-linux-arm64 .

echo "Building darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -o dist/grd-macos-amd64 .

echo "Building darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -o dist/grd-macos-arm64 .

echo "Building windows/amd64..."
GOOS=windows GOARCH=amd64 go build -o dist/grd-windows-amd64.exe .

echo ""
echo "Build complete. Artifacts:"
ls -lh dist/
