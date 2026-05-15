// Mock 数据源用于隔离真实后端，保证演示版可在微信小程序与 H5 环境中独立运行。
// 实现步骤：
// 1. 在一个文件内集中维护首页、登录、订单与服务详情数据。
// 2. 所有 API 文件只读取这里的数据，避免页面直接硬编码后端结构。
// 3. 后续联调真实接口时，只需要替换 api/request.js 或具体 API 方法。

export const MOCK_USER = {
  id: 10001,
  nickname: '越南体验官',
  avatarUrl: '/static/img.png',
  phone: '未绑定',
  role: 'client',
  vipLevel: 1,
  balance: 0,
  managerName: 'Linh 专属管家',
}

export const MOCK_SERVICES = [
  {
    id: 'company',
    icon: '🏢',
    color: '#1565C0',
    bg: '#E3F0FF',
    name: '公司注册',
    sub: '外资/内资 · 一站式办理',
    price: '¥3000',
    unit: '起',
    tags: ['合规', '全程代办'],
    description: '覆盖越南公司核名、注册地址、营业执照、税务登记与银行开户前置咨询。',
  },
  {
    id: 'visa',
    icon: '🛂',
    color: '#00897B',
    bg: '#E0F2F1',
    name: '签证办理',
    sub: '旅游/商务/工作签证',
    price: '¥600',
    unit: '起',
    tags: ['加急可办', '拒签退费'],
    description: '根据出行目的匹配签证方案，协助准备材料、递交申请并跟进进度。',
  },
  {
    id: 'bank',
    icon: '🏦',
    color: '#F57C00',
    bg: '#FFF3E0',
    name: '银行开户',
    sub: '企业账户 · 快速开通',
    price: '¥800',
    unit: '起',
    tags: ['本地银行', '高效安全'],
    description: '协助匹配越南本地银行，准备开户材料，并安排线下或线上开户流程。',
  },
  {
    id: 'airport',
    icon: '✈️',
    color: '#1565C0',
    bg: '#E3F0FF',
    name: '机场接机',
    sub: '落地无忧 · 中文司机',
    price: '¥280',
    unit: '起',
    tags: ['24h服务', '中文'],
    description: '胡志明、河内等主要机场中文接机，支持商务车、举牌与夜间航班。',
  },
  {
    id: 'rent',
    icon: '🏠',
    color: '#7B1FA2',
    bg: '#F3E5F5',
    name: '租房找房',
    sub: '真实房源 · 精准匹配',
    price: '¥500',
    unit: '起',
    tags: ['真实房源', '1对1带看'],
    description: '按预算、区域、通勤和居住偏好筛选房源，并提供中文陪同看房。',
  },
  {
    id: 'translate',
    icon: '🗣️',
    color: '#00695C',
    bg: '#E0F2F1',
    name: '翻译陪同',
    sub: '商务/生活翻译服务',
    price: '¥200',
    unit: '/时',
    tags: ['越南语', '商务'],
    description: '提供商务谈判、政府办事、医院陪诊、生活沟通等越南语翻译陪同。',
  },
  {
    id: 'car',
    icon: '🚗',
    color: '#E53935',
    bg: '#FFEBEE',
    name: '商务用车',
    sub: '专车服务 · 安全舒适',
    price: '¥350',
    unit: '起',
    tags: ['7座', '专人服务'],
    description: '覆盖市内商务出行、跨城包车、客户接待与日租司机服务。',
  },
  {
    id: 'medical',
    icon: '🏥',
    color: '#00897B',
    bg: '#E0F2F1',
    name: '医疗陪诊',
    sub: '就医陪同 · 专业协助',
    price: '¥500',
    unit: '/天',
    tags: ['正规机构', '翻译'],
    description: '协助预约医院、整理病历、现场翻译和就诊流程陪同。',
  },
  {
    id: 'vip',
    icon: '👑',
    color: '#B8860B',
    bg: '#FFF8E1',
    name: '高端通道',
    sub: '私密·专属·高效',
    price: '面议',
    unit: '',
    tags: ['VIP专属', '保密'],
    description: '为企业主、高净值客户与紧急事务提供定制化越南本地资源协调。',
  },
]

export const MOCK_NEWS_CATEGORIES = ['全部', '签证政策', '房产投资', '生活指南', '商务资讯']

export const MOCK_NEWS = [
  { id: 'n1', title: '越南签证最新政策解读：商务签与旅游签的本质区别', tag: '签证政策', source: '官方发布', date: '05-02', top: true },
  { id: 'n2', title: '胡志明市高端公寓租房避坑指南，这5个小区最受华人欢迎', tag: '房产投资', source: '管家精选', date: '05-01', top: true },
  { id: 'n3', title: '在越南开办外资公司的全流程解析及税务注意事项', tag: '商务资讯', source: '政策解读', date: '04-28' },
  { id: 'n4', title: '初到胡志明市：交通出行与防坑全攻略', tag: '生活指南', source: '用户分享', date: '04-25' },
]

export const MOCK_ORDERS = [
  {
    id: 'YS20260515001',
    orderNo: 'YS20260515001',
    serviceName: '公司注册',
    icon: '🏢',
    status: 'processing',
    sk: 'processing',
    managerName: 'Linh 专属管家',
    price: '¥3000 起',
  },
]

export const PLATFORM_GUARANTEES = [
  { icon: '🥇', title: '官方直营', desc: '标准流程透明交付' },
  { icon: '⚡', title: '极速响应', desc: '中越双语专员跟进' },
  { icon: '🛡️', title: '资金担保', desc: '节点验收后再确认' },
  { icon: '🇨🇳', title: '全中文服务', desc: '沟通无障碍' },
]
