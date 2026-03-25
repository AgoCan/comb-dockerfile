# Docker 镜像组合构建平台设计方案

## 一、系统概述

### 1.1 目标用户
面向普罗大众，包括算法工程师、开发者、数据科学家等

### 1.2 核心价值
解决 Docker 镜像组合构建的痛点：
- **避免重复劳动**：组件化分层，按需组合
- **降低使用门槛**：可视化选择，自动生成 Dockerfile
- **可追溯可复现**：每个镜像来源清晰，支持版本回溯

### 1.3 核心场景
- 用户选择组件组合（如 cuda10 + py36 + tensorflow2.3）
- 系统自动生成组合 Dockerfile
- 用户可自定义修改任意层的 Dockerfile
- 一键构建镜像或导出 Dockerfile

---

## 二、数据库设计

### 2.1 ER 关系图

```
user
  ↓ 1:N
combination_template (用户组合模板)
  
user
  ↓ 1:N
combination_task (构建任务)
  ↓ 1:N
level_task (层级构建任务)
  ↓ N:1
level (层级定义)
  ↓ 1:N
dockerfile (层 Dockerfile 模板)
  ↓ 1:N
dockerfile_version (模板版本历史)

combination_task
  ↓ 1:1
image_record (镜像记录)
```

---

### 2.2 表结构设计

#### user（用户表）
```sql
CREATE TABLE user (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(64) NOT NULL UNIQUE,
    email           VARCHAR(128) UNIQUE,
    password_hash   VARCHAR(256),
    nickname        VARCHAR(64),
    avatar_url      VARCHAR(512),
    status          TINYINT DEFAULT 1 COMMENT '0-禁用 1-正常',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME,
    INDEX idx_username (username),
    INDEX idx_email (email)
);
```

---

#### level（层级定义表）
```sql
CREATE TABLE level (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    name                VARCHAR(64) NOT NULL COMMENT '层级名称，如 cuda、python、opencv',
    order_index         INT NOT NULL COMMENT '构建顺序，数字越小越先构建',
    parent_id           BIGINT DEFAULT 0 COMMENT '父级ID，0表示顶级',
    description         TEXT COMMENT '层级说明',
    icon_url            VARCHAR(512) COMMENT '图标URL',
    is_required         TINYINT DEFAULT 0 COMMENT '是否必选层',
    status              TINYINT DEFAULT 1 COMMENT '0-禁用 1-正常',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME,
    INDEX idx_parent (parent_id),
    INDEX idx_order (order_index)
);
```

**示例数据：**
```
id | name      | order_index | parent_id | description
---|-----------|-------------|-----------|-------------
1  | cuda      | 1           | 0         | NVIDIA CUDA
2  | 10.0      | 1           | 1         | CUDA 10.0
3  | 10.0-cudnn| 1           | 2         | CUDA 10.0 + cuDNN
4  | python    | 2           | 0         | Python 环境
5  | py36      | 2           | 4         | Python 3.6
6  | py37      | 2           | 4         | Python 3.7
```

---

#### dockerfile（Dockerfile 模板表）
```sql
CREATE TABLE dockerfile (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    level_id            BIGINT NOT NULL COMMENT '关联层级',
    name                VARCHAR(128) NOT NULL COMMENT '模板名称',
    content             TEXT NOT NULL COMMENT 'Dockerfile 内容',
    description         TEXT COMMENT '模板说明',
    creator_id          BIGINT COMMENT '创建者ID，NULL表示系统预设',
    is_public           TINYINT DEFAULT 1 COMMENT '是否公开',
    usage_count         INT DEFAULT 0 COMMENT '使用次数',
    status              TINYINT DEFAULT 1,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME,
    INDEX idx_level (level_id),
    INDEX idx_creator (creator_id),
    FOREIGN KEY (level_id) REFERENCES level(id)
);
```

---

#### dockerfile_version（模板版本历史表）
```sql
CREATE TABLE dockerfile_version (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    dockerfile_id       BIGINT NOT NULL,
    version             INT NOT NULL COMMENT '版本号',
    content             TEXT NOT NULL COMMENT 'Dockerfile 内容快照',
    change_note         VARCHAR(512) COMMENT '变更说明',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_dockerfile (dockerfile_id, version),
    FOREIGN KEY (dockerfile_id) REFERENCES dockerfile(id)
);
```

