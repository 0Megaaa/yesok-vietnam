# Yesok Vietnam 2.0 全表 Mock JSON 数据结构文档

本文档由沙盒 SQLite 联调库自动导出，覆盖用户要求的 **8 张核心业务表**，并记录 B 端与 C 端完整数据闭环的真实样例。

| 表名 | 业务职责 | 样例行数 |
|---|---|---:|
| `app_users` | C 端客户用户表，承载微信/手机号客户画像。 | 5 |
| `sys_users` | B 端管家员工账号表，admin/123456 种子账号来自此表。 | 1 |
| `sys_services` | 服务配置表，由 B 端维护并动态驱动 C 端分类、价格、卡片与表单。 | 5 |
| `sys_workflow_nodes` | 流程大脑表，定义订单从待审核、报价、支付到完成的状态节点。 | 20 |
| `orders` | 订单主表，form_data 为 JSON 字段，保存客户提交的动态需求。 | 4 |
| `order_timelines` | 订单轨迹表，记录后台每次状态推进。 | 5 |
| `payment_records` | 财务对账流水表，后台确认支付时自动生成。 | 1 |
| `sys_configs` | 系统动态配置表，驱动 C 端 Banner、主题、热线等公开配置。 | 6 |

## 核心闭环说明

B 端通过 `sys_services` 配置服务名称、图标、价格、上下架与动态表单；C 端首页与服务页通过 `GET /api/v1/services` 实时读取这些配置。用户下单时，`POST /api/v1/orders` 将客户资料写入 `app_users`，将服务、金额、联系人与 `form_data` JSON 写入 `orders`，同时生成首条 `order_timelines`。管家后台推进订单状态时，系统继续写入 `order_timelines`，当状态进入 `paid` 或支付触发节点时，自动写入 `payment_records`，从而形成服务、订单、履约与财务闭环。

## 表结构与 Mock JSON

### app_users

C 端客户用户表，承载微信/手机号客户画像。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `wechat_open_id` | `TEXT` | 否 | 否 |
| `telegram_id` | `TEXT` | 否 | 否 |
| `apple_id` | `TEXT` | 否 | 否 |
| `phone` | `TEXT` | 否 | 否 |
| `nickname` | `TEXT` | 否 | 否 |
| `avatar_url` | `TEXT` | 否 | 否 |
| `balance` | `INTEGER` | 否 | 否 |
| `vip_level` | `INTEGER` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "wechat_open_id": "",
    "telegram_id": "",
    "apple_id": "",
    "phone": "+84901234567",
    "nickname": "陈先生",
    "avatar_url": "",
    "balance": 0,
    "vip_level": 2,
    "created_at": "2026-05-15 17:33:17.092598856+00:00",
    "updated_at": "2026-05-15 17:33:17.092598856+00:00"
  },
  {
    "id": 2,
    "wechat_open_id": "",
    "telegram_id": "",
    "apple_id": "",
    "phone": "+84999888778",
    "nickname": "沙盒用户",
    "avatar_url": "",
    "balance": 0,
    "vip_level": 1,
    "created_at": "2026-05-15 17:34:05.313690498+00:00",
    "updated_at": "2026-05-15 17:34:05.313690498+00:00"
  },
  {
    "id": 3,
    "wechat_open_id": "",
    "telegram_id": "",
    "apple_id": "",
    "phone": "+84999111222",
    "nickname": "闭环自测客户",
    "avatar_url": "",
    "balance": 0,
    "vip_level": 1,
    "created_at": "2026-05-15 17:38:56.500861487+00:00",
    "updated_at": "2026-05-15 17:38:56.500861487+00:00"
  },
  {
    "id": 4,
    "wechat_open_id": "",
    "telegram_id": "",
    "apple_id": "",
    "phone": "+84999111333",
    "nickname": "财务闭环客户",
    "avatar_url": "",
    "balance": 0,
    "vip_level": 1,
    "created_at": "2026-05-15 17:39:20.762254762+00:00",
    "updated_at": "2026-05-15 17:39:20.762254762+00:00"
  },
  {
    "id": 5,
    "wechat_open_id": "",
    "telegram_id": "",
    "apple_id": "",
    "phone": "+84999111444",
    "nickname": "财务闭环客户",
    "avatar_url": "",
    "balance": 0,
    "vip_level": 1,
    "created_at": "2026-05-15 17:40:17.298515752+00:00",
    "updated_at": "2026-05-15 17:40:17.298515752+00:00"
  }
]
```

### sys_users

B 端管家员工账号表，admin/123456 种子账号来自此表。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `username` | `TEXT` | 是 | 否 |
| `password_hash` | `TEXT` | 是 | 否 |
| `real_name` | `TEXT` | 否 | 否 |
| `role` | `TEXT` | 否 | 否 |
| `status` | `INTEGER` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "username": "admin",
    "password_hash": "$2a$10$Za8li6bqdi7BaGWoIcpFa.fxzbONmg0yhtW56YvDltRbdo4JCm7kS",
    "real_name": "Yesok 总管家",
    "role": "admin",
    "status": 1,
    "created_at": "2026-05-15 17:33:17.06625437+00:00",
    "updated_at": "2026-05-15 17:33:17.06625437+00:00"
  }
]
```

