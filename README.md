# 博客系统后端服务器

基于Golang + Gin + PostgreSQL + Redis构建的简易博客系统后端API，支持完整的内容管理和数据分析功能。

## 功能特性

- 🔐 **管理员认证系统**：用户注册/登录（需要密令验证）、JWT身份验证
- 📝 **文章管理**：支持Markdown格式的文章创建、编辑、删除（软删除）、内容分表查询
- 👤 **公共信息管理**：个人资料、技能、联系方式等信息管理，供前端博客首页调用
- 🖼️ **图片存储功能**：集成Cloudflare R2对象存储，支持图片上传和删除、自动生成唯一文件名
- 📊 **完整日志系统**：记录所有API调用，包含函数名、级别、错误信息、响应时间，敏感数据过滤
- 📈 **完整数据分析系统**：
  - 用户行为追踪（访问时间、路径、IP、User-Agent、事件类型）
  - 实时在线用户统计
  - 页面访问热力图
  - 文章点击统计
  - Redis缓存 + PostgreSQL存储
  - 自动定时数据转存
- 🔄 **数据库迁移**：版本化的数据库结构管理，支持自动迁移和回滚
- ⚡ **Redis缓存**：支持Redis URL和分离参数两种连接方式，实时数据缓存和定时数据转存

## 技术栈

- **后端框架**: Gin (Golang)
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis
- **身份验证**: JWT + bcrypt
- **文件存储**: Cloudflare R2 (S3兼容)
- **配置管理**: godotenv
- **部署**: Docker + Dokploy

## 快速开始

### 环境要求

- Go 1.24+
- PostgreSQL 12+
- Redis 6+

### 本地开发

1. **克隆项目**
```bash
git clone <your-repo-url>
cd blog-server
```

2. **安装依赖**
```bash
go mod tidy
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

### 数据分析系统
- `POST /api/analytics/track` - 数据收集接口（无需认证）
- `GET /api/analytics/realtime` - 实时统计数据（无需认证）
- `GET /api/analytics/daily` - 每日统计数据 🔒
- `GET /api/analytics/range` - 日期范围统计 🔒
- `GET /api/analytics/top-pages` - 热门页面统计 🔒
- `GET /api/analytics/events` - 详细访问记录查询 🔒
- `GET /api/analytics/ip-stats` - IP访问统计 🔒
- `GET /api/analytics/user-agent-stats` - User-Agent统计 🔒
- `GET /api/analytics/referer-stats` - 来源统计 🔒
- `GET /api/analytics/session-stats` - 会话统计 🔒
- `GET /api/analytics/event-type-stats` - 事件类型统计 🔒
- `GET /api/analytics/hourly-stats` - 按小时统计 🔒
- `GET /api/analytics/path-analysis` - 路径详细分析 🔒
- `GET /api/analytics/advanced-stats` - 高级统计数据 🔒

🔒 = 需要JWT认证

## 环境变量配置

创建 `.env` 文件：

```env
# 数据库配置
DATABASE_URL=postgresql://username:password@localhost:5432/blog_db

# JWT配置
JWT_SECRET=your-jwt-secret-key
REGISTER_PASSWORD=your-register-secret-password

# 服务器配置
PORT=8080
GIN_MODE=debug

# Cloudflare R2配置
R2_ACCESS_KEY_ID=your-r2-access-key-id
R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
R2_BUCKET_NAME=your-bucket-name
R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
R2_PUBLIC_URL=https://your-custom-domain.com

# Redis配置（两种方式二选一）
# 方式1：完整URL（适用于云服务）
REDIS_URL=redis://default:password@host:6379/0

# 方式2：分离参数（适用于本地开发）
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Docker部署

### 使用Dokploy部署

1. 在GitHub上创建仓库并推送代码
2. 在Dokploy中连接GitHub仓库
3. 配置环境变量
4. 部署应用

项目包含完整的Dockerfile，支持：
- 多阶段构建，减少镜像大小
- 包含迁移工具
- 自动设置时区
- 健康检查支持
- 自动数据库迁移

## 项目结构

