# homelab-reader

一个跑在自己服务器上的轻量聚合阅读器：书架（上传 TXT 阅读）＋ RSS 资讯聚合。

纯 Go 标准库实现，SQLite 存储，前端为服务端渲染的内嵌模板，无任何第三方 Web 框架与前端构建链。

> 本项目定位是「自用」。RSS 的条目按需实时拉取、不落库；订阅源为全站共享（不区分用户）。

---

## 功能

- 账号体系：注册 / 登录 / 登出，基于服务端会话 Cookie
- 书架：上传 `.txt`（阅读）、上传 `.epub`（暂只入库不提供解析浏览），按 `offset/length` 分片读取正文
- RSS：订阅源增删查、按源分组、一键拉取所有源并解析条目、条目详情中间页 + 跳转原文
- 日/夜主题切换；灵动岛式悬浮导航

## 技术要点

| 项 | 说明 |
|---|---|
| 语言 | Go 1.24（仅标准库 `net/http` + `html/template`） |
| 存储 | SQLite（`mattn/go-sqlite3`，CGO），WAL 模式，`SetMaxOpenConns(1)` |
| 前端 | 内嵌模板 + 原生 JS，无框架/无构建 |
| 鉴权 | 登录即发 `session_token` Cookie（`HttpOnly`），过期前 3 天自动续期 7 天 |
| 部署 | 单二进制，静态资源 `/templates/` 直接通过 `http.FileServer` 提供 |

## 快速开始

```bash
go run .                 # 默认监听 0.0.0.0:8087
# 或指定端口
go run . -addr :9000
```

首次启动会自动建库 `./data/data.db` 并创建相关表；上传书籍存放在 `./books/`。

浏览器打开 `http://localhost:8087/dashboard`。

### 交叉编译

由于依赖 `mattn/go-sqlite3`（CGO），交叉编译时 **必须 `CGO_ENABLED=1`** 并带目标平台交叉编译器，不能 `=`0。

```bash
# Windows（需要 mingw 交叉编译器 x86_64-w64-mingw32-gcc）
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc go build -o hlreader.exe .

# Linux amd64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o hlreader .
```

## 目录结构

```
homelab-reader/
│
├── main.go                 # 程序入口：嵌入模板 -> 初始化 -> 启动服务
│
├── bootstrap/              # 装配层：把各部分接起来
│   ├── run.go              #   Run()：创建 data/ 目录、初始化 SQLite、拉起服务
│   └── web.go              #   注册全部路由、构造 *http.Server（超时/监听端口）
│
├── pkg/                    # 业务逻辑层，按职责分包
│   ├── models/             # models.go  数据模型（User/Session/Book/RSSFeed/RSSItem...）
│   ├── database/           # database.go SQLite 连接 + 自动建表（WAL、单连接）
│   ├── auth/               # auth.go    密码 sha256 哈希 / 校验 / session token 生成
│   ├── middleware/         # auth.go    会话鉴权中间件：校验 Cookie、按剩余有效期自动续期、注入 userID 到 context
│   ├── handlers/           # HTTP 处理层
│   │   ├── handlers.go     #   AppServer 结构 + Dashboard/Books/Rss/RssItem/User 页面渲染 + / 302 重定向
│   │   ├── auth.go         #   注册 / 登录 / 登出
│   │   ├── books.go        #   书籍列表 / 上传 / 删除 / 分片正文
│   │   └── rss.go          #   RSS 增删查 / 一键拉取（best-effort）
│   ├── reader/             # reader.go  TXT 分片读取（UTF-8 边界对齐防乱码）
│   └── rss/                # rss.go     RSS 地址校验、网络拉取 + XML 解析（限 5MB）
│
├── templates/              # 前端（go:embed 打包进二进制）
│   ├── *.html              #   dashboard/books/rss/rss_item/user 五个页面模板
│   ├── app.css             #   共享设计系统（日/夜主题）
│   └── app.js              #   共享前端逻辑（登录态 / 主题 / 导航 / 工具函数）
│
├── data/                   # 运行时生成：SQLite 数据库（data.db + WAL）
└── books/                  # 运行时生成：上传的书籍文件
```