---

#### combination_template（组合模板表）
```sql
CREATE TABLE combination_template (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id             BIGINT NOT NULL,
    name                VARCHAR(128) NOT NULL COMMENT '模板名称',
    description         TEXT,
    config_json         JSON NOT NULL COMMENT '组合配置：[{level_id, dockerfile_id}, ...]',
    is_public           TINYINT DEFAULT 0 COMMENT '是否公开分享',
    use_count           INT DEFAULT 0 COMMENT '使用次数',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME,
    INDEX idx_user (user_id),
    INDEX idx_public (is_public),
    FOREIGN KEY (user_id) REFERENCES user(id)
);
```

**config_json 示例：**
```json
[
    {"level_id": 3, "dockerfile_id": 101, "level_name": "cuda10.0-cudnn"},
    {"level_id": 5, "dockerfile_id": 205, "level_name": "py36"},
    {"level_id": 12, "dockerfile_id": 312, "custom_content": "FROM nvidia/cuda:10.0-cudnn7\nRUN pip install opencv-python"}
]
```

---

#### combination_task（组合构建任务表）
```sql
CREATE TABLE combination_task (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id             BIGINT NOT NULL,
    template_id         BIGINT COMMENT '来源模板ID',
    name                VARCHAR(128) COMMENT '任务名称',
    config_json         JSON NOT NULL COMMENT '组合配置',
    status              VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/running/success/failed/cancelled',
    progress            INT DEFAULT 0 COMMENT '进度百分比',
    error_message       TEXT COMMENT '错误信息',
    build_mode          TINYINT DEFAULT 1 COMMENT '1-仅生成Dockerfile 2-构建镜像',
    require_gpu         TINYINT DEFAULT 0 COMMENT '构建是否需要GPU',
    started_at          DATETIME,
    finished_at         DATETIME,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME,
    INDEX idx_user (user_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at),
    FOREIGN KEY (user_id) REFERENCES user(id)
);
```

---

#### level_task（层级构建任务表）
```sql
CREATE TABLE level_task (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    combination_task_id BIGINT NOT NULL,
    level_id            BIGINT NOT NULL,
    parent_level_task_id BIGINT COMMENT '父级构建任务ID',
    dockerfile_id       BIGINT COMMENT '使用的模板ID',
    custom_dockerfile   TEXT COMMENT '自定义Dockerfile内容',
    order_index         INT NOT NULL COMMENT '构建顺序',
    status              VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/running/success/failed/skipped',
    image_name          VARCHAR(256) COMMENT '生成的镜像名',
    image_id            VARCHAR(128) COMMENT '镜像ID',
    build_log_url       VARCHAR(512) COMMENT '构建日志URL',
    build_duration      INT COMMENT '构建耗时(秒)',
    error_message       TEXT,
    started_at          DATETIME,
    finished_at         DATETIME,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_combination (combination_task_id),
    INDEX idx_status (status),
    FOREIGN KEY (combination_task_id) REFERENCES combination_task(id),
    FOREIGN KEY (level_id) REFERENCES level(id)
);
```

---

#### image_record（镜像记录表）
```sql
CREATE TABLE image_record (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    combination_task_id BIGINT NOT NULL UNIQUE,
    user_id             BIGINT NOT NULL,
    full_image_name     VARCHAR(512) COMMENT '完整镜像名 registry/group/name:tag',
    short_image_name    VARCHAR(256) COMMENT '短镜像名 name:tag',
    image_id            VARCHAR(128) COMMENT '镜像ID',
    image_info          VARCHAR(1024) COMMENT '镜像信息，注入ENV',
    full_dockerfile     TEXT COMMENT '完整Dockerfile',
    size_bytes          BIGINT COMMENT '镜像大小',
    push_status         TINYINT DEFAULT 0 COMMENT '0-未推送 1-推送中 2-已推送 3-推送失败',
    pushed_at           DATETIME,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_task (combination_task_id),
    FOREIGN KEY (combination_task_id) REFERENCES combination_task(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);
```

