# WebSocket 图片处理协议

树莓派 `recipe-gateway` 与 `family-recipe-server` 之间的通信协议。

## 传输安全

- **必须** 使用 `wss://`（TLS 1.2+）
- 禁止 `ws://` 明文连接
- Pi 端启动时校验 `server.ws_url` 以 `wss://` 开头
- Server 端校验 `X-Forwarded-Proto: https` 或直连 TLS
- Pi **不** 持有 MySQL 凭证；数据库更新由 server 在收到 `task_result` 后执行
- OSS 访问使用 HTTPS

## 端点

```
wss://www.zzzjc.xin/api/ws/image-worker?token=<GATEWAY_SECRET>&worker_id=pi-b4-001
```

| 参数 | 说明 |
|------|------|
| token | 与 server `image_worker.token` 一致的共享密钥 |
| worker_id | 可选，网关标识 |

## 消息类型

### registered（Server → Gateway）

连接成功后 server 发送：

```json
{ "type": "registered", "worker_id": "pi-b4-001" }
```

### task（Server → Gateway）

```json
{
  "type": "task",
  "task_id": "uuid",
  "action": "compress",
  "oss_key": "recipe/173000.webp",
  "oss_url": "https://cdn.example.com/recipe/173000.webp",
  "meta": { "recipe_id": 123 }
}
```

- `action`: `compress` | `recognize`
- `meta.recipe_id`: 强烈建议携带，用于 key 变更时更新 DB

### task_result（Gateway → Server）

压缩成功（格式转换）：

```json
{
  "type": "task_result",
  "task_id": "uuid",
  "status": "ok",
  "action": "compress",
  "oss_key": "recipe/173000.webp",
  "detail": {
    "skipped": false,
    "format": "webp",
    "output_format": "png",
    "new_oss_key": "recipe/173000.png",
    "original_bytes": 102400,
    "compressed_bytes": 45000
  },
  "meta": { "recipe_id": 123 }
}
```

GIF 跳过：

```json
{ "detail": { "skipped": true, "reason": "gif_not_supported" } }
```

识别成功：

```json
{
  "detail": { "ingredients": ["苹果", "胡萝卜"] },
  "meta": { "recipe_id": 123 }
}
```

### ping / pong

```json
{ "type": "ping" }
{ "type": "pong" }
```

## DB 同步（Server 侧）

当 `detail.new_oss_key` 存在且与 `oss_key` 不同时：

1. `UPDATE recipes SET image_key=?, cover_url=? WHERE id=? OR image_key=?`
2. `DeleteObject(old_key)` 清理 OSS

## 压缩策略

| 格式 | 处理 |
|------|------|
| JPG/JPEG | mozjpeg，同 key 覆盖 |
| PNG | oxipng，同 key 覆盖 |
| WebP/BMP/TIFF 等 | 转 PNG + oxipng，新 key `.png` |
| GIF | 跳过 |
