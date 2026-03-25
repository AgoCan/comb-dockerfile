# 多模式场景设计

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

参见 [数据库设计 - 场景配置表](./02-database.md#scenario_config场景配置表)
