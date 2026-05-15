# YesOK 越南管家后端宝塔部署说明

本项目后端采用 **Go + Gin + GORM + MySQL**，部署目标是宝塔面板手动部署，当前版本明确不使用 Docker。生产环境中的数据库地址、密码、Telegram Bot Token 等敏感配置必须通过宝塔的环境变量或系统服务配置注入，禁止写入代码仓库。

## 部署步骤

| 步骤 | 操作 |
|---|---|
| 1 | 在宝塔中创建 MySQL 数据库，例如 `yesok_vn`。 |
| 2 | 在数据库管理中执行 `server/migrations/20260515_core_tables.sql`。 |
| 3 | 在 `web/` 目录执行前端构建，生成 `web/dist`。 |
| 4 | 在服务器上进入 `server/` 目录编译 Go 二进制。 |
| 5 | 通过宝塔进程守护或 systemd 启动服务，并注入环境变量。 |

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_PORT` | `7625` | Go 服务监听端口。 |
| `STATIC_DIR` | `../web/dist` | H5 前端构建产物目录。 |
| `DB_HOST` | `127.0.0.1` | MySQL 主机。 |
| `DB_PORT` | `3306` | MySQL 端口。 |
| `DB_USER` | `root` | MySQL 用户。 |
| `DB_PASSWORD` | 空 | MySQL 密码，生产必须配置。 |
| `DB_NAME` | `yesok_vn` | 数据库名。 |
| `TG_BOT_TOKEN` | 空 | Telegram 机器人 Token，当前前端演示版暂不联调。 |

## 注意事项

后端启动时会对 **7 张核心业务表** 执行 GORM AutoMigrate，但生产环境仍建议先手动执行 SQL 脚本并备份数据库。当前前端演示版已经通过 Mock API 与后端隔离，因此 H5 和小程序页面可以先独立验收视觉与交互，后续再逐步切换真实接口。
