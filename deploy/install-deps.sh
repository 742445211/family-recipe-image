#!/usr/bin/env bash
# Raspberry Pi (Debian aarch64) dependency installer for recipe-gateway.
#
# Run interactively on the Pi (sudo password required):
#   ssh raspberry
#   sudo bash /tmp/install-deps.sh
#
# Or from dev machine after scp:
#   scp deploy/install-deps.sh raspberry:/tmp/
#   ssh raspberry   # then: sudo bash /tmp/install-deps.sh
set -euo pipefail

PREFIX="${PREFIX:-/usr/local}"
OPT_DIR="${OPT_DIR:-/opt/recipe-image}"
ONNX_VERSION="${ONNX_VERSION:-1.20.1}"

echo "==> Installing apt packages..."
export DEBIAN_FRONTEND=noninteractive
sudo apt-get update -qq
sudo apt-get install -y --no-install-recommends \
  build-essential cmake git curl ca-certificates \
  imagemagick webp pkg-config \
  nasm yasm

echo "==> Installing oxipng..."
if ! command -v oxipng >/dev/null 2>&1; then
  if command -v cargo >/dev/null 2>&1; then
    cargo install oxipng --locked
    sudo ln -sf "$HOME/.cargo/bin/oxipng" "$PREFIX/bin/oxipng"
  else
    OXIPNG_URL="https://github.com/shssoichiro/oxipng/releases/download/v9.1.4/oxipng-9.1.4-aarch64-unknown-linux-musl.tar.gz"
    TMP=$(mktemp -d)
    curl -fsSL "$OXIPNG_URL" -o "$TMP/oxipng.tgz"
    tar -xzf "$TMP/oxipng.tgz" -C "$TMP"
    sudo install -m 755 "$TMP"/oxipng "$PREFIX/bin/oxipng"
    rm -rf "$TMP"
  fi
fi
oxipng --version || oxipng -V

echo "==> Building mozjpeg..."
if ! command -v cjpeg >/dev/null 2>&1; then
  MOZJPEG_DIR=$(mktemp -d)
  git clone --depth 1 --branch v4.1.5 https://github.com/mozilla/mozjpeg.git "$MOZJPEG_DIR"
  cmake -S "$MOZJPEG_DIR" -B "$MOZJPEG_DIR/build" -DCMAKE_INSTALL_PREFIX="$PREFIX"
  cmake --build "$MOZJPEG_DIR/build" -j"$(nproc)"
  sudo cmake --install "$MOZJPEG_DIR/build"
  rm -rf "$MOZJPEG_DIR"
fi
cjpeg -version 2>&1 | head -1
djpeg -version 2>&1 | head -1

echo "==> Installing ONNX Runtime (aarch64)..."
sudo mkdir -p "$PREFIX/lib"
if ! ls "$PREFIX/lib"/libonnxruntime.so* >/dev/null 2>&1; then
  ONNX_TGZ="onnxruntime-linux-aarch64-${ONNX_VERSION}.tgz"
  ONNX_URL="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${ONNX_TGZ}"
  TMP=$(mktemp -d)
  curl -fsSL "$ONNX_URL" -o "$TMP/onnx.tgz"
  tar -xzf "$TMP/onnx.tgz" -C "$TMP"
  sudo cp "$TMP"/onnxruntime-linux-aarch64-"${ONNX_VERSION}"/lib/libonnxruntime.so.* "$PREFIX/lib/"
  sudo ldconfig
  rm -rf "$TMP"
fi
ls -la "$PREFIX/lib"/libonnxruntime.so*

echo "==> Preparing model directory..."
sudo mkdir -p "$OPT_DIR/models"
if [[ ! -f "$OPT_DIR/models/yolov8n.onnx" ]]; then
  echo "WARNING: $OPT_DIR/models/yolov8n.onnx not found."
  echo "Export on dev machine: pip install ultralytics && yolo export model=yolov8n.pt format=onnx simplify=True"
  echo "Then: scp yolov8n.onnx raspberry:$OPT_DIR/models/"
fi

magick -version | head -1
ffmpeg -version 2>&1 | head -1
go version

echo "==> All dependencies installed."
