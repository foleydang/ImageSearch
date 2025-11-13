# ImageSearch

基于Go语言的图片搜索系统，支持图片上传、管理和相似图片搜索功能。

## 功能特点

- **图片上传**：支持JPEG和PNG格式图片上传
- **图片管理**：查看、列出和删除图片
- **相似图片搜索**：根据图片内容搜索相似图片
- **RESTful API**：提供标准的RESTful API接口
- **数据持久化**：使用SQLite数据库存储图片信息和嵌入向量
- **图片处理**：自动调整图片大小和格式

## 技术栈

- **Go 1.20**：主要开发语言
- **Gin**：Web框架
- **GORM**：ORM框架
- **SQLite**：数据库
- **UUID**：唯一标识生成
- **image**：图片处理
- **resize**：图片大小调整

## 项目结构

```
├── assets/           # 静态资源目录
│   └── images/       # 上传的图片存储目录
├── bin/              # 编译后的可执行文件
├── cmd/              # 命令行入口
│   └── server/       # 服务器启动入口
├── internal/         # 内部包
│   ├── api/          # API处理器
│   ├── config/       # 配置管理
│   ├── model/        # 数据模型
│   ├── repository/   # 数据仓库
│   ├── service/      # 业务逻辑层
│   └── utils/        # 工具函数
├── pkg/              # 可导出的包
├── script/           # 脚本文件
├── go.mod            # Go模块定义
├── go.sum            # 依赖版本锁定
└── imagesearch.db    # SQLite数据库文件
```

## 快速开始

### 环境要求

- Go 1.20或更高版本
- 支持的操作系统：Linux、macOS、Windows

### 安装和运行

1. **克隆项目**

```bash
git clone https://github.com/bytedance/ImageSearch.git
cd ImageSearch
```

2. **启动服务**

```bash
./script/start.sh
```

启动脚本会自动下载依赖、编译项目并启动服务。服务默认监听在 `0.0.0.0:8080` 地址。

## API文档

### 1. 获取API信息

```
GET /
```

返回API的基本信息，包括版本、状态和可用的端点列表。

### 2. 健康检查

```
GET /health
```

检查服务是否正常运行。

**响应示例：**
```json
{
  "message": "ImageSearch API is running",
  "status": "ok"
}
```

### 3. 上传图片

```
POST /api/images
```

**请求参数**（multipart/form-data）：
- `file`：图片文件（必需）
- `name`：图片名称（可选）
- `description`：图片描述（可选）

**响应示例：**
```json
{
  "id": "075c9b4c-fb6d-43ab-9e69-24f86d4b87be",
  "file_name": "test_image.png",
  "file_path": "assets/images/fa8f9f35-5f65-4f7c-97d0-bb411aae0222.png",
  "extension": "png",
  "width": 800,
  "height": 800,
  "size": 4293,
  "created_at": "2025-11-13T17:19:14.811131+08:00",
  "updated_at": "2025-11-13T17:19:14.811131+08:00"
}
```

### 4. 获取图片列表

```
GET /api/images?page=1&page_size=10
```

**查询参数：**
- `page`：页码（默认1）
- `page_size`：每页大小（默认10，最大100）

### 5. 获取单个图片

```
GET /api/images/:id
```

### 6. 删除图片

```
DELETE /api/images/:id
```

### 7. 相似图片搜索

```
POST /api/images/search
```

**请求参数**（multipart/form-data）：
- `file`：用于搜索的图片文件（必需）

**响应示例：**
```json
{
  "results": [
    {
      "image": {
        "id": "075c9b4c-fb6d-43ab-9e69-24f86d4b87be",
        "file_name": "test_image.png",
        "file_path": "assets/images/fa8f9f35-5f65-4f7c-97d0-bb411aae0222.png",
        "extension": "png",
        "width": 800,
        "height": 800,
        "size": 4293,
        "created_at": "2025-11-13T17:19:14.811131+08:00",
        "updated_at": "2025-11-13T17:19:14.811131+08:00"
      },
      "distance": 0,
      "image_url": "/images/fa8f9f35-5f65-4f7c-97d0-bb411aae0222.png"
    }
  ],
  "total": 1
}
```

### 8. 访问图片文件

```
GET /images/:filename
```

## 使用示例

### 上传图片

```bash
curl -X POST -F "file=@path/to/your/image.jpg" -F "name=测试图片" http://localhost:8080/api/images
```

### 搜索相似图片

```bash
curl -X POST -F "file=@path/to/your/search_image.jpg" http://localhost:8080/api/images/search
```

### 获取图片列表

```bash
curl http://localhost:8080/api/images?page=1&page_size=10
```

## 注意事项

1. 目前使用的是简化的图像嵌入向量生成方法（基于平均颜色），在生产环境中建议集成更高级的图像特征提取模型。
2. SQLite数据库适合小规模应用，大规模应用建议使用PostgreSQL等更强大的数据库。
3. 图片存储在本地文件系统中，生产环境可以考虑使用云存储服务。

## 未来优化方向

1. 集成深度学习模型进行图像特征提取
2. 添加图片分类和标签功能
3. 支持更多图片格式
4. 实现图片压缩和优化
5. 添加用户认证和权限管理
6. 支持图片批量上传
7. 实现图片元数据提取和搜索

## 许可证

本项目采用 MIT 许可证 - 查看 LICENSE 文件了解详情