---

#### resource（资源镜像表）
```sql
CREATE TABLE resource (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    image_name          VARCHAR(256) NOT NULL COMMENT '镜像名',
    dockerfile_url      VARCHAR(512) COMMENT 'Dockerfile来源URL',
    description         TEXT,
    skip_from_replace   TINYINT DEFAULT 1 COMMENT '构建时是否跳过FROM替换',
    status              TINYINT DEFAULT 1,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME,
    INDEX idx_image (image_name)
);
```

**说明：** 资源镜像（如 nvidia/cuda:10.0-base）作为基础镜像，构建时不替换其 FROM

---

#### level_compatibility（层级兼容性表）
```sql
CREATE TABLE level_compatibility (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    level_id_a          BIGINT NOT NULL COMMENT '层级A',
    level_id_b          BIGINT NOT NULL COMMENT '层级B',
    is_compatible       TINYINT DEFAULT 1 COMMENT '是否兼容 0-不兼容 1-兼容',
    note                VARCHAR(256) COMMENT '兼容性说明',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_level_a (level_id_a),
    INDEX idx_level_b (level_id_b),
    UNIQUE KEY uk_levels (level_id_a, level_id_b)
);
```

**示例：**
```
level_id_a | level_id_b | is_compatible | note
-----------|------------|---------------|------------------
3          | 5          | 1             | cuda10.0 支持 py36
3          | 7          | 0             | cuda10.0 不支持 py39
```

---

#### config（系统配置表）
```sql
CREATE TABLE config (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key          VARCHAR(64) NOT NULL UNIQUE,
    config_value        TEXT,
    description         VARCHAR(256),
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_key (config_key)
);
```

---

## 三、核心流程

### 3.1 创建组合流程

```
用户登录
    ↓
选择层级（前端按 order_index 排序展示）
    ↓
每层选择模板 OR 自定义 Dockerfile
    ↓
系统校验兼容性（level_compatibility）
    ↓
预览完整 Dockerfile
    ↓
选择构建模式（仅生成 / 构建镜像）
    ↓
创建 combination_task
    ↓
    ├─ 模式1：直接返回 Dockerfile，任务结束
    └─ 模式2：创建 level_task，加入构建队列
```

### 3.2 构建镜像流程

```
构建队列消费任务
    ↓
按 order_index 顺序构建 level_task
    ↓
每个 level_task：
    ├─ 获取 Dockerfile（模板 or 自定义）
    ├─ 替换 FROM（除 resource 表中的镜像）
    ├─ 注入 ENV IMAGE_INFO / IMAGE_NAME
    ├─ 执行 docker build
    └─ 记录构建结果
    ↓
全部成功 → 生成 image_record
    ↓
推送到镜像仓库（可选）
```

### 3.3 镜像缓存复用机制

```
构建前检查：
    ├─ 查询相同 config_json 的历史成功任务
    ├─ 检查对应镜像是否仍存在
    └─ 存在 → 提示用户"已有相同镜像，是否复用？"
```

---

## 四、API 设计

### 4.1 层级管理

```
GET    /api/levels              # 获取层级树
GET    /api/levels/{id}/dockerfiles  # 获取某层的所有模板
POST   /api/dockerfiles         # 创建自定义模板
PUT    /api/dockerfiles/{id}    # 更新模板（自动创建版本）
GET    /api/dockerfiles/{id}/versions  # 获取模板版本历史
```

### 4.2 组合构建

```
POST   /api/combinations/check  # 校验组合兼容性
POST   /api/combinations/preview # 预览完整Dockerfile
POST   /api/combinations        # 创建构建任务
GET    /api/combinations/{id}   # 获取任务详情
GET    /api/combinations/{id}/log # 获取构建日志
POST   /api/combinations/{id}/cancel # 取消任务
```

### 4.3 镜像管理

```
GET    /api/images              # 获取用户镜像列表
GET    /api/images/{id}         # 获取镜像详情
GET    /api/images/{id}/dockerfile # 获取完整Dockerfile
DELETE /api/images/{id}         # 删除镜像记录
```

### 4.4 模板管理

