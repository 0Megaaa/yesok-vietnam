-- YesOK 越南管家核心业务表结构
-- 执行说明：
-- 1. 在宝塔面板 MySQL 管理中选择 yesok_vn 数据库。
-- 2. 执行本脚本创建 C 端客户、B 端员工、服务、流程、订单、轨迹和支付 7 张核心表。
-- 3. 生产环境请先备份数据库，再执行结构变更。

CREATE TABLE IF NOT EXISTS app_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  wechat_open_id VARCHAR(128) DEFAULT '' COMMENT '微信OpenID，用于微信小程序静默鉴权',
  telegram_id VARCHAR(128) DEFAULT '' COMMENT 'Telegram用户ID，预留Mini App登录',
  apple_id VARCHAR(128) DEFAULT '' COMMENT 'Apple用户ID，预留iOS登录',
  phone VARCHAR(32) DEFAULT '' COMMENT '手机号',
  nickname VARCHAR(64) DEFAULT '' COMMENT '用户昵称',
  avatar_url VARCHAR(512) DEFAULT '' COMMENT '用户头像地址',
  balance BIGINT NOT NULL DEFAULT 0 COMMENT '账户余额，单位为分',
  vip_level INT NOT NULL DEFAULT 0 COMMENT 'VIP等级，0为普通用户',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_app_users_wechat_open_id (wechat_open_id),
  KEY idx_app_users_telegram_id (telegram_id),
  KEY idx_app_users_apple_id (apple_id),
  KEY idx_app_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='C端客户表';

CREATE TABLE IF NOT EXISTS sys_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  username VARCHAR(64) NOT NULL COMMENT 'B端员工登录账号',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希值，禁止明文存储',
  real_name VARCHAR(64) DEFAULT '' COMMENT '员工真实姓名',
  role VARCHAR(32) NOT NULL DEFAULT 'manager' COMMENT '员工角色：admin管理员，manager经理',
  status INT NOT NULL DEFAULT 1 COMMENT '账号状态：1启用，0禁用',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_sys_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='B端员工表';

CREATE TABLE IF NOT EXISTS sys_services (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  service_name VARCHAR(128) NOT NULL COMMENT '服务名称',
  icon VARCHAR(512) DEFAULT '' COMMENT '服务图标地址或图标标识',
  base_price BIGINT NOT NULL DEFAULT 0 COMMENT '基础价格，单位为分',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务品类表';

CREATE TABLE IF NOT EXISTS sys_workflow_nodes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  service_id BIGINT UNSIGNED NOT NULL COMMENT '关联服务ID',
  current_status VARCHAR(64) NOT NULL COMMENT '当前状态码',
  button_name VARCHAR(64) NOT NULL COMMENT 'B端操作按钮名称',
  target_status VARCHAR(64) NOT NULL COMMENT '点击按钮后的目标状态码',
  required_material TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必传资料',
  trigger_payment TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否触发支付',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_sys_workflow_nodes_service_id (service_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='动态流程配置表';

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_no VARCHAR(64) NOT NULL COMMENT '订单号',
  app_user_id BIGINT UNSIGNED NOT NULL COMMENT 'C端客户ID',
  service_id BIGINT UNSIGNED NOT NULL COMMENT '服务品类ID',
  total_amount BIGINT NOT NULL DEFAULT 0 COMMENT '订单总金额，单位为分',
  current_status VARCHAR(64) NOT NULL COMMENT '当前状态码',
  form_data JSON DEFAULT NULL COMMENT '动态表单数据，存储各类业务自定义详情',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_no (order_no),
  KEY idx_orders_app_user_id (app_user_id),
  KEY idx_orders_service_id (service_id),
  KEY idx_orders_current_status (current_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主订单表';

CREATE TABLE IF NOT EXISTS order_timelines (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '关联订单ID',
  before_status VARCHAR(64) DEFAULT '' COMMENT '变更前状态码',
  after_status VARCHAR(64) NOT NULL COMMENT '变更后状态码',
  operator VARCHAR(128) DEFAULT '' COMMENT '操作人，可记录员工ID或系统标识',
  remark VARCHAR(1000) DEFAULT '' COMMENT '备注或对客留言',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_order_timelines_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态轨迹表';

CREATE TABLE IF NOT EXISTS payment_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '关联订单ID',
  app_user_id BIGINT UNSIGNED NOT NULL COMMENT '付款客户ID',
  pay_amount BIGINT NOT NULL DEFAULT 0 COMMENT '支付金额，单位为分',
  pay_method VARCHAR(32) NOT NULL COMMENT '支付方式：wechat微信，tg_pay Telegram支付，balance余额',
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '支付状态：pending待支付，success成功，failed失败，refunded已退款',
  third_trade_no VARCHAR(128) DEFAULT '' COMMENT '第三方支付流水号',
  created_at DATETIME(3) DEFAULT NULL COMMENT '创建时间',
  updated_at DATETIME(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_payment_records_order_id (order_id),
  KEY idx_payment_records_app_user_id (app_user_id),
  KEY idx_payment_records_third_trade_no (third_trade_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付财务对账表';
