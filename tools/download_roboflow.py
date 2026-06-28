#!/usr/bin/env python3
"""Download Roboflow Universe dataset (YOLOv8) to data/roboflow/."""
from __future__ import annotations
import os, sys
from pathlib import Path
ROOT = Path(__file__).resolve().parents[1]
ENV_FILE = ROOT / "data" / ".roboflow.env"
OUT_DIR = ROOT / "data" / "roboflow"
WORKSPACE = "visual-captioning-for-food"
PROJECT = "ingredients-detection-yolov8-npkkb"
VERSION = 3

def load_api_key() -> str:
    key = os.environ.get("ROBOFLOW_API_KEY", "").strip()
    if key:
        return key
    if ENV_FILE.is_file():
        for line in ENV_FILE.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if line.startswith("ROBOFLOW_API_KEY="):
                return line.split("=", 1)[1].strip()
    raise SystemExit(f"Missing ROBOFLOW_API_KEY in env or {ENV_FILE}")

def main() -> None:
    api_key = load_api_key()
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    try:
        from roboflow import Roboflow
    except ImportError:
        raise SystemExit("Install: pip install roboflow")
    rf = Roboflow(api_key=api_key)
    project = rf.workspace(WORKSPACE).project(PROJECT)
    dataset = project.version(VERSION).download(model_format="yolov8", location=str(OUT_DIR), overwrite=True)
    print(f"Downloaded to: {dataset.location}")

if __name__ == "__main__":
    main()
