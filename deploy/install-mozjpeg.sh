#!/bin/bash
set -euo pipefail
MOZ=/tmp/mozjpeg-build
rm -rf "$MOZ"
git clone --depth 1 --branch v4.1.5 https://github.com/mozilla/mozjpeg.git "$MOZ"
cmake -S "$MOZ" -B "$MOZ/build" -DCMAKE_INSTALL_PREFIX=/usr/local
cmake --build "$MOZ/build" -j"$(nproc)"
cmake --install "$MOZ/build"
echo "=== versions ==="
cjpeg -version 2>&1 | head -1
djpeg -version 2>&1 | head -1
magick -version | head -1