```
├── cmd/
│   └── migrate/          # 数据库迁移工具
├── config/               # 配置管理
├── controllers/          # API控制器
│   ├── analytics.go      # 数据分析控制器
│   ├── article.go        # 文章管理控制器
│   ├── auth.go          # 认证控制器
│   ├── profile.go       # 公共信息控制器
│   ├── upload.go        # 图片上传控制器
│   └── user.go          # 用户信息控制器
├── middleware/           # 中间件
├── models/               # 数据模型
│   ├── analytics.go      # 数据分析模型
│   ├── article.go        # 文章模型
│   ├── migration.go      # 数据库迁移
│   ├── profile.go        # 公共信息模型
│   └── user.go          # 用户模型
├── routes/               # 路由定义
├── utils/                # 工具函数
│   ├── redis.go         # Redis工具
│   ├── scheduler.go     # 定时任务
│   └── storage.go       # 存储工具
├── scripts/              # 脚本文件
├── main.go              # 程序入口
├── Dockerfile           # Docker配置
└── README.md            # 项目说明
```

## 数据库模型

- `User`: 管理员用户表
- `Article`: 文章表（支持Markdown）
- `Profile`: 公共信息表
- `APILog`: API日志记录表
- `TrackingEvent`: 用户行为追踪事件表
- `DailyStats`: 每日统计数据表
- `PageHeatmap`: 页面热力图数据表

## 特殊功能详解

### 数据分析系统
- **数据收集**：通过`/api/analytics/track`接口收集用户行为数据
- **实时缓存**：使用Redis按日期存储实时数据
- **定时转存**：每日凌晨0:05自动将Redis数据转存到PostgreSQL
- **统计功能**：
  - 在线用户数量（基于IP，30分钟TTL）
  - 页面访问统计（PV/UV）
  - 文章点击热度
  - 页面热力图数据
- **安全处理**：User-Agent通过SHA256哈希存储
- **API接口**：
  - 实时统计：无需认证，供前端展示
  - 历史数据：需要认证，管理员查看
  - 详细分析：提供多维度数据查询和统计

### 日志系统
- 自动记录所有API调用到数据库
- 记录内容包括：请求方法、路径、状态码、响应时间、用户信息、函数名、错误信息等
- 敏感数据（如密码、密令）自动过滤
- 支持不同日志级别：info、warn、error

### 图片上传功能
- 支持格式：jpg、jpeg、png、gif、webp
- 文件大小限制：5MB
- 自动生成唯一文件名（时间戳+随机字符串）
- 存储到Cloudflare R2，返回公共访问URL
- 文件类型和大小验证

### Redis连接支持
- **方式1**：完整URL格式（`REDIS_URL`），适用于云服务如Dokploy
- **方式2**：分离参数（`REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB`），适用于本地开发
- 自动向后兼容，优先使用REDIS_URL

### 数据库迁移
- 版本化管理，支持增量迁移
- 事务保护，确保数据一致性
- 自动执行，部署时自动运行迁移
- 支持回滚到指定版本

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

# 查看依赖
go mod tidy
```

## 生产部署注意事项

### 环境变量配置
在Dokploy中需要配置以下环境变量：
```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable
JWT_SECRET=your-production-jwt-secret
REGISTER_PASSWORD=your-production-register-password
GIN_MODE=release
PORT=8080

# Cloudflare R2配置
R2_ACCESS_KEY_ID=your-key-id
R2_SECRET_ACCESS_KEY=your-secret-key
R2_BUCKET_NAME=your-bucket-name
R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
R2_PUBLIC_URL=https://your-custom-domain.com

# Redis配置（推荐使用Redis URL）
REDIS_URL=redis://default:your-redis-password@your-redis-host:6379/0
```

### 数据库迁移流程

#### 本地开发
```bash
# 查看迁移状态
./scripts/migrate.sh status

# 执行所有未应用的迁移
./scripts/migrate.sh up

# 回滚特定版本
./scripts/migrate.sh down 002
```

#### 生产部署（Dokploy）
1. **首次部署**：在Dokploy中配置环境变量，应用会自动执行数据库迁移
2. **更新数据库结构**：在`models/migration.go`中添加新的迁移项，推送代码到GitHub，Dokploy重新部署时会自动执行新迁移
3. **紧急回滚**：
   ```bash
   # 在服务器容器中执行
   docker exec -it <container_name> ./migrate -action=down -version=<version>
   ```

## 许可证

MIT License