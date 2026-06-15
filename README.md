# family-recipe-image

树莓派图片处理网关：通过 **WSS** 连接阿里云 [`family-recipe-server`](https://github.com/742445211/family-recipe-server)，执行图片压缩（mozjpeg / oxipng）与食材识别（CulinaryVision ONNX）。

配对仓库：

- 后端 API：[`742445211/family-recipe-server`](https://github.com/742445211/family-recipe-server)
- 微信小程序：[`742445211/family-recipe-miniapp`](https://github.com/742445211/family-recipe-miniapp)

## 架构

- **recipe-gateway**（Go 单进程）：WebSocket 客户端 + 压缩 + firdgemate 识别
- **family-recipe-server**：`ImageWorkerHub` 派发任务、接收结果、更新 DB

详见 [docs/ws-protocol.md](docs/ws-protocol.md)。

## 树莓派部署

### 1. 安装依赖

```bash
ssh raspberry 'bash -s' < deploy/install-deps.sh
```

### 2. 配置

```bash
sudo mkdir -p /opt/recipe-image/models
cp config.yaml.example /opt/recipe-image/config.yaml
chmod 600 /opt/recipe-image/config.yaml
# 填写 OSS 凭证、wss:// URL、gateway token（与 server image_worker.token 一致）
```

### 3. 模型

默认使用 [CulinaryVision-YOLOv8n](https://huggingface.co/HimanshuRay/CulinaryVision-YOLOv8n)（47 类食材）。详见 [docs/firdgemate-model.md](docs/firdgemate-model.md)。

```bash
# 开发机导出后复制到 Pi
scp culinaryvision.onnx raspberry:~/recipe-image/models/
```

### 4. 编译（在 Pi 上，需 CGO + ONNX Runtime）

```bash
cd /path/to/family-recipe-image
export CGO_ENABLED=1
export LD_LIBRARY_PATH=/usr/local/lib
go build -o recipe-gateway ./cmd/gateway
sudo install -m 755 recipe-gateway /opt/recipe-image/
```

### 5. systemd

```bash
sudo cp deploy/recipe-gateway.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now recipe-gateway
journalctl -u recipe-gateway -f
```

## 压缩规则

| 格式 | 处理 |
|------|------|
| JPG | mozjpeg |
| PNG | oxipng |
| 其他（非 GIF） | 转 PNG + oxipng，OSS key 改 `.png` |
| GIF | 跳过 |

## 开发

```bash
go test ./...
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o recipe-gateway ./cmd/gateway  # stub firdgemate
```

Mock WSS server（需 TLS 证书）：

```bash
GATEWAY_TOKEN=test-token TLS_CERT=cert.pem TLS_KEY=key.pem go run ./scripts/mock_ws_server.go
```

## SSH 别名

- 树莓派：`ssh raspberry`
- 阿里云：`ssh ali_24`
