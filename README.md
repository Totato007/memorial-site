# 纪念站 — 轻社交互动纪念平台

基于 Go + Gin + GORM + SQLite 的轻量化社交纪念网站，专为阿里云 2核2G 轻量服务器设计。服务器渲染 HTML，零前端框架，极简风格，严控资源与流量消耗。

## 快速开始（本地测试）

```bash
# 1. 确保已安装 Go 1.21+
go version

# 2. 设置代理（国内加速）
set GOPROXY=https://goproxy.cn,direct

# 3. 编译运行
go build -o memorial-site.exe .
memorial-site.exe

# 4. 浏览器访问 http://localhost:8080
```

**默认管理员**: `admin` / `admin123`（首次启动自动创建）

## 功能清单

| 模块 | 说明 |
|---|---|
| 账号系统 | 手机号/账号注册登录，个人资料编辑，JWT Cookie 鉴权 |
| 好友系统 | 多人关系网络，支持情侣(限1)/好友/家人/自定义，在线状态显示 |
| 可见权限 | 四级权限 — 仅自己 / 仅情侣 / 仅好友 / 公开 |
| 心情日记 | 心情 + 文字 + 可选图片，可见范围选择，按月筛选 |
| 公共广场 | 展示全站用户公开分享的日记 |
| 未来计划 | CRUD + 分类筛选 + 状态切换 + 可见范围 |
| 互动留言 | 选择好友一对一聊天，文字 + 压缩图片，4 秒轮询 |
| 专属相册 | 创建相册、上传照片（自动压缩+缩略图） |
| 管理后台 | 独立入口，站点数据概览，用户启用/禁用 |

## 功能入口

**前台用户页面**（需登录）：

| 页面 | URL | 说明 |
|---|---|---|
| 首页仪表盘 | `/dashboard` | 好友列表 + 在线状态 + 快捷入口 |
| 好友列表 | `/friends` | 按类型分组展示所有好友 |
| 添加好友 | `/friends/add` | 搜索用户，选择关系类型 |
| 好友详情 | `/friends/:id` | 相处天数、纪念日、发消息 |
| 互动留言 | `/chat?with=:id` | 选择好友后一对一聊天 |
| 心情日记 | `/diary` | 写日记，选可见范围 |
| 公共广场 | `/public` | 浏览全站公开日记 |
| 未来计划 | `/plans` | 创建计划，选可见范围 |
| 专属相册 | `/albums` | 相册管理 + 照片上传 |
| 个人资料 | `/profile` | 改昵称、简介、密码 |

**管理员后台**（需 admin 角色）：

| 页面 | URL | 说明 |
|---|---|---|
| 管理概览 | `/admin/` | 用户数、关系数、消息数、照片数 |
| 用户管理 | `/admin/users` | 查看所有用户，启用/禁用 |
| 修改密码 | 侧边栏 → 修改密码 | 跳转到 `/profile` |

## 关系类型约束

| 类型 | 数量限制 | 说明 |
|---|---|---|
| 情侣 (couple) | 每人最多 1 个 | 双向绑定 |
| 好友 (friend) | 不限 | 默认类型 |
| 家人 (family) | 不限 | |
| 自定义 (custom) | 不限 | 可自定义名称 |

## 可见权限说明

| 级别 | 谁能看到 |
|---|---|
| `private` | 仅自己 |
| `couple` | 自己 + 所有情侣关系好友 |
| `friends` | 自己 + 所有活跃关系好友 |
| `public` | 全站所有登录用户（出现在公共广场） |

## 在线状态

每次请求自动更新最后活跃时间，好友列表中显示：
- 5 分钟内 → **在线**（绿色）
- 5 分钟 ~ 1 小时 → X 分钟前在线（黄色）
- 1 ~ 24 小时 → X 小时前在线（黄色）
- 超过 24 小时 → 离线（灰色）

## 技术栈

| 层 | 技术 | 说明 |
|---|---|---|
| 语言 | Go 1.21+ | 编译为 ~24MB 二进制，空闲内存 ~20MB |
| HTTP | Gin | 高性能路由框架 |
| 数据库 | SQLite (纯 Go 驱动) | 免部署、零运维、可交叉编译 |
| 认证 | JWT (httpOnly Cookie) | 72h 过期，无状态 |
| 模板 | Go html/template | 服务端渲染，无 SPA 框架 |
| 样式 | 纯 CSS (~300行) | 无框架，极简柔和风格 |
| JS | 原生 JS (~100行) | 仅聊天轮询和表单交互 |

