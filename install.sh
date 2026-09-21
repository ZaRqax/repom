#!/usr/bin/env bash
set -euo pipefail

REPO="ZaRqax/repom"
APP="repom"
VERSION="${REPOM_VERSION:-latest}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux) os=linux ;;
  darwin) os=darwin ;;
  *)
    echo "error: unsupported OS: $os" >&2
    echo "see https://github.com/$REPO/releases for manual downloads" >&2
    exit 1
    ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *)
    echo "error: unsupported architecture: $arch" >&2
    echo "see https://github.com/$REPO/releases for manual downloads" >&2
    exit 1
    ;;
esac

if [ -n "${REPOM_INSTALL_DIR:-}" ]; then
  install_dir="$REPOM_INSTALL_DIR"
elif [ -w /usr/local/bin ] 2>/dev/null; then
  install_dir=/usr/local/bin
else
  install_dir="$HOME/.local/bin"
fi

if [ "$VERSION" = "latest" ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  version="${VERSION#v}"
  base="https://github.com/$REPO/releases/download/v$version"
fi

asset="${APP}_${os}_${arch}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "downloading $asset"
curl -fsSL "$base/$asset" -o "$tmp/$asset"
tar -xzf "$tmp/$asset" -C "$tmp"

mkdir -p "$install_dir"
install -m 0755 "$tmp/$APP" "$install_dir/$APP"

echo "installed $APP to $install_dir/$APP"

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "note: add $install_dir to your PATH" ;;
esac