依赖关系（自底向上）：
- `handlers` →（调用）`database` / `models` / `auth` / `reader` / `rss`
- `middleware` → `database` / `models`
- `bootstrap` →（组装）`handlers` / `middleware` / `database`
- `main.go` → `bootstrap` → 启动

## 数据存储

SQLite 复用单连接。表结构（`createTables` 自动建表）：

| 表 | 用途 | 备注 |
|---|---|---|
| `users` | 用户 | `username` 唯一，密码为 sha256 哈希 |
| `sessions` | 登录会话 | `token` 唯一，7 天过期，外键级联删除 |
| `rss_feeds` | RSS 订阅源 | `url` 唯一，**全站共享**，无用户归属 |
| `books` | 书籍元数据 | 按 `user_id` 归属，外键级联删除 |

## 页面路由（服务端渲染）

| 路径 | 说明 |
|---|---|
| `/` | 302 重定向到 `/dashboard` |
| `/dashboard` | 概览 |
| `/dashboard/books` | 书架（上传 / 阅读） |
| `/dashboard/rss` | 资讯（订阅源管理 / 刷新 / 条目展开） |
| `/dashboard/rss/item` | RSS 条目详情中间页（摘要 + 前往原文） |
| `/dashboard/user` | 我的 |
| `/templates/*` | 静态资源（css / js） |

## API 接口

除 `auth` 系列外，其余接口均需携带登录 Cookie（`session_token`）。

### 鉴权

| 方法 | 路径 | 请求体 | 成功 | 说明 |
|---|---|---|---|---|
| POST | `/api/auth/register` | `{"username","password"}` | `201` | 用户名重复返回 `409` |
| POST | `/api/auth/login` | `{"username","password"}` | `200`（写 Cookie） | 失败 `401` |
| POST | `/api/auth/logout` | — | `204` | 服务端删除会话并清 Cookie |

### 书架

| 方法 | 路径 | 参数 | 成功返回 | 说明 |
|---|---|---|---|---|
| GET | `/api/books` | — | `200` 数组 | 当前用户全部书籍（Book 结构） |
| POST | `/api/books` | multipart `file` | `201` `{"id":n}` | 仅 `.txt`/`.epub`，≤20MB |
| DELETE | `/api/books/{id}` | 路径 id | `204` | 同时删除文件；不存在 `404` |
| GET | `/api/books/{id}/content` | query `offset` `length` | `200` `{"content","is_end"}` | `length ≤ 1MB`（防 OOM） |

`content` 返回结构：

```json
{ "content": "…", "is_end": false }
```

### RSS

| 方法 | 路径 | 请求体 | 成功返回 | 说明 |
|---|---|---|---|---|
| GET | `/api/rss` | — | `200` 数组 | 全部订阅源（RSSFeed） |
| POST | `/api/rss` | `{"title","url","category"}` | `201` | URL 重复返回 `409` |
| DELETE | `/api/rss/{id}` | 路径 id | `204` | 不存在 `404` |
| POST | `/api/rss/fetch` | — | `200` | 拉取并解析所有源，best-effort |

`fetch` 返回结构（单个源失败不影响其余）：

```json
{
  "feeds": [
    { "id": 1, "title": "…", "url": "…", "category": "…", "items": [
        { "title": "…", "link": "…", "description": "…", "pub_date": "…" }
    ]}
  ],
  "failures": [ { "url": "…", "error": "…" } ]
}
```

## 已知限制

- RSS 解析仅支持 **RSS 2.0（`<item>`）**；Atom（`<entry>`，如 Go 官方博客、The Verge 等）无法解析出条目
- RSS 条目**实时拉取、不落库**，刷新随网络波动，失败源会出现在 `failures` 中
- TXT 若为 **GBK 编码**会出现乱码（未做转码）
- `.epub` 目前只入库，未提供在线浏览
- 订阅源全站共享，未按用户隔离；静态资源与业务同源，未做 CORS 配置（默认不跨域）