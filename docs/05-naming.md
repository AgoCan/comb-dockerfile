# 镜像命名规范

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