### sys_services

服务配置表，由 B 端维护并动态驱动 C 端分类、价格、卡片与表单。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `service_code` | `TEXT` | 是 | 否 |
| `service_name` | `TEXT` | 是 | 否 |
| `display_name` | `TEXT` | 否 | 否 |
| `icon` | `TEXT` | 否 | 否 |
| `cover_image` | `TEXT` | 否 | 否 |
| `description` | `TEXT` | 否 | 否 |
| `base_price` | `INTEGER` | 否 | 否 |
| `currency` | `TEXT` | 否 | 否 |
| `unit` | `TEXT` | 否 | 否 |
| `sort_order` | `INTEGER` | 否 | 否 |
| `status` | `INTEGER` | 否 | 否 |
| `is_hot` | `numeric` | 否 | 否 |
| `form_schema` | `json` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "service_code": "airport_transfer",
    "service_name": "越南机场接送",
    "display_name": "豪华接机",
    "icon": "✈️",
    "cover_image": "/static/img.png",
    "description": "双语管家举牌接机，商务车直达酒店。",
    "base_price": 65000000,
    "currency": "VND",
    "unit": "次",
    "sort_order": 1,
    "status": 1,
    "is_hot": 1,
    "form_schema": "{\"fields\":[\"flight_no\",\"arrival_time\",\"hotel_address\"]}",
    "created_at": "2026-05-15 17:33:17.067729398+00:00",
    "updated_at": "2026-05-15 17:39:28.688835276+00:00"
  },
  {
    "id": 2,
    "service_code": "visa",
    "service_name": "越南签证加急",
    "display_name": "签证加急",
    "icon": "🛂",
    "cover_image": "/static/img.png",
    "description": "商务、旅游、落地签资料审核与加急通道。",
    "base_price": 120000000,
    "currency": "VND",
    "unit": "单",
    "sort_order": 2,
    "status": 1,
    "is_hot": 1,
    "form_schema": "{\"fields\":[\"passport_name\",\"passport_no\",\"entry_date\"]}",
    "created_at": "2026-05-15 17:33:17.069639599+00:00",
    "updated_at": "2026-05-15 17:39:28.690049109+00:00"
  },
  {
    "id": 3,
    "service_code": "charter",
    "service_name": "商务包车",
    "display_name": "商务包车",
    "icon": "🚘",
    "cover_image": "/static/img.png",
    "description": "胡志明、河内、岘港商务包车与行程规划。",
    "base_price": 180000000,
    "currency": "VND",
    "unit": "天",
    "sort_order": 3,
    "status": 1,
    "is_hot": 1,
    "form_schema": "{\"fields\":[\"city\",\"use_date\",\"route\"]}",
    "created_at": "2026-05-15 17:33:17.070936562+00:00",
    "updated_at": "2026-05-15 17:39:28.690812692+00:00"
  },
  {
    "id": 4,
    "service_code": "translation",
    "service_name": "商务翻译",
    "display_name": "随行翻译",
    "icon": "🌐",
    "cover_image": "/static/img.png",
    "description": "中越英随行翻译、会议陪同与商务谈判支持。",
    "base_price": 150000000,
    "currency": "VND",
    "unit": "天",
    "sort_order": 4,
    "status": 1,
    "is_hot": 0,
    "form_schema": "{\"fields\":[\"language\",\"meeting_time\",\"scene\"]}",
    "created_at": "2026-05-15 17:33:17.071742509+00:00",
    "updated_at": "2026-05-15 17:39:28.691572897+00:00"
  },
  {
    "id": 5,
    "service_code": "business",
    "service_name": "企业落地咨询",
    "display_name": "企业落地",
    "icon": "🏢",
    "cover_image": "/static/img.png",
    "description": "公司注册、选址、财税和本地资源对接。",
    "base_price": 350000000,
    "currency": "VND",
    "unit": "项",
    "sort_order": 5,
    "status": 1,
    "is_hot": 0,
    "form_schema": "{\"fields\":[\"company_name\",\"industry\",\"need\"]}",
    "created_at": "2026-05-15 17:33:17.072365216+00:00",
    "updated_at": "2026-05-15 17:39:28.692763522+00:00"
  }
]
```

### sys_workflow_nodes

流程大脑表，定义订单从待审核、报价、支付到完成的状态节点。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `service_id` | `INTEGER` | 是 | 否 |
| `current_status` | `TEXT` | 是 | 否 |
| `button_name` | `TEXT` | 是 | 否 |
| `target_status` | `TEXT` | 是 | 否 |
| `required_material` | `numeric` | 否 | 否 |
| `trigger_payment` | `numeric` | 否 | 否 |
| `sort_order` | `INTEGER` | 否 | 否 |
| `remark` | `TEXT` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "service_id": 1,
    "current_status": "pending",
    "button_name": "去报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.072980815+00:00",
    "updated_at": "2026-05-15 17:33:17.072980815+00:00"
  },
  {
    "id": 2,
    "service_id": 1,
    "current_status": "quoted",
    "button_name": "确认收款",
    "target_status": "paid",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 2,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.073599252+00:00",
    "updated_at": "2026-05-15 17:33:17.073599252+00:00"
  },
  {
    "id": 3,
    "service_id": 1,
    "current_status": "paid",
    "button_name": "开始履约",
    "target_status": "in_progress",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 3,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.074205984+00:00",
    "updated_at": "2026-05-15 17:33:17.074205984+00:00"
  },
  {
    "id": 4,
    "service_id": 1,
    "current_status": "in_progress",
    "button_name": "完成订单",
    "target_status": "completed",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 4,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.074800629+00:00",
    "updated_at": "2026-05-15 17:33:17.074800629+00:00"
  },
  {
    "id": 5,
    "service_id": 2,
    "current_status": "pending",
    "button_name": "去报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.075565877+00:00",
    "updated_at": "2026-05-15 17:33:17.075565877+00:00"
  },
  {
    "id": 6,
    "service_id": 2,
    "current_status": "quoted",
    "button_name": "确认收款",
    "target_status": "paid",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 2,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.076567553+00:00",
    "updated_at": "2026-05-15 17:33:17.076567553+00:00"
  },
  {
    "id": 7,
    "service_id": 2,
    "current_status": "paid",
    "button_name": "开始履约",
    "target_status": "in_progress",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 3,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.077404487+00:00",
    "updated_at": "2026-05-15 17:33:17.077404487+00:00"
  },
  {
    "id": 8,
    "service_id": 2,
    "current_status": "in_progress",
    "button_name": "完成订单",
    "target_status": "completed",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 4,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.07809095+00:00",
    "updated_at": "2026-05-15 17:33:17.07809095+00:00"
  },
  {
    "id": 9,
    "service_id": 2,
    "current_status": "pending",
    "button_name": "审核资料",
    "target_status": "reviewing",
    "required_material": 1,
    "trigger_payment": 0,
    "sort_order": 0,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.078825133+00:00",
    "updated_at": "2026-05-15 17:33:17.078825133+00:00"
  },
  {
    "id": 10,
    "service_id": 2,
    "current_status": "reviewing",
    "button_name": "资料通过并报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.079649781+00:00",
    "updated_at": "2026-05-15 17:33:17.079649781+00:00"
  },
  {
    "id": 11,
    "service_id": 3,
    "current_status": "pending",
    "button_name": "去报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.08031433+00:00",
    "updated_at": "2026-05-15 17:33:17.08031433+00:00"
  },
  {
    "id": 12,
    "service_id": 3,
    "current_status": "quoted",
    "button_name": "确认收款",
    "target_status": "paid",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 2,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.080976115+00:00",
    "updated_at": "2026-05-15 17:33:17.080976115+00:00"
  },
  {
    "id": 13,
    "service_id": 3,
    "current_status": "paid",
    "button_name": "开始履约",
    "target_status": "in_progress",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 3,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.081568213+00:00",
    "updated_at": "2026-05-15 17:33:17.081568213+00:00"
  },
  {
    "id": 14,
    "service_id": 3,
    "current_status": "in_progress",
    "button_name": "完成订单",
    "target_status": "completed",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 4,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.082179874+00:00",
    "updated_at": "2026-05-15 17:33:17.082179874+00:00"
  },
  {
    "id": 15,
    "service_id": 4,
    "current_status": "pending",
    "button_name": "去报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.082767331+00:00",
    "updated_at": "2026-05-15 17:33:17.082767331+00:00"
  },
  {
    "id": 16,
    "service_id": 4,
    "current_status": "quoted",
    "button_name": "确认收款",
    "target_status": "paid",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 2,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.083350711+00:00",
    "updated_at": "2026-05-15 17:33:17.083350711+00:00"
  },
  {
    "id": 17,
    "service_id": 4,
    "current_status": "paid",
    "button_name": "开始履约",
    "target_status": "in_progress",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 3,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.083986674+00:00",
    "updated_at": "2026-05-15 17:33:17.083986674+00:00"
  },
  {
    "id": 18,
    "service_id": 4,
    "current_status": "in_progress",
    "button_name": "完成订单",
    "target_status": "completed",
    "required_material": 0,
    "trigger_payment": 0,
    "sort_order": 4,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.084603637+00:00",
    "updated_at": "2026-05-15 17:33:17.084603637+00:00"
  },
  {
    "id": 19,
    "service_id": 5,
    "current_status": "pending",
    "button_name": "去报价",
    "target_status": "quoted",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 1,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.085198395+00:00",
    "updated_at": "2026-05-15 17:33:17.085198395+00:00"
  },
  {
    "id": 20,
    "service_id": 5,
    "current_status": "quoted",
    "button_name": "确认收款",
    "target_status": "paid",
    "required_material": 0,
    "trigger_payment": 1,
    "sort_order": 2,
    "remark": "系统种子流程",
    "created_at": "2026-05-15 17:33:17.085812536+00:00",
    "updated_at": "2026-05-15 17:33:17.085812536+00:00"
  }
]
```