```
POST   /api/templates           # 保存组合模板
GET    /api/templates           # 获取我的模板
GET    /api/templates/public    # 获取公开模板
POST   /api/templates/{id}/use  # 使用模板创建任务
```

---

## 五、镜像命名规范

### 5.1 镜像信息注入

每个构建的镜像都会注入 ENV：

```dockerfile
ENV IMAGE_INFO="cuda10.0-cudnn7-py36-tf2.3-openvino-lab-coding"
ENV IMAGE_NAME="myapp/cuda10-py36-tf23:v1.0.0"
ENV BUILD_TIME="2024-01-15T10:30:00Z"
ENV BUILD_TASK_ID="12345"
```

### 5.2 镜像命名方案

**推荐方案：短名称 + ENV 详情**

```
镜像名格式：
    {registry}/{group}/{project}:{version}

示例：
    registry.example.com/team/project:v1.0.0

详细内容通过 docker inspect 查看 ENV.IMAGE_INFO
```

**优点：**
- 镜像名简短易记
- 详细信息通过 ENV 完整保留
- 支持任意复杂组合

---

## 六、多模式场景设计

### 6.1 场景分类

系统支持多种镜像构建场景，每种场景有独立的层级配置和推荐组合：

| 模式 | 说明 | 典型层级 |
|------|------|----------|
| 🎯 训练镜像 | 深度学习/机器学习训练环境 | cuda → python → framework → tools |
| 🚀 推理镜像 | 模型部署和推理服务 | cuda/cpu → python → inference-engine → web-framework |
| 📊 数据处理 | ETL、数据清洗、特征工程 | python → data-tools → compute-engine → storage |
| 💻 开发环境 | 云 IDE、在线开发环境 | base-os → runtime → editor → dev-tools |
| 🔧 CI/CD | 构建流水线 Runner | base-os → language → build-tools → deploy-tools |
| 🌐 微服务 | 后端服务镜像 | base-os → language → framework → middleware |
| 🧪 测试镜像 | 自动化测试环境 | base-os → language → test-framework → browser |
| 🔒 安全扫描 | 代码/镜像安全检查 | base-os → scan-tools → report-tools |
| 📦 大数据 | 数据湖、流批处理 | base-os → runtime → bigdata-engine → connectors |

### 6.2 数据结构调整

#### level 表增加场景分类

```sql
ALTER TABLE level ADD COLUMN category VARCHAR(32) NOT NULL DEFAULT 'training' 
COMMENT '场景分类: training/inference/data/dev/cicd/service/testing/security/bigdata';

ALTER TABLE level ADD INDEX idx_category (category);
```

#### 场景配置表

```sql
CREATE TABLE scenario_config (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    category            VARCHAR(32) NOT NULL UNIQUE COMMENT '场景标识',
    display_name        VARCHAR(64) NOT NULL COMMENT '显示名称',
    description         TEXT COMMENT '场景描述',
    icon                VARCHAR(64) COMMENT '图标标识',
    sort_order          INT DEFAULT 0 COMMENT '排序',
    default_levels      JSON COMMENT '默认选中的层级ID列表',
    required_categories JSON COMMENT '必选层级分类',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_category (category)
);
```

**示例数据：**
```sql
INSERT INTO scenario_config (category, display_name, icon, sort_order, default_levels, required_categories) VALUES
('training', '训练镜像', '🎯', 1, '["cuda", "python", "framework"]', '["cuda", "python"]'),
('inference', '推理镜像', '🚀', 2, '["cuda", "python", "inference", "web"]', '["python"]'),
('data', '数据处理', '📊', 3, '["python", "data-tools"]', '["python"]'),
('dev', '开发环境', '💻', 4, '["base", "runtime", "editor"]', '["base", "runtime"]'),
('cicd', 'CI/CD', '🔧', 5, '["base", "language", "build"]', '["base"]');
```

---

## 七、前端页面设计

### 7.1 页面结构

