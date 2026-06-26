# firdgemate 模型说明

## 当前部署模型

项目使用 **[CulinaryVision-YOLOv8n](https://huggingface.co/HimanshuRay/CulinaryVision-YOLOv8n)**（MIT 许可）：

- **47 类常见食材**（苹果、番茄、胡萝卜、鸡蛋等）
- 已导出为 ONNX：`models/culinaryvision.onnx`
- 配置项 `firdgemate.num_classes: 47`

## 配置（config.yaml）

```yaml
firdgemate:
  model_path: "/home/zjc/recipe-image/models/culinaryvision.onnx"
  onnx_lib_path: "/home/zjc/local/lib/libonnxruntime.so"
  num_classes: 47          # 必须与模型类别数一致
  conf_threshold: 0.30
  iou_threshold: 0.5
  input_size: 640
  intra_op_threads: 4
```

更换模型后务必同步修改 `num_classes` 和 `internal/firdgemate/labels.go` 中的类别列表。

---

## 如何训练自定义模型

### 1. 准备数据集

推荐 [Roboflow](https://universe.roboflow.com/) 或自建标注：

- 格式：**YOLOv8**（每张图对应 `.txt` 标注 + `data.yaml`）
- 示例公开集：[Ingredients detection YOLOv8](https://universe.roboflow.com/visual-captioning-for-food/ingredients-detection-yolov8-npkkb)（12547 张）

`data.yaml` 示例：

```yaml
path: /path/to/dataset
train: images/train
val: images/val
names:
  0: tomato
  1: egg
  # ...
```

### 2. 训练（开发机，需 GPU 更佳）

```bash
pip install ultralytics -i https://mirrors.aliyun.com/pypi/simple/
yolo detect train model=yolov8n.pt data=data.yaml epochs=100 imgsz=640 batch=16 device=0
```

输出权重：`runs/detect/train/weights/best.pt`   

### 3. 导出 ONNX（树莓派推理用）

```bash
yolo export model=runs/detect/train/weights/best.pt format=onnx simplify=True imgsz=640 opset=12 dynamic=False
```

将生成的 `best.onnx` 复制到 Pi：

```bash
scp best.onnx raspberry:~/recipe-image/models/my-model.onnx
```

### 4. 更新网关配置与标签

1. 查看类别名：

```python
from ultralytics import YOLO
print(YOLO("best.pt").names)
```

1. 更新 `[internal/firdgemate/labels.go](../internal/firdgemate/labels.go)` 中的 `culinaryLabels` 与 `englishToChinese` 映射
2. 修改 `config.yaml`：

```yaml
firdgemate:
  model_path: "/home/zjc/recipe-image/models/my-model.onnx"
  num_classes: <你的类别数>
```

1. 在 Pi 上重新编译并重启：

```bash
export GOPROXY=https://goproxy.cn,direct CGO_ENABLED=1 LD_LIBRARY_PATH=$HOME/local/lib
go build -o recipe-gateway ./cmd/gateway
systemctl --user restart recipe-gateway
```

### 5. 树莓派性能建议


| 模型       | 参数量  | Pi 4 推理速度 | 推荐     |
| -------- | ---- | --------- | ------ |
| yolov8n  | ~3M  | ~1-3 s/张  | **推荐** |
| yolov8s  | ~11M | ~3-6 s/张  | 精度更高时  |
| yolov8m+ | 更大   | 较慢        | 不推荐    |


训练时使用 `yolov8n.pt` 作为基座，导出 `imgsz=640`，Pi 上 `intra_op_threads: 4`。

---

## 备选基线模型

若无自定义数据，可继续使用：


| 模型                         | 说明                                              |
| -------------------------- | ----------------------------------------------- |
| **CulinaryVision-YOLOv8n** | 47 类食材，**当前使用**                                 |
| yolov8n (COCO)             | 80 类通用物体，仅部分为食物，需 `num_classes: 80` 并恢复 COCO 标签 |


COCO 模型导出：

```bash
yolo export model=yolov8n.pt format=onnx simplify=True imgsz=640
```

