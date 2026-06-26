#!/usr/bin/env python3
"""Generate Pi gateway config (run locally, pass PI_HOME env or default /home/zjc)."""
import os
from pathlib import Path
import yaml
import sys

pi_home = os.environ.get("PI_HOME", "/home/zjc")
server_cfg = Path(sys.argv[1]) if len(sys.argv) > 1 else None
if server_cfg and server_cfg.exists():
    server = yaml.safe_load(server_cfg.read_text())
else:
    server = yaml.safe_load(sys.stdin.read())
oss = server.get("oss", {})
iw = server.get("image_worker", {})
h = pi_home
print(f"""server:
  ws_url: "wss://www.zzzjc.xin/api/ws/image-worker"
  token: "{iw.get('token', 'CHANGE_ME')}"
  worker_id: "pi-b4-001"
  reconnect_sec: 5
  ping_interval_sec: 30

oss:
  endpoint: "{oss.get('endpoint', '')}"
  access_key_id: "{oss.get('access_key_id', '')}"
  access_key_secret: "{oss.get('access_key_secret', '')}"
  bucket: "{oss.get('bucket', '')}"
  custom_domain: "{oss.get('custom_domain', '')}"

worker:
  max_concurrent: 2
  temp_dir: "/tmp/recipe-image"

compress:
  oxipng_path: "/usr/local/bin/oxipng"
  oxipng_flags: ["-o", "4", "--strip", "safe"]
  oxipng_threads: 4
  cjpeg_path: "/usr/local/bin/cjpeg"
  djpeg_path: "/usr/local/bin/djpeg"
  cjpeg_quality: 85
  magick_path: "/usr/bin/magick"
  ffmpeg_path: "/usr/bin/ffmpeg"

firdgemate:
  model_path: "{h}/recipe-image/models/culinaryvision.onnx"
  onnx_lib_path: "{h}/local/lib/libonnxruntime.so"
  num_classes: 47
  conf_threshold: 0.30
  iou_threshold: 0.5
  input_size: 640
  intra_op_threads: 4
""")