```
1. 首页（场景选择）
   ┌─────────────────────────────────────────────────────────┐
   │  选择镜像类型                                             │
   ├─────────────────────────────────────────────────────────┤
   │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
   │  │ 🎯      │  │ 🚀      │  │ 📊      │  │ 💻      │     │
   │  │ 训练镜像 │  │ 推理镜像 │  │ 数据处理 │  │ 开发环境 │     │
   │  │ [开始]  │  │ [开始]  │  │ [开始]  │  │ [开始]  │     │
   │  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │
   │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
   │  │ 🔧      │  │ 🌐      │  │ 🧪      │  │ 🔒      │     │
   │  │ CI/CD   │  │ 微服务  │  │ 测试镜像 │  │ 安全扫描 │     │
   │  │ [开始]  │  │ [开始]  │  │ [开始]  │  │ [开始]  │     │
   │  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │
   ├─────────────────────────────────────────────────────────┤
   │  最近构建记录 | 热门组合模板                               │
   └─────────────────────────────────────────────────────────┘

2. 组合创建页（根据选择的场景动态展示层级）
   - 场景标题：🎯 训练镜像
   - 层级选择器（仅展示该场景相关层级）
   - 每层模板选择/自定义
   - 兼容性实时校验
   - Dockerfile 预览
   - 构建选项

3. 构建监控页
   - 任务列表（按场景分类筛选）
   - 实时日志
   - 进度展示

4. 镜像管理页
   - 镜像列表（按场景分类）
   - 搜索过滤
   - Dockerfile 查看
   - 推送状态

5. 模板中心
   - 按场景分类展示
   - 我的模板
   - 公开模板
   - 收藏/使用

6. Dockerfile 编辑器
   - 语法高亮
   - 版本对比
   - 保存为新模板
```

### 7.2 用户流程

```
首页
  │
  ├─ 选择场景 ─────────────────────┐
  │   (训练镜像)                    │
  ↓                                │
场景专属组合页                      │
  │                                │
  ├─ 加载该场景的层级配置            │
  ├─ 必选层高亮提示                  │
  ├─ 按场景顺序引导选择              │
  │                                │
  ├────────────────────────────────┤
  │                                │
  ↓                                │
组合详情页                          │
  │                                │
  ├─ 预览 Dockerfile               │
  ├─ 选择构建模式                   │
  └─ 提交构建                       │
                                   │
其他场景同理 ←─────────────────────┘
```

### 7.3 场景专属层级选择

**训练镜像场景：**
```
第1步：选择 GPU 环境
    ┌─────────────────────────────────────┐
    │ 🎯 训练镜像                          │
    ├─────────────────────────────────────┤
    │ 1. CUDA 环境（必选）                 │
    │    ○ cuda 10.0                      │
    │    ○ cuda 10.2                      │
    │    ○ cuda 11.0                      │
    │    ○ cuda 11.7                      │
    │    ○ cpu-only（无GPU）              │
    └─────────────────────────────────────┘

第2步：选择 Python 版本
    ┌─────────────────────────────────────┐
    │ 2. Python 环境（必选）               │
    │    ○ Python 3.6                     │
    │    ○ Python 3.7                     │
    │    ○ Python 3.8                     │
    │    ○ Python 3.9                     │
    │    ○ Python 3.10                    │
    └─────────────────────────────────────┘

第3步：选择深度学习框架
    ┌─────────────────────────────────────┐
    │ 3. 深度学习框架                      │
    │    ○ TensorFlow 2.x                 │
    │    ○ PyTorch 1.x                    │
    │    ○ PaddlePaddle                   │
    │    ○ 无框架（自定义）                │
    └─────────────────────────────────────┘

... 继续后续层级
```

**推理镜像场景：**
```
第1步：选择运行环境
    ┌─────────────────────────────────────┐
    │ 🚀 推理镜像                          │
    ├─────────────────────────────────────┤
    │ 1. 运行环境（必选）                  │
    │    ○ GPU（CUDA）                    │
    │    ○ CPU Only                       │
    └─────────────────────────────────────┘

第2步：选择 Python 版本
    ...

第3步：选择推理引擎
    ┌─────────────────────────────────────┐
    │ 3. 推理引擎                          │
    │    ○ ONNX Runtime                   │
    │    ○ TensorRT                       │
    │    ○ OpenVINO                       │
    │    ○ TorchServe                     │
    │    ○ TFServing                      │
    └─────────────────────────────────────┘

第4步：选择服务框架
    ┌─────────────────────────────────────┐
    │ 4. 服务框架                          │
    │    ○ FastAPI                        │
    │    ○ Flask                          │
    │    ○ gRPC                           │
    │    ○ Triton                         │
    └─────────────────────────────────────┘
```

