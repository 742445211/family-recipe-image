#!/usr/bin/env python3
"""查询 fridge_scans 最近记录。在 ali_24 上运行：python3 tools/query_scans.py"""

import subprocess
from pathlib import Path

import yaml

cfg = yaml.safe_load(Path("/root/projects/recipe-server/config.yaml").read_text())
m = cfg["mysql"]
q = "SELECT id,task_id,status,image_key FROM fridge_scans ORDER BY id DESC LIMIT 5"
cmd = [
    "mysql",
    "-h",
    m["host"],
    "-P",
    str(m.get("port", 3306)),
    "-u",
    m["user"],
    f"-p{m['password']}",
    m["database"],
    "-N",
    "-e",
    q,
]
print(subprocess.check_output(cmd, text=True))
