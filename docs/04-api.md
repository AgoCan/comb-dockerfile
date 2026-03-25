# API 设计

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

### 4.5 场景相关

```
GET    /api/scenarios                    # 获取所有场景列表
GET    /api/scenarios/{category}/levels  # 获取某场景的层级配置
POST   /api/combinations                 # 创建构建任务（增加 category 参数）
GET    /api/combinations?category=training # 按场景筛选任务
GET    /api/images?category=inference    # 按场景筛选镜像
```
