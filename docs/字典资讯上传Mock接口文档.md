# Yesok Vietnam 字典、资讯与本地上传 Mock API 文档

本文档记录本次紧急改造新增的 **B 端内容配置能力** 与 **C 端动态资讯读取能力**。所有接口均挂载在 Go 服务统一前缀下，后台接口需要携带登录后返回的 `Authorization: Bearer <token>` 请求头，公开资讯接口无需鉴权。

## 验收范围

| 模块 | 文件与表 | 交付能力 | 验收入口 |
|---|---|---|---|
| 字典类型 | `sys_dict_types` / `server/models/sys_dict_type.go` | B 端可新增、编辑、删除枚举分类，例如服务分类、资讯分类、订单状态。 | `/api/admin/dict-types` |
| 字典数据 | `sys_dict_data` / `server/models/sys_dict_data.go` | B 端可维护枚举明细，支持停用状态 `status=0` 显式保存。 | `/api/admin/dict-data` |
| 资讯文章 | `sys_articles` / `server/models/sys_article.go` | B 端可发布、草稿、编辑和删除资讯，C 端首页与资讯 Tab 动态读取。 | `/api/admin/articles`、`/api/v1/articles` |
| 本地上传 | `server/uploads/YYYYMMDD/*` | B 端可上传封面图片，后端返回可静态访问的 `/uploads/...` URL。 | `/api/admin/upload` |
| C 端资讯页 | `web/src/pages/news/index.vue` | 新增四栏 Tabbar 的资讯页，按分类筛选后端资讯。 | `/pages/news/index` |

## 认证接口

后台登录沿用既有账号。沙盒种子数据会创建 `admin / 123456`，登录成功后将 `token` 用于后续后台请求。

```http
POST /api/admin/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456"
}
```

```json
{
  "token": "<jwt-token>",
  "user": {
    "id": 1,
    "username": "admin",
    "real_name": "Yesok 总管家",
    "role": "admin"
  }
}
```

## 字典类型 Mock

### 查询字典类型

```http
GET /api/admin/dict-types
Authorization: Bearer <token>
```

```json
{
  "list": [
    {
      "id": 1,
      "dict_name": "服务分类",
      "dict_code": "service_category",
      "status": 1,
      "remark": "C 端服务分类枚举",
      "created_at": "2026-05-16T00:00:00Z",
      "updated_at": "2026-05-16T00:00:00Z"
    },
    {
      "id": 2,
      "dict_name": "资讯分类",
      "dict_code": "article_category",
      "status": 1,
      "remark": "C 端资讯筛选枚举",
      "created_at": "2026-05-16T00:00:00Z",
      "updated_at": "2026-05-16T00:00:00Z"
    }
  ]
}
```

### 新增或更新字典类型

```http
POST /api/admin/dict-types
Authorization: Bearer <token>
Content-Type: application/json

{
  "dict_name": "资讯分类",
  "dict_code": "article_category",
  "status": 1,
  "remark": "C 端资讯筛选枚举"
}
```

```http
PUT /api/admin/dict-types/2
Authorization: Bearer <token>
Content-Type: application/json

{
  "dict_name": "资讯分类",
  "dict_code": "article_category",
  "status": 0,
  "remark": "临时停用验证"
}
```

```json
{
  "id": 2,
  "dict_name": "资讯分类",
  "dict_code": "article_category",
  "status": 0,
  "remark": "临时停用验证"
}
```

## 字典数据 Mock

### 查询字典数据

```http
GET /api/admin/dict-data?dict_code=article_category
Authorization: Bearer <token>
```

```json
{
  "list": [
    {
      "id": 1,
      "dict_code": "article_category",
      "dict_label": "落地指南",
      "dict_value": "guide",
      "sort_order": 1,
      "status": 1,
      "remark": "签证、接机、商务落地"
    },
    {
      "id": 2,
      "dict_code": "article_category",
      "dict_label": "城市灵感",
      "dict_value": "city",
      "sort_order": 2,
      "status": 1,
      "remark": "胡志明、河内、岘港内容"
    },
    {
      "id": 3,
      "dict_code": "article_category",
      "dict_label": "服务公告",
      "dict_value": "notice",
      "sort_order": 3,
      "status": 1,
      "remark": "服务政策与运营通知"
    }
  ]
}
```

### 新增或更新字典数据

```http
POST /api/admin/dict-data
Authorization: Bearer <token>
Content-Type: application/json

{
  "dict_code": "article_category",
  "dict_label": "城市灵感",
  "dict_value": "city",
  "sort_order": 2,
  "status": 1,
  "remark": "胡志明、河内、岘港内容"
}
```

