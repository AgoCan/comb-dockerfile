# 核心流程

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
