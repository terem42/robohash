#!/bin/bash

if ! command -v podman &> /dev/null; then
    echo "❌ Podman is not installed."
    echo "Please install it with command:"
    echo "  sudo apt-get update && sudo apt-get install podman"
    exit 1
else
    echo "✅ Podman installed: $(podman --version)"
fi

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -x

cd "$workdir" || exit 1

build_version=$(git describe --tags --always 2>/dev/null || echo "v0.0.1")

rm -rf app-bin/*

podman build --build-arg BUILD_VERSION="$build_version" \
  -t "ghcr.io/terem42/robohash:$build_version" .
