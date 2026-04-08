#!/bin/bash
set -e

REPO="coffeegraph/coffeegraph"
VERSION="${COFFEEGRAPH_VERSION:-latest}"
INSTALL_DIR="${COFFEEGRAPH_INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64)       ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "☕ Installing CoffeeGraph..."
echo "   OS: $OS  ARCH: $ARCH"

if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/coffeegraph_${OS}_${ARCH}.tar.gz"
else
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/v${VERSION}/coffeegraph_${OS}_${ARCH}.tar.gz"
fi

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

echo "   Downloading from $DOWNLOAD_URL"
curl -fsSL "$DOWNLOAD_URL" | tar -xz -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/coffeegraph" "$INSTALL_DIR/coffeegraph"
  chmod +x "$INSTALL_DIR/coffeegraph"
else
  sudo mv "$TMP/coffeegraph" "$INSTALL_DIR/coffeegraph"
  sudo chmod +x "$INSTALL_DIR/coffeegraph"
fi

echo ""
echo "✓ CoffeeGraph installed at $INSTALL_DIR/coffeegraph"
echo ""
echo "Next steps:"
echo "  coffeegraph init my-agency"
echo "  cd my-agency"
echo "  coffeegraph add sales-closer"
echo "  coffeegraph dashboard"
echo ""
echo "Docs: https://coffeegraph.dev"