### orders

订单主表，form_data 为 JSON 字段，保存客户提交的动态需求。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `order_no` | `TEXT` | 是 | 否 |
| `app_user_id` | `INTEGER` | 是 | 否 |
| `service_id` | `INTEGER` | 是 | 否 |
| `service_name` | `TEXT` | 否 | 否 |
| `contact_name` | `TEXT` | 否 | 否 |
| `contact_phone` | `TEXT` | 否 | 否 |
| `total_amount` | `INTEGER` | 否 | 否 |
| `currency` | `TEXT` | 否 | 否 |
| `current_status` | `TEXT` | 是 | 否 |
| `payment_status` | `TEXT` | 否 | 否 |
| `form_data` | `json` | 否 | 否 |
| `remark` | `TEXT` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "order_no": "YS202605151734051731",
    "app_user_id": 2,
    "service_id": 1,
    "service_name": "豪华接机",
    "contact_name": "沙盒用户",
    "contact_phone": "+84999888778",
    "total_amount": 65000000,
    "currency": "VND",
    "current_status": "pending",
    "payment_status": "unpaid",
    "form_data": "{\"arrival_time\":\"2026-05-20 11:00\",\"flight_no\":\"VN888\",\"hotel_address\":\"Park Hyatt Saigon\",\"service_code\":\"airport_transfer\",\"service_name\":\"越南机场接送\",\"submitted_at\":\"2026-05-15T17:34:05Z\"}",
    "remark": "",
    "created_at": "2026-05-15 17:34:05.314711424+00:00",
    "updated_at": "2026-05-15 17:34:05.314711424+00:00"
  },
  {
    "id": 2,
    "order_no": "YS202605151738567900",
    "app_user_id": 3,
    "service_id": 1,
    "service_name": "豪华接机",
    "contact_name": "闭环自测客户",
    "contact_phone": "+84999111222",
    "total_amount": 65000000,
    "currency": "VND",
    "current_status": "pending",
    "payment_status": "unpaid",
    "form_data": "{\"arrival_time\":\"2026-05-21 09:30\",\"flight_no\":\"VN999\",\"hotel_address\":\"The Reverie Saigon\",\"service_code\":\"airport_transfer\",\"service_name\":\"越南机场接送\",\"submitted_at\":\"2026-05-15T17:38:56Z\"}",
    "remark": "",
    "created_at": "2026-05-15 17:38:56.502468874+00:00",
    "updated_at": "2026-05-15 17:38:56.502468874+00:00"
  },
  {
    "id": 3,
    "order_no": "YS202605151739204977",
    "app_user_id": 4,
    "service_id": 1,
    "service_name": "豪华接机",
    "contact_name": "财务闭环客户",
    "contact_phone": "+84999111333",
    "total_amount": 65000000,
    "currency": "VND",
    "current_status": "pending",
    "payment_status": "unpaid",
    "form_data": "{\"flight_no\":\"VN888\",\"hotel_address\":\"Park Hyatt Saigon\",\"service_code\":\"airport_transfer\",\"service_name\":\"越南机场接送\",\"submitted_at\":\"2026-05-15T17:39:20Z\"}",
    "remark": "",
    "created_at": "2026-05-15 17:39:20.763211517+00:00",
    "updated_at": "2026-05-15 17:39:20.763211517+00:00"
  },
  {
    "id": 4,
    "order_no": "YS202605151740179181",
    "app_user_id": 5,
    "service_id": 1,
    "service_name": "豪华接机",
    "contact_name": "财务闭环客户",
    "contact_phone": "+84999111444",
    "total_amount": 65000000,
    "currency": "VND",
    "current_status": "paid",
    "payment_status": "paid",
    "form_data": "{\"flight_no\":\"VN777\",\"hotel_address\":\"Park Hyatt Saigon\",\"service_code\":\"airport_transfer\",\"service_name\":\"越南机场接送\",\"submitted_at\":\"2026-05-15T17:40:17Z\"}",
    "remark": "闭环支付确认",
    "created_at": "2026-05-15 17:40:17.299714409+00:00",
    "updated_at": "2026-05-15 17:40:17.335370235+00:00"
  }
]
```

### order_timelines

订单轨迹表，记录后台每次状态推进。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `order_id` | `INTEGER` | 是 | 否 |
| `before_status` | `TEXT` | 否 | 否 |
| `after_status` | `TEXT` | 是 | 否 |
| `operator` | `TEXT` | 否 | 否 |
| `remark` | `TEXT` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "order_id": 1,
    "before_status": "",
    "after_status": "pending",
    "operator": "C端客户",
    "remark": "客户提交订单，等待管家处理",
    "created_at": "2026-05-15 17:34:05.315009673+00:00",
    "updated_at": "2026-05-15 17:34:05.315009673+00:00"
  },
  {
    "id": 2,
    "order_id": 2,
    "before_status": "",
    "after_status": "pending",
    "operator": "C端客户",
    "remark": "客户提交订单，等待管家处理",
    "created_at": "2026-05-15 17:38:56.502833004+00:00",
    "updated_at": "2026-05-15 17:38:56.502833004+00:00"
  },
  {
    "id": 3,
    "order_id": 3,
    "before_status": "",
    "after_status": "pending",
    "operator": "C端客户",
    "remark": "客户提交订单，等待管家处理",
    "created_at": "2026-05-15 17:39:20.763449989+00:00",
    "updated_at": "2026-05-15 17:39:20.763449989+00:00"
  },
  {
    "id": 4,
    "order_id": 4,
    "before_status": "",
    "after_status": "pending",
    "operator": "C端客户",
    "remark": "客户提交订单，等待管家处理",
    "created_at": "2026-05-15 17:40:17.300048507+00:00",
    "updated_at": "2026-05-15 17:40:17.300048507+00:00"
  },
  {
    "id": 5,
    "order_id": 4,
    "before_status": "pending",
    "after_status": "paid",
    "operator": "后台管家",
    "remark": "闭环支付确认",
    "created_at": "2026-05-15 17:40:17.33555669+00:00",
    "updated_at": "2026-05-15 17:40:17.33555669+00:00"
  }
]
```

