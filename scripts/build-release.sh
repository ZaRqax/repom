#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?usage: build-release.sh <version>}"
APP="repom"
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

cd "$(dirname "$0")/.."
DIST="$(pwd)/${DIST_DIR:-dist}"

rm -rf "$DIST"
mkdir -p "$DIST"

for platform in $PLATFORMS; do
  os="${platform%%/*}"
  arch="${platform##*/}"
  bin="$APP"
  if [ "$os" = "windows" ]; then
    bin="$APP.exe"
  fi

  name="${APP}_${os}_${arch}"
  stage="$DIST/$name"
  mkdir -p "$stage"

  echo "building $name"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" \
    -o "$stage/$bin" .

  if [ "$os" = "windows" ]; then
    (cd "$stage" && zip -q -r "$DIST/$name.zip" "$bin")
  else
    tar -czf "$DIST/$name.tar.gz" -C "$stage" "$bin"
  fi
  rm -rf "$stage"
done

(cd "$DIST" && sha256sum *.tar.gz *.zip >checksums.txt)

echo "artifacts in $DIST:"
ls -1 "$DIST"
