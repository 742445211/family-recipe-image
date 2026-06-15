from pathlib import Path
p = Path("/www/server/panel/vhost/nginx/zzzjc.xin.conf")
text = p.read_text()
if "/api/ws/image-worker" in text:
    print("already configured")
    raise SystemExit(0)
block = """    # WebSocket image worker gateway
    location = /api/ws/image-worker {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
"""
needle = "    location = /api/ws {"
if needle not in text:
    raise SystemExit("anchor missing")
p.write_text(text.replace(needle, block + "\n" + needle))
print("patched")
