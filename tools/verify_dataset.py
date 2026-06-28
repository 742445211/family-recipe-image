#!/usr/bin/env python3
"""Verify YOLO dataset under data/roboflow/."""
from pathlib import Path
import yaml

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "data" / "roboflow"
yaml_path = DATA / "data.yaml"

def main():
    if not yaml_path.is_file():
        raise SystemExit(f"missing {yaml_path}")
    cfg = yaml.safe_load(yaml_path.read_text(encoding="utf-8"))
    names = cfg.get("names") or []
    nc = cfg.get("nc")
    print(f"nc={nc} names={len(names)}")
    if "egg" in names:
        print(f"egg class id: {names.index('egg')}")
    for split in ("train", "valid", "test"):
        img_dir = DATA / split / "images"
        if img_dir.is_dir():
            n = len(list(img_dir.glob("*.*")))
            print(f"{split}: {n} images")
    total = sum(len(list((DATA/s/"images").glob("*.*"))) for s in ("train","valid","test") if (DATA/s/"images").is_dir())
    size_gb = sum(f.stat().st_size for f in DATA.rglob("*") if f.is_file()) / (1024**3)
    print(f"total images: {total}  disk: {size_gb:.2f} GB")

if __name__ == "__main__":
    main()
