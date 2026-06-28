# Roboflow 公开数据集下载指南（$0，不上传私有数据）

本指南只用于从 **Roboflow Universe** 下载**公开** YOLO 数据集，供本机微调使用。  
**不要**把家庭照片、OSS 菜谱图等私有数据上传到 Roboflow Public 项目（Public 计划会公开到 Universe）。

---

## 目标数据集

- **名称**：Ingredients detection YOLOv8
- **Universe 链接**：https://universe.roboflow.com/visual-captioning-for-food/ingredients-detection-yolov8-npkkb
- **规模**：约 12,547 张（以页面显示为准）
- **格式**：YOLOv8
- **本机落盘目录**：`F:\AI_Project\family-recipe-image\data\roboflow\`

---

## 费用说明

| 操作 | 是否收费 |
|------|----------|
| 注册 Public 免费账号 | 否 |
| 浏览 / 下载 Universe **公开**数据集 | 否 |
| 在本机 `yolo train` | 否（用本机 GPU） |
| 上传私有项目、云端 Train、Deploy API | 是（Credits / Core 套餐） |

---

## 前置条件

1. 浏览器可访问 Roboflow（若直连慢，开系统代理 `http://192.168.1.88:7890`）
2. 本机已创建目录：

```powershell
mkdir F:\AI_Project\family-recipe-image\data\roboflow
```

3. `data/` 已在 `.gitignore`，**不会**被 git 提交

---

## 操作步骤

### 1. 注册账号（无需信用卡）

1. 打开 https://roboflow.com/
2. 点击 **Sign Up**，用邮箱或 GitHub 注册
3. 选择 **Public** 免费计划（个人/研究用途即可）
4. **不要**在此步骤绑定信用卡，除非你要用 Core 私有项目

> 仅下载公开集时，用 Public 计划足够。

### 2. 打开目标公开数据集

1. 浏览器访问：  
   https://universe.roboflow.com/visual-captioning-for-food/ingredients-detection-yolov8-npkkb
2. 确认页面标注为 **Public / Open Source**（公开数据集）
3. 浏览 **Classes**，确认包含 `egg` 等你需要的类别

### 3. 下载数据集（关键步骤）

1. 在数据集页面点击 **Download Dataset**（或 **Download**）
2. **Format** 选择：**YOLOv8**
3. 选择 split：
   - 优先选 **train / valid / test** 完整包（若有「Download zip」含全部 split 选它）
   - 不要选「COCO」「TensorFlow」等其他格式
4. 点击下载，得到 zip（文件名类似 `ingredients-detection-yolov8-npkkb.v*.zip`）

若页面要求 **Fork to Workspace** 再下载：

1. 点 **Fork Dataset** →  fork 到你自己的 **Public** workspace（数据仍公开，但可导出）
2. Fork 完成后进入该项目 → **Export** → 格式选 **YOLOv8** → **Download**

### 4. 解压到本机目录

```powershell
# 将下载的 zip 路径替换为实际路径
Expand-Archive -Path "$env:USERPROFILE\Downloads\*.zip" -DestinationPath "F:\AI_Project\family-recipe-image\data\roboflow" -Force
```

解压后目录应类似：

```
data/roboflow/
├── data.yaml          # 或 train/data.yaml，取决于导出包结构
├── train/
│   ├── images/
│   └── labels/
├── valid/             # 有的包叫 val/
│   ├── images/
│   └── labels/
└── test/              # 可选
    ├── images/
    └── labels/
```

### 5. 校验 data.yaml

打开 `data.yaml`，确认：

```yaml
path: ...        # 相对路径或绝对路径
train: train/images
val: valid/images   # 或 val/images
names:
  0: class_a
  1: egg
  # ...
nc: <类别数>
```

在 PowerShell 中快速检查：

```powershell
cd F:\AI_Project\family-recipe-image\data\roboflow
Get-Content .\data.yaml
(Get-ChildItem -Recurse train\images).Count
(Get-ChildItem -Recurse valid\images).Count
```

### 6. 本机试训（可选，验证下载成功）

```powershell
$env:HTTP_PROXY="http://192.168.1.88:7890"
$env:HTTPS_PROXY="http://192.168.1.88:7890"
pip install ultralytics -i https://mirrors.aliyun.com/pypi/simple/

cd F:\AI_Project\family-recipe-image
yolo detect train model=yolov8n.pt data=data/roboflow/data.yaml epochs=3 imgsz=640 batch=8 device=0 project=models/runs name=roboflow-smoke
```

跑 3 个 epoch 仅为冒烟测试；正式训练见 [firdgemate-model.md](firdgemate-model.md)。

---

## 私有数据怎么处理（不要上传 Roboflow）

| 数据类型 | 存放位置 | 标注工具 |
|----------|----------|----------|
| OSS / 家庭误判图 | `data/custom/images/` + `data/custom/labels/` | Label Studio、CVAT、本地 Roboflow 仅离线标注后导出 |
| 合并后训练集 | `data/merged/` | 本机 Python 脚本合并 |

**原则**：私有图只留在 `data/custom/`，与 Roboflow 云端账号无关。

---

## 常见问题

### Q：下载按钮灰色 / 403？

- 开代理后重试，或换 Chrome 无痕窗口登录
- 先 Fork 到 workspace 再 Export
- 换时段重试（Universe 偶发限流）

### Q：必须 Create Project 吗？

- 只下载 Universe 公开集：**不必**自建项目
- 只有 fork + export 或自己标注公开集时才需要 workspace 项目

### Q：下载后类别名和 `labels.go` 不一致？

- 以 `data.yaml` 的 `names` 为准
- 微调完成后更新 `internal/firdgemate/labels.go` 与 `num_classes`

### Q：会不会误把私有数据公开？

- **不要**在 Roboflow 网页 Upload 家庭照片
- **不要**把私有项目设在 Public 计划下
- 只下载别人已公开的 Universe 数据集 → 零泄露风险

---

## 下一步

1. 收集误判样本到 `data/custom/`（如鸡蛋图）
2. 合并 Roboflow + custom → `data/merged/data.yaml`
3. 本机正式训练 → 导出 ONNX → 部署 Pi `~/recipe-image/models/`

详见 [firdgemate-model.md](firdgemate-model.md)。
