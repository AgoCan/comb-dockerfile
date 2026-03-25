# 技术选型建议

## 八、技术选型建议

### 8.1 后端

| 组件 | 建议 | 理由 |
|------|------|------|
| 语言 | Go / Python | Go 适合高并发构建，Python 生态丰富 |
| 框架 | Gin / FastAPI | 轻量高效 |
| ORM | GORM / SQLAlchemy | 成熟稳定 |
| 队列 | Redis + Celery / RabbitMQ | 构建任务异步处理 |
| 构建引擎 | Docker SDK / Kaniko | 无 daemon 构建 |

### 8.2 前端

| 组件 | 建议 | 理由 |
|------|------|------|
| 框架 | Vue 3 / React 18 | 现代化 |
| 组件库 | Element Plus / Ant Design | 企业级 |
| 编辑器 | Monaco Editor | VS Code 同款 |
| 状态管理 | Pinia / Zustand | 轻量 |

### 8.3 基础设施

| 组件 | 建议 |
|------|------|
| 镜像仓库 | Harbor / Docker Registry |
| 日志存储 | MinIO / S3 |
| 缓存 | Redis |
| 数据库 | MySQL 8.0 / PostgreSQL |