**开发环境场景：**
```
第1步：选择基础系统
    ┌─────────────────────────────────────┐
    │ 💻 开发环境                          │
    ├─────────────────────────────────────┤
    │ 1. 基础系统（必选）                  │
    │    ○ Ubuntu 22.04                   │
    │    ○ Ubuntu 20.04                   │
    │    ○ Debian 11                      │
    │    ○ Alpine 3.18                    │
    └─────────────────────────────────────┘

第2步：选择运行时
    ┌─────────────────────────────────────┐
    │ 2. 运行时环境（必选）                │
    │    ○ Python 3.11                    │
    │    ○ Node.js 18                     │
    │    ○ Go 1.21                        │
    │    ○ Java 17                        │
    │    ○ 多语言（自定义）                │
    └─────────────────────────────────────┘

第3步：选择编辑器
    ┌─────────────────────────────────────┐
    │ 3. 开发工具                          │
    │    ○ VS Code Server                 │
    │    ○ JupyterLab                     │
    │    ○ Vim + Plugins                  │
    │    ○ 无编辑器                        │
    └─────────────────────────────────────┘
```

### 7.4 API 调整

```
GET    /api/scenarios                    # 获取所有场景列表
GET    /api/scenarios/{category}/levels  # 获取某场景的层级配置
POST   /api/combinations                 # 创建构建任务（增加 category 参数）
GET    /api/combinations?category=training # 按场景筛选任务
GET    /api/images?category=inference    # 按场景筛选镜像
```

---

## 七、技术选型建议

### 7.1 后端

| 组件 | 建议 | 理由 |
|------|------|------|
| 语言 | Go / Python | Go 适合高并发构建，Python 生态丰富 |
| 框架 | Gin / FastAPI | 轻量高效 |
| ORM | GORM / SQLAlchemy | 成熟稳定 |
| 队列 | Redis + Celery / RabbitMQ | 构建任务异步处理 |
| 构建引擎 | Docker SDK / Kaniko | 无 daemon 构建 |

### 7.2 前端

| 组件 | 建议 | 理由 |
|------|------|------|
| 框架 | Vue 3 / React 18 | 现代化 |
| 组件库 | Element Plus / Ant Design | 企业级 |
| 编辑器 | Monaco Editor | VS Code 同款 |
| 状态管理 | Pinia / Zustand | 轻量 |

### 7.3 基础设施

| 组件 | 建议 |
|------|------|
| 镜像仓库 | Harbor / Docker Registry |
| 日志存储 | MinIO / S3 |
| 缓存 | Redis |
| 数据库 | MySQL 8.0 / PostgreSQL |

---

## 八、扩展设计

### 8.1 权限控制

```
用户角色：
    - 普通用户：创建组合、管理自己的镜像和模板
    - 管理员：管理层级、模板、系统配置
    - 访客：查看公开模板和镜像（无需登录）
```

### 8.2 配额管理

```
限制：
    - 单用户同时运行任务数
    - 单用户镜像数量上限
    - 构建频率限制
    - 镜像存储空间上限
```

### 8.3 统计分析

```
指标：
    - 各层级使用频率
    - 热门组合 TOP10
    - 构建成功率
    - 平均构建时间
    - 用户活跃度
```

---

## 九、总结

本设计相比原方案的改进：

| 改进点 | 原方案 | 新方案 |
|--------|--------|--------|
| 用户体系 | 无 | 完整的用户、权限、配额 |
| Dockerfile 管理 | 重复存储 | 模板化 + 版本管理 |
| 构建状态 | 简单 status | 详细状态机 + 进度追踪 |
| 镜像缓存 | 无 | 复用检测机制 |
| 兼容性校验 | 无 | 层级兼容性表 |
| 模板机制 | 无 | 用户可保存常用组合 |

核心思想：**在保持灵活性的同时，通过模板化、版本化、缓存复用等机制提升效率和可维护性。**