### payment_records

财务对账流水表，后台确认支付时自动生成。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `order_id` | `INTEGER` | 是 | 否 |
| `app_user_id` | `INTEGER` | 是 | 否 |
| `payer_name` | `TEXT` | 否 | 否 |
| `pay_amount` | `INTEGER` | 否 | 否 |
| `pay_method` | `TEXT` | 是 | 否 |
| `status` | `TEXT` | 否 | 否 |
| `third_trade_no` | `TEXT` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "order_id": 4,
    "app_user_id": 5,
    "payer_name": "财务闭环客户",
    "pay_amount": 65000000,
    "pay_method": "bank_transfer",
    "status": "success",
    "third_trade_no": "YS-PAY-1778866817335758923",
    "created_at": "2026-05-15 17:40:17.335779623+00:00",
    "updated_at": "2026-05-15 17:40:17.335779623+00:00"
  }
]
```

### sys_configs

系统动态配置表，驱动 C 端 Banner、主题、热线等公开配置。

| 字段 | 类型 | 必填 | 主键 |
|---|---|---|---|
| `id` | `INTEGER` | 否 | 是 |
| `config_key` | `TEXT` | 是 | 否 |
| `config_value` | `TEXT` | 否 | 否 |
| `value_type` | `TEXT` | 否 | 否 |
| `group_name` | `TEXT` | 否 | 否 |
| `remark` | `TEXT` | 否 | 否 |
| `is_public` | `numeric` | 否 | 否 |
| `created_at` | `datetime` | 否 | 否 |
| `updated_at` | `datetime` | 否 | 否 |

```json
[
  {
    "id": 1,
    "config_key": "app_name",
    "config_value": "Yesok Vietnam",
    "value_type": "string",
    "group_name": "brand",
    "remark": "应用名称",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.088512563+00:00",
    "updated_at": "2026-05-15 17:39:28.694808964+00:00"
  },
  {
    "id": 2,
    "config_key": "hero_title",
    "config_value": "越南高端生活服务管家",
    "value_type": "string",
    "group_name": "home",
    "remark": "首页主标题",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.089148604+00:00",
    "updated_at": "2026-05-15 17:39:28.69556349+00:00"
  },
  {
    "id": 3,
    "config_key": "hero_subtitle",
    "config_value": "接机、签证、包车、翻译、企业落地一站式托管",
    "value_type": "string",
    "group_name": "home",
    "remark": "首页副标题",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.089814872+00:00",
    "updated_at": "2026-05-15 17:39:28.69634133+00:00"
  },
  {
    "id": 4,
    "config_key": "banner_image",
    "config_value": "/static/img.png",
    "value_type": "string",
    "group_name": "home",
    "remark": "首页 Banner 图",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.0906035+00:00",
    "updated_at": "2026-05-15 17:39:28.69704508+00:00"
  },
  {
    "id": 5,
    "config_key": "primary_color",
    "config_value": "#0F3D3E",
    "value_type": "string",
    "group_name": "theme",
    "remark": "主色",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.091321365+00:00",
    "updated_at": "2026-05-15 17:39:28.697673697+00:00"
  },
  {
    "id": 6,
    "config_key": "hotline",
    "config_value": "+84 888 666 168",
    "value_type": "string",
    "group_name": "contact",
    "remark": "管家热线",
    "is_public": 1,
    "created_at": "2026-05-15 17:33:17.092015965+00:00",
    "updated_at": "2026-05-15 17:39:28.698325134+00:00"
  }
]
```
