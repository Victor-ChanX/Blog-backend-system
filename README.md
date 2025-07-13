# Blog Server

基于Golang + Gin + PostgreSQL构建的简易blog系统后端API。

## 功能特性

- 🔐 **管理员认证系统**：注册/登录（需要密令验证）
- 📝 **文章管理**：支持Markdown格式的文章创建、编辑、删除（软删除）
- 👤 **公共信息管理**：个人资料、技能、联系方式等信息管理
- 🖼️ **图片存储**：集成Cloudflare R2对象存储
- 📊 **完整日志系统**：记录所有API调用，包含函数名、级别、错误信息等
- 🔄 **数据库迁移**：版本化的数据库结构管理

## 技术栈

- **后端框架**: Gin (Golang)
- **数据库**: PostgreSQL + GORM
- **身份验证**: JWT + bcrypt
- **文件存储**: Cloudflare R2 (S3兼容)
- **部署**: Docker + Dokploy

## 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL 12+

### 本地开发

1. **克隆项目**
```bash
git clone <your-repo-url>
cd blog-server
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑.env文件，配置数据库连接和其他参数
```

4. **启动服务**
```bash
go run .
```

服务将在 `http://localhost:8080` 启动

### 数据库迁移

```bash
# 查看迁移状态
./scripts/migrate.sh status

# 执行迁移
./scripts/migrate.sh up

# 回滚迁移
./scripts/migrate.sh down <version>
```

## API文档

### 认证接口
- `POST /api/auth/register` - 用户注册（需要密令）
- `POST /api/auth/login` - 用户登录

### 文章管理
- `GET /api/articles` - 获取文章列表
- `GET /api/articles/:id` - 获取单篇文章
- `POST /api/articles` - 创建文章 🔒
- `PUT /api/articles/:id` - 更新文章 🔒
- `DELETE /api/articles/:id` - 删除文章 🔒

### 公共信息
- `GET /api/profile` - 获取公共信息
- `PUT /api/profile` - 更新公共信息 🔒

### 图片上传
- `POST /api/upload/image` - 上传图片 🔒
- `DELETE /api/upload/image` - 删除图片 🔒

### 用户信息
- `GET /api/user/profile` - 获取当前用户信息 🔒

🔒 = 需要JWT认证

## 环境变量配置

创建 `.env` 文件：

```env
# 数据库配置
DATABASE_URL=postgres://username:password@localhost:5432/blog_db?sslmode=disable

# JWT配置
JWT_SECRET=your-jwt-secret-key
REGISTER_PASSWORD=your-register-secret-password

# 服务器配置
PORT=8080
GIN_MODE=debug

# Cloudflare R2配置（可选）
R2_ACCESS_KEY_ID=your-r2-access-key-id
R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
R2_BUCKET_NAME=your-bucket-name
R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
R2_PUBLIC_URL=https://your-custom-domain.com
```

## Docker部署

### 使用Dokploy部署

1. 在GitHub上创建仓库并推送代码
2. 在Dokploy中连接GitHub仓库
3. 配置环境变量
4. 部署应用

项目包含完整的Dockerfile，支持：
- 多阶段构建
- 自动数据库迁移
- 健康检查

## 项目结构

```
├── cmd/
│   └── migrate/          # 数据库迁移工具
├── config/               # 配置管理
├── controllers/          # API控制器
├── middleware/           # 中间件
├── models/               # 数据模型
├── routes/               # 路由定义
├── utils/                # 工具函数
├── scripts/              # 脚本文件
├── main.go              # 程序入口
├── Dockerfile           # Docker配置
└── README.md            # 项目说明
```

## 特殊功能

### 日志系统
- 自动记录所有API调用
- 包含请求详情、响应时间、用户信息
- 敏感数据自动过滤
- 支持错误级别分类

### 图片上传
- 支持格式：jpg、jpeg、png、gif、webp
- 文件大小限制：5MB
- 自动生成唯一文件名
- 存储到Cloudflare R2

### 数据库迁移
- 版本化管理
- 事务保护
- 自动执行
- 支持回滚

## 开发命令

```bash
# 启动开发服务器
go run .

# 编译生产版本
go build -o blog-server .

# 运行测试
go test ./...

# 数据库迁移
./scripts/migrate.sh up
```

## 许可证

MIT License