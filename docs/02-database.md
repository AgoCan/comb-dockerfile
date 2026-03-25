# 数据库设计

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

#### scenario_config（场景配置表）

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

#### level 表增加场景分类

```sql
ALTER TABLE level ADD COLUMN category VARCHAR(32) NOT NULL DEFAULT 'training' 
COMMENT '场景分类: training/inference/data/dev/cicd/service/testing/security/bigdata';

ALTER TABLE level ADD INDEX idx_category (category);
```
