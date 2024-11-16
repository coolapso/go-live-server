#!/bin/bash

## check for root privileges
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

# Variables
REPO="coolapso/go-live-server"
VERSION=${VERSION:-"latest"}
INSTALL_DIR="/usr/local/bin"

# Determine OS and Architecture
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]]; then
    ARCH="arm64"
elif [[ "$ARCH" == "i386" || "$ARCH" == "i686" ]]; then
    ARCH="386"
fi

# Fetch the latest release if no version is specified
if [[ "$VERSION" == "latest" ]]; then
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

# Download URL for the release file and checksum
FILE="go-live-server_${VERSION:1}_${OS}_${ARCH}.tar.gz"
FILE_URL="https://github.com/$REPO/releases/download/$VERSION/$FILE"

# Download the file and checksum
if ! curl -LO "$FILE_URL"; then
  echo "Failed to download $FILE_URL"
  exit 1
fi

# Extract and install
if ! tar xzf "$FILE"; then
  echo "Failed to extract $FILE"
  exit 1
fi

## Install the binary
if ! chmod +x live-server; then
  echo "Failed to make live-server executable"
  exit 1
fi

if ! mv live-server "$INSTALL_DIR" ; then
  echo "Failed to move live-server to $INSTALL_DIR"
  exit 1
fi

# Cleanup
if ! rm "$FILE"; then
  echo "Failed to remove $FILE"
  exit 1
fi

echo "Installation of $REPO version $VERSION complete."
