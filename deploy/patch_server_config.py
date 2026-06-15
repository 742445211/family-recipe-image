from pathlib import Path
p = Path("/root/projects/recipe-server/config.yaml")
text = p.read_text()
if "image_worker:" not in text:
    text = text.rstrip() + """

image_worker:
  enabled: true
  path: "/api/ws/image-worker"
  token: "recipe-pi-gateway-2026"
  ping_interval_sec: 30
  read_timeout_sec: 120
"""
    p.write_text(text + "\n")
    print("config added")
else:
    print("config exists")
