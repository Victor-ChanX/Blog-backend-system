# Blog API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **认证方式**: Bearer Token (JWT)

## 认证说明

需要认证的接口在请求头中包含：
```
Authorization: Bearer <jwt_token>
```

---

## 1. 认证接口

### 1.1 用户注册

**POST** `/api/auth/register`

**请求体**:
```json
{
  "username": "admin",
  "email": "admin@example.com", 
  "password": "123456",
  "secret": "your-register-secret-password"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**错误响应**:
```json
{
  "error": "注册密令错误"
}
```

### 1.2 用户登录

**POST** `/api/auth/login`

**请求体**:
```json
{
  "username": "admin",
  "password": "123456"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin", 
    "email": "admin@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## 2. 文章管理

### 2.1 获取文章列表

**GET** `/api/articles`

**查询参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10  
- `status` (可选): 文章状态，`draft` 或 `published`

**示例**: `/api/articles?page=1&limit=10&status=published`

**响应**:
```json
{
  "articles": [
    {
      "id": 1,
      "title": "我的第一篇文章",
      "content": "# 标题\n\n这是文章内容...",
      "summary": "文章摘要",
      "status": "published",
      "user_id": 1,
      "user": {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 50,
  "page": 1, 
  "limit": 10
}
```

### 2.2 获取单篇文章

**GET** `/api/articles/:id`

**路径参数**:
- `id`: 文章ID

**响应**:
```json
{
  "id": 1,
  "title": "我的第一篇文章",
  "content": "# 标题\n\n这是文章内容...",
  "summary": "文章摘要", 
  "status": "published",
  "user_id": 1,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com"
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 2.3 创建文章 🔒

**POST** `/api/articles`

**请求体**:
```json
{
  "title": "文章标题",
  "content": "# 标题\n\n文章内容，支持Markdown格式",
  "summary": "文章摘要（可选）",
  "status": "draft"
}
```

**字段说明**:
- `title`: 必填，文章标题
- `content`: 必填，文章内容（Markdown格式）
- `summary`: 可选，文章摘要
- `status`: 可选，文章状态（`draft` 或 `published`），默认 `draft`

**响应**:
```json
{
  "id": 1,
  "title": "文章标题",
  "content": "# 标题\n\n文章内容，支持Markdown格式",
  "summary": "文章摘要",
  "status": "draft",
  "user_id": 1,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com"
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 2.4 更新文章 🔒

**PUT** `/api/articles/:id`

**路径参数**:
- `id`: 文章ID

**请求体**:
```json
{
  "title": "更新的标题", 
  "content": "更新的内容",
  "summary": "更新的摘要",
  "status": "published"
}
```

**响应**: 同创建文章

### 2.5 删除文章 🔒

**DELETE** `/api/articles/:id`

**路径参数**:
- `id`: 文章ID

**响应**:
```json
{
  "message": "文章删除成功"
}
```

---

## 3. 公共信息管理

### 3.1 获取公共信息

**GET** `/api/profile`

**响应**:
```json
{
  "id": 1,
  "name": "张三",
  "email": "zhangsan@example.com",
  "bio": "全栈开发工程师，热爱技术分享",
  "skills": "[\"JavaScript\", \"Go\", \"React\", \"Node.js\"]",
  "avatar": "https://example.com/avatar.jpg",
  "website": "https://zhangsan.dev",
  "github": "https://github.com/zhangsan",
  "linkedin": "https://linkedin.com/in/zhangsan", 
  "twitter": "https://twitter.com/zhangsan",
  "location": "北京",
  "company": "科技公司",
  "position": "高级工程师",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 3.2 更新公共信息 🔒

**PUT** `/api/profile`

**请求体**:
```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "bio": "全栈开发工程师，热爱技术分享",
  "skills": "[\"JavaScript\", \"Go\", \"React\", \"Node.js\"]",
  "avatar": "https://example.com/avatar.jpg",
  "website": "https://zhangsan.dev",
  "github": "https://github.com/zhangsan",
  "linkedin": "https://linkedin.com/in/zhangsan",
  "twitter": "https://twitter.com/zhangsan", 
  "location": "北京",
  "company": "科技公司",
  "position": "高级工程师"
}
```

**字段说明**:
- `name`: 必填，姓名
- 其他字段均为可选

**响应**: 同获取公共信息

---

## 4. 图片上传

### 4.1 上传图片 🔒

**POST** `/api/upload/image`

**请求类型**: `multipart/form-data`

**请求参数**:
- `image`: 图片文件

**支持格式**: jpg, jpeg, png, gif, webp
**文件大小**: 最大5MB

**响应**:
```json
{
  "url": "https://your-domain.com/images/1234567890_abcdef.jpg",
  "file_name": "avatar.jpg",
  "size": 1024000
}
```

### 4.2 删除图片 🔒

**DELETE** `/api/upload/image`

**请求体**:
```json
{
  "url": "https://your-domain.com/images/1234567890_abcdef.jpg"
}
```

**响应**:
```json
{
  "message": "图片删除成功"
}
```

---

## 5. 用户信息

### 5.1 获取当前用户信息 🔒

**GET** `/api/user/profile`

**响应**:
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "created_at": "2024-01-01T00:00:00Z", 
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## 6. 系统接口

### 6.1 健康检查

**GET** `/health`

**响应**:
```json
{
  "status": "ok",
  "message": "Blog server is running"
}
```

---

## 错误响应格式

所有错误响应都采用统一格式：

```json
{
  "error": "错误信息描述"
}
```

**常见错误状态码**:
- `400`: 请求参数错误
- `401`: 未授权（token无效或缺失）
- `403`: 禁止访问（权限不足）
- `404`: 资源不存在
- `409`: 资源冲突（如用户名已存在）
- `500`: 服务器内部错误

---

## 使用示例

### JavaScript (Fetch API)

```javascript
// 登录
const login = async () => {
  const response = await fetch('http://localhost:8080/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      username: 'admin',
      password: '123456'
    })
  });
  
  const data = await response.json();
  if (response.ok) {
    localStorage.setItem('token', data.token);
    return data;
  } else {
    throw new Error(data.error);
  }
};

// 创建文章
const createArticle = async (article) => {
  const token = localStorage.getItem('token');
  const response = await fetch('http://localhost:8080/api/articles', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(article)
  });
  
  const data = await response.json();
  if (response.ok) {
    return data;
  } else {
    throw new Error(data.error);
  }
};

// 上传图片
const uploadImage = async (file) => {
  const token = localStorage.getItem('token');
  const formData = new FormData();
  formData.append('image', file);
  
  const response = await fetch('http://localhost:8080/api/upload/image', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });
  
  const data = await response.json();
  if (response.ok) {
    return data;
  } else {
    throw new Error(data.error);
  }
};
```

### cURL 示例

```bash
# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'

# 获取文章列表
curl -X GET "http://localhost:8080/api/articles?page=1&limit=10"

# 创建文章（需要token）
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title":"测试文章","content":"# 标题\n\n内容","status":"published"}'

# 上传图片（需要token）
curl -X POST http://localhost:8080/api/upload/image \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@/path/to/image.jpg"
```

---

## 注意事项

1. **认证Token**: JWT token有24小时有效期，过期需要重新登录
2. **文件上传**: 图片上传仅支持指定格式，大小不超过5MB
3. **CORS**: 服务器已配置CORS，支持跨域请求
4. **日志记录**: 所有API调用都会被记录到数据库
5. **软删除**: 文章删除为软删除，不会真正删除数据