```json
{
  "id": 2,
  "dict_code": "article_category",
  "dict_label": "城市灵感",
  "dict_value": "city",
  "sort_order": 2,
  "status": 1,
  "remark": "胡志明、河内、岘港内容"
}
```

## 资讯文章 Mock

### C 端公开资讯列表

`limit` 支持数字字符串，服务端会在默认值和上限之间做保护。该接口只返回 `status=1` 的已发布资讯，并按 `sort_order` 与 `id` 倒序展示。

```http
GET /api/v1/articles?limit=3
```

```json
{
  "list": [
    {
      "id": 1,
      "title": "抵达胡志明后的 6 小时黄金动线",
      "cover_img": "/static/img.png",
      "summary": "从机场接送、酒店入住到商务晚宴，Yesok 管家为高净值客户拆解首日抵达节奏。",
      "content": "抵达越南后的第一天决定了整趟行程的效率。建议提前锁定航班信息、车辆规格、酒店入住窗口与晚宴动线，由双语管家统一协调司机、酒店和餐厅。",
      "category": "guide",
      "author": "Yesok Vietnam",
      "status": 1,
      "sort_order": 1,
      "view_count": 168
    }
  ]
}
```

### B 端资讯管理

```http
GET /api/admin/articles
Authorization: Bearer <token>
```

```http
POST /api/admin/articles
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "签证加急资料准备清单",
  "cover_img": "/static/img.png",
  "summary": "护照、入境日期、酒店地址与联系人信息，提前准备可显著缩短办理时间。",
  "content": "签证加急的关键在于资料准确性。客户应提前确认护照有效期、入境日期、停留天数、越南联系人和酒店地址，管家会在提交前完成二次校验。",
  "category": "notice",
  "author": "Yesok Vietnam",
  "status": 0,
  "sort_order": 4,
  "view_count": 76
}
```

```json
{
  "id": 4,
  "title": "签证加急资料准备清单",
  "cover_img": "/static/img.png",
  "summary": "护照、入境日期、酒店地址与联系人信息，提前准备可显著缩短办理时间。",
  "content": "签证加急的关键在于资料准确性。客户应提前确认护照有效期、入境日期、停留天数、越南联系人和酒店地址，管家会在提交前完成二次校验。",
  "category": "notice",
  "author": "Yesok Vietnam",
  "status": 0,
  "sort_order": 4,
  "view_count": 76
}
```

> **验收重点**：本次修复已移除资讯与字典模型中 `status` 的数据库默认标签，并在新增处理器中强制写入零值字段，因此 `status=0` 可稳定保存为草稿或停用，而不会被默认值覆盖为 `1`。

## 本地上传 Mock

```http
POST /api/admin/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file=@cover.png
```

```json
{
  "url": "/uploads/20260516/cover_1778912345678900000.png",
  "name": "cover_1778912345678900000.png",
  "size": 128
}
```

后端将上传文件保存到 `server/uploads/YYYYMMDD/`，并通过 `r.Static("/uploads", "./uploads")` 暴露静态访问路径。B 端资讯编辑表单会把返回的 `url` 写入 `cover_img`，C 端首页资讯卡片和资讯 Tab 会直接使用该地址渲染封面。

## 沙盒自测记录

| 自测项 | 命令或入口 | 结果 |
|---|---|---|
| Go 后端编译测试 | `go test ./...` | 通过 |
| SQLite 沙盒服务启动 | `DB_DRIVER=sqlite DB_DSN=/tmp/yesok_vn_test.db SERVER_PORT=7625 go run .` | 通过 |
| 健康检查 | `GET /health` | 返回 `ok` |
| 公开资讯 limit | `GET /api/v1/articles?limit=1` | 返回 1 条发布资讯 |
| 后台登录 | `POST /api/admin/login` | 返回 token |
| 字典接口 | `GET /api/admin/dict-types`、`GET /api/admin/dict-data` | 返回种子字典 |
| 资讯草稿 | `POST /api/admin/articles`，`status=0` | 数据库保存为 `0` |
| 上传接口 | `POST /api/admin/upload` | 返回 `/uploads/...` JSON |
| 前端 H5 构建 | `pnpm build` | 通过 |
| C 端标签规范扫描 | `grep -RInE '<(/)?(button|img|svg|path|rect|line|polyline)\\b'` | 首页、资讯、服务、我的四页无残留 |

## 页面入口

| 端 | 页面 | 路径 |
|---|---|---|
| B 端 | 管理后台 | `/pages/admin/index` |
| C 端 | 首页 | `/pages/home/index` |
| C 端 | 资讯 Tab | `/pages/news/index` |
| C 端 | 服务 Tab | `/pages/services/index` |
| C 端 | 我的 Tab | `/pages/profile/index` |
