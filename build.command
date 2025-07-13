#!/bin/bash

echo
echo "========================"

# ins Verzeichnis wechseln
cd "$(dirname "$0")"

# Kompilieren für Linux
echo "Kompilieren für Linux"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w" -o programme/klassenmischer-linux-amd64 .
echo

# Kompilieren für Linux
echo "Kompilieren für Linux"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64
go build -ldflags="-s -w" -o programme/klassenmischer-linux-arm64 .
echo

# Kompilieren für Apple Silicon
echo "Kompilieren für Apple Silicon"
export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=arm64
go build -ldflags="-s -w" -o programme/klassenmischer-macos-silicon .
echo

# Kompilieren für Apple Intel
echo "Kompilieren für Apple Intel"
export CGO_ENABLED=0
export GOOS=darwin
export GOARCH=amd64
go build -ldflags="-s -w" -o programme/klassenmischer-macos-intel .
echo

# Kompilieren für Windows
echo "Kompilieren für Windows"
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
go build -ldflags="-s -w" -o programme/klassenmischer-win-64.exe .
echo