## 项目结构

```
memorial-site/
├── main.go                     # 入口：Gin 引擎、中间件、路由注册
├── config/config.go            # 配置集中处（图片压缩、限流参数）
├── models/                     # GORM 数据模型 (7个)
├── database/db.go              # SQLite 初始化、自动迁移、管理员种子
├── handlers/                   # 请求处理器 (7个)
├── middleware/                  # JWT 鉴权 + 在线状态更新 + 管理员校验
├── services/                   # 密码哈希、图片压缩、日期计算、在线状态
├── templates/                  # HTML 模板 (16个)
├── static/                     # CSS + JS
├── uploads/                    # 上传文件目录
├── deploy/                     # 部署配置 (nginx/supervisor/.env)
└── data/                       # SQLite 数据库文件
```

## 交叉编译

```bash
# Windows 本地编译 Linux 版本（纯 Go，无需 CGO）
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o memorial-site .
```

## 阿里云 2C2G 部署

### 1. 服务器环境

```bash
# Ubuntu 18.04 / 20.04 / 22.04 均可
sudo apt update
sudo apt install nginx supervisor -y
```

### 2. 上传文件

```powershell
# PowerShell 中执行
scp memorial-site root@<服务器IP>:/opt/memorial/
scp -r templates/ static/ uploads/ deploy/ root@<服务器IP>:/opt/memorial/
```

### 3. 配置 Supervisor

```bash
sudo cp /opt/memorial/deploy/supervisord.conf /etc/supervisor/conf.d/memorial.conf
# 务必修改 JWT_SECRET 和 ADMIN_PASS
sudo vi /etc/supervisor/conf.d/memorial.conf
sudo mkdir -p /var/log/memorial /opt/memorial/data
sudo chown -R www-data:www-data /opt/memorial
sudo supervisorctl reread && sudo supervisorctl update
sudo supervisorctl start memorial
```

### 4. 配置 Nginx

```bash
sudo cp /opt/memorial/deploy/nginx.conf.example /etc/nginx/sites-available/memorial
sudo ln -s /etc/nginx/sites-available/memorial /etc/nginx/sites-enabled/
sudo vi /etc/nginx/sites-available/memorial  # 修改 server_name
sudo nginx -t && sudo systemctl reload nginx
```

### 5. 防火墙

```bash
sudo ufw allow 22 && sudo ufw allow 80 && sudo ufw allow 443
sudo ufw enable
```

## 迭代升级

```bash
# 1. 停服务 → 替换二进制 → 启动，数据库自动迁移
supervisorctl stop memorial
scp memorial-site root@<IP>:/opt/memorial/
scp -r templates/ root@<IP>:/opt/memorial/
supervisorctl start memorial
```

数据文件 `data/memorial.db` 和 `uploads/` 不动，用户数据不会丢失。

## 图片压缩配置

**config/config.go**:
```go
MaxImgSize:  2 << 20  // 最大上传 2MB
ImgMaxWidth: 800      // 宽度超过 800px 等比缩小
ImgQuality:  75       // JPEG 质量 1-100
```

压缩逻辑见 [services/image_service.go](services/image_service.go)：解码 → 等比缩放 → JPEG 重编码 → 200px 缩略图，典型输出 80-150KB。

## 安全措施

- bcrypt 密码哈希 (cost=10)
- JWT httpOnly SameSite Cookie (72h)
- GORM 参数化查询防 SQL 注入
- html/template 自动转义防 XSS
- 上传 MIME 服务端校验 + 重编码剥离 EXIF
- 管理员独立路由 + 角色校验中间件

## 注意事项

- **彻底禁用**: 视频上传、直播、动图批量加载、音频 — 全站无相关代码
- 聊天使用 HTTP 轮询（非 WebSocket），适合低频互动
- 仅支持 JPG/PNG/WebP 图片格式
- SQLite 数据库文件位于 `data/memorial.db`，建议定期备份
- 首次启动自动创建 `data/` 和 `uploads/` 子目录
- 纯 Go SQLite 驱动，交叉编译无需 CGO
