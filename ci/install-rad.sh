#!/bin/bash
set -euo pipefail

# Install Rad for CI usage
# Downloads the latest release binary for Linux amd64

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="rad"
PLATFORM="linux_amd64"

echo "🔽 Installing Rad for CI..."

# Get the latest release URL
LATEST_RELEASE_URL="https://github.com/amterp/rad/releases/latest/download/rad_${PLATFORM}.tar.gz"

echo "📥 Downloading Rad binary from: $LATEST_RELEASE_URL"

# Download and extract
curl -fsSL "$LATEST_RELEASE_URL" | tar -xz --strip-components=0 -C /tmp

# Move to install directory
sudo mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify installation
echo "✅ Rad installed successfully!"
echo "📍 Location: ${INSTALL_DIR}/${BINARY_NAME}"
echo "🔍 Version: $(rad -v)"