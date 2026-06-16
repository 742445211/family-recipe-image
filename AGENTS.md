# AGENTS.md — family-recipe-image

树莓派图片处理网关。配对后端仓库：`742445211/family-recipe-server`、前端：`742445211/family-recipe-miniapp`。

## 技术栈

- Go 1.24+、gorilla/websocket、阿里云 OSS SDK
- ONNX Runtime + CulinaryVision-YOLOv8n（firdgemate 食材识别）
- 压缩：mozjpeg（JPG）、oxipng（PNG）、ImageMagick（格式转换）
- 模块名：`recipe-image`
- 与 server 通信：**WSS** `wss://www.zzzjc.xin/api/ws/image-worker`

## 目录结构

```
cmd/gateway/           # 入口
internal/
  ws/                  # WSS 客户端（强制 TLS）
  dispatcher/          # 任务调度
  worker/              # compress / recognize
  compress/            # mozjpeg / oxipng / 转 PNG
  firdgemate/          # ONNX 推理
  oss/                 # OSS 下载/上传
  protocol/            # WebSocket 消息类型
deploy/                # install-deps.sh、systemd、nginx 补丁
docs/                  # ws-protocol.md、firdgemate-model.md
scripts/               # 测试/调试脚本（detect_*、mock WSS server）
tools/                 # 正式辅助工具（OSS 识别、DB 查询等）
```

## 变更部署流程

**先在本机提交并推送到 GitHub，再去树莓派拉取更新。** 不要直接在 Pi 上改代码而不回传仓库。

1. **本机**（`F:\AI_Project\family-recipe-image`）：
   ```bash
   git add -A && git commit -m "..." && git push origin master
   ```
2. **树莓派**（`ssh raspberry`，代码目录 `~/recipe-image`）：
   ```bash
   cd ~/recipe-image && git pull origin master
   export CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct
   export LD_LIBRARY_PATH=$HOME/local/lib:/usr/local/lib
   go build -o recipe-gateway ./cmd/gateway
   systemctl --user restart recipe-gateway
   journalctl --user -u recipe-gateway -f
   ```

## 部署

- 树莓派：`ssh raspberry`，用户级 systemd `recipe-gateway.service`
- 阿里云 server 需配置 `image_worker` 块并 Nginx 代理 `/api/ws/image-worker`
- 敏感配置：`config.yaml`（勿提交，参考 `config.yaml.example`）

## 相关文档

- [docs/ws-protocol.md](docs/ws-protocol.md) — 与 server 的 WSS 协议
- [docs/firdgemate-model.md](docs/firdgemate-model.md) — 模型与训练说明
