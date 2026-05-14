<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const serviceId = ref('')
const serviceData = ref(null)
const loading = ref(true)

onLoad((options) => {
  serviceId.value = options.id || ''
  loadService(serviceId.value)
})

const SVCS_MAP = {
  company: {
    icon: '🏢', color: '#1565C0', bg: '#E3F0FF',
    name: '公司注册', price: '¥3000', unit: '起',
    tags: ['合规', '全程代办', '快速出签'],
    desc: '提供外资/内资公司一站式注册办理服务。从公司名称核准、注册资本确认、注册地址挂靠到营业执照申领，全程专业团队跟进，合规高效，省心省力。覆盖越南胡志明市、河内、岘港等主要城市。',
    includes: ['公司名称核准', '注册地址挂靠', '营业执照申领', '税务登记（5-7工作日）', '银行开户指导', '公章制作'],
    steps: [
      { n: '1', t: '提交材料', d: '提供护照、签证、公司名称、经营范围等基础材料' },
      { n: '2', t: '名称核准', d: '越南计划投资部系统提交，1-2工作日完成' },
      { n: '3', t: '注册审批', d: '向注册地址所在省/市计划投资厅提交，5-7工作日' },
      { n: '4', t: '领取执照', d: '获批后1工作日领取营业执照正副本' },
      { n: '5', t: '后续事项', d: '协助办理税务登记、银行开户、社保登记等' },
    ],
    hero: '🏢', heroBg: 'linear-gradient(135deg, #E3F0FF 0%, #C5D9F5 100%)',
  },
  visa: {
    icon: '🛂', color: '#00897B', bg: '#E0F2F1',
    name: '签证办理', price: '¥600', unit: '起',
    tags: ['加急可办', '拒签退费', '全程指导'],
    desc: '专业办理越南旅游签证、商务签证、工作签证（E/E1/E2/E3/E4）、探亲签证等多种类型。支持加急服务，拒签全额退款，材料简化，全程中文指导。',
    includes: ['签证类型评估', '材料清单指导', '表格填写', '代送代取', '入境指引'],
    steps: [
      { n: '1', t: '类型评估', d: '根据出行目的推荐最优签证类型' },
      { n: '2', t: '材料准备', d: '提供完整材料清单，全程指导准备' },
      { n: '3', t: '提交申请', d: '代为向领事馆/落地签批文系统提交' },
      { n: '4', t: '进度跟踪', d: '实时跟踪审批状态，第一时间通知结果' },
      { n: '5', t: '领取/出境', d: '加急1工作日出，E签证可协助落地接机' },
    ],
    hero: '🛂', heroBg: 'linear-gradient(135deg, #E0F2F1 0%, #B2DFDB 100%)',
  },
  bank: {
    icon: '🏦', color: '#F57C00', bg: '#FFF3E0',
    name: '银行开户', price: '¥800', unit: '起',
    tags: ['本地银行', '高效安全', '中文服务'],
    desc: '协助在越南主要本地银行（Vietcombank、TPBank、VPBank等）及外资银行开设企业账户或个人账户。材料准备、预约陪同、开户指导一站式服务。',
    includes: ['银行推荐选型', '材料清单', '预约开户时间', '专人全程陪同', '网银开通指导'],
    steps: [
      { n: '1', t: '银行选型', d: '根据业务需求推荐最合适的银行及账户类型' },
      { n: '2', t: '材料准备', d: '企业账户需公司注册文件、个人账户需护照签证' },
      { n: '3', t: '预约开户', d: '协助预约客户经理，缩短等待时间' },
      { n: '4', t: '现场办理', d: '专人陪同到银行，全程翻译指导，1-2小时完成' },
      { n: '5', t: '账户激活', d: '指导网银/APP开通，设置安全密码，存入激活金' },
    ],
    hero: '🏦', heroBg: 'linear-gradient(135deg, #FFF3E0 0%, #FFE0B2 100%)',
  },
  airport: {
    icon: '✈️', color: '#1565C0', bg: '#E3F0FF',
    name: '机场接机', price: '¥280', unit: '起',
    tags: ['24h服务', '中文司机', '落地无忧'],
    desc: '越南主要机场（胡志明新山一TAN SON NHAT、河内内排 NOI BAI、岘港 DANANG 等）接机服务。专业中文司机、准时准点、举牌接机，让您的越南之旅从第一步就安心。',
    includes: ['航班实时跟踪', '举牌接机', '中文司机', '行李协助', '免费等候90分钟', '车型任选（5座/7座/商务）'],
    steps: [
      { n: '1', t: '预约用车', d: '提供航班号、日期、人数、行李数量' },
      { n: '2', t: '确认车型', d: '根据人数匹配最佳车型，确认价格' },
      { n: '3', t: '航班监控', d: '提前2小时监控航班动态，确认准时' },
      { n: '4', t: '准时接机', d: '司机提前30分钟到达，举牌等候' },
      { n: '5', t: '送达目的地', d: '安全送达酒店/住所，协助搬运行李' },
    ],
    hero: '✈️', heroBg: 'linear-gradient(135deg, #E3F0FF 0%, #BBDEFB 100%)',
  },
  rent: {
    icon: '🏠', color: '#7B1FA2', bg: '#F3E5F5',
    name: '租房找房', price: '¥500', unit: '起',
    tags: ['真实房源', '1对1带看', '中介把关'],
    desc: '覆盖胡志明市1/2/3/7区、河内西湖/西湖/纸桥等热门区域。真实在租房源，专业经纪人1对1带看，帮您找到安全、舒适、性价比高的理想居所。',
    includes: ['区域需求沟通', '房源精准匹配', '实地带看陪同', '合同审核', '押金协商', '入住交接'],
    steps: [
      { n: '1', t: '需求沟通', d: '了解预算、区域、房型、入住时间等需求' },
      { n: '2', t: '房源筛选', d: '从真实在租房源中精选3-5套备选' },
      { n: '3', t: '实地看房', d: '经纪人1对1专车接送带看，现场介绍' },
      { n: '4', t: '价格协商', d: '专业经纪人协助谈判，争取最优价格' },
      { n: '5', t: '签约入住', d: '审核越文合同，协助办理入住交接' },
    ],
    hero: '🏠', heroBg: 'linear-gradient(135deg, #F3E5F5 0%, #E1BEE7 100%)',
  },
  translate: {
    icon: '🗣️', color: '#00695C', bg: '#E0F2F1',
    name: '翻译陪同', price: '¥200', unit: '/时',
    tags: ['越南语', '商务', '同声传译'],
    desc: '提供越南语商务陪同翻译、会议同声传译、合同文件翻译（附公证）、生活陪同翻译等服务。持证译员，专业可靠，保护商业机密。',
    includes: ['持证译员', '商务/生活双模式', '合同翻译+公证', '全程保密', '灵活计时/包天'],
    steps: [
      { n: '1', t: '需求确认', d: '明确翻译场景（商务/生活）、时长、语言方向' },
      { n: '2', t: '译员匹配', d: '根据场景匹配合适的越南语持证译员' },
      { n: '3', t: '译前准备', d: '提供相关文件提前熟悉专业词汇' },
      { n: '4', t: '现场翻译', d: '全程陪同，精准传达双方意图' },
      { n: '5', t: '译后整理', d: '提供会议纪要翻译、文件翻译终稿' },
    ],
    hero: '🗣️', heroBg: 'linear-gradient(135deg, #E0F2F1 0%, #B2DFDB 100%)',
  },
  car: {
    icon: '🚗', color: '#E53935', bg: '#FFEBEE',
    name: '商务用车', price: '¥350', unit: '起',
    tags: ['7座', '专人服务', '安全舒适'],
    desc: '越南全境商务包车服务，提供5座/7座/12座中巴等多种车型。专职司机，熟悉路况，全程待命。适合商务考察、团组接待、外出办事等多种场景。',
    includes: ['多款车型可选', '专职中文司机', '全天/半日/单程灵活', '车内WiFi/充电', '行程定制'],
    steps: [
      { n: '1', t: '行程规划', d: '告知出发地、目的地、途经点、用车时长' },
      { n: '2', t: '车型确认', d: '根据人数和行李量推荐最优车型' },
      { n: '3', t: '司机安排', d: '匹配熟悉该路线的专业司机' },
      { n: '4', t: '准时出发', d: '司机提前到达出发点，全程待命' },
      { n: '5', t: '安全送达', d: '送达目的地，协助行李搬运' },
    ],
    hero: '🚗', heroBg: 'linear-gradient(135deg, #FFEBEE 0%, #FFCDD2 100%)',
  },
  medical: {
    icon: '🏥', color: '#00897B', bg: '#E0F2F1',
    name: '医疗陪诊', price: '¥500', unit: '/天',
    tags: ['正规机构', '全程翻译', '就诊陪同'],
    desc: '在胡志明市国际医院（FV Hospital、Columbia Asia、BFVH等）及本地大型公立医院提供就医陪同服务。帮助预约挂号、全程翻译、协助与医生沟通、取药缴费一条龙。',
    includes: ['医院预约', '全程越/英语翻译', '医生沟通协助', '取药缴费', '病历整理', '复诊安排'],
    steps: [
      { n: '1', t: '病情了解', d: '详细了解症状、既往病史、就医诉求' },
      { n: '2', t: '医院匹配', d: '推荐合适的越南正规医院科室' },
      { n: '3', t: '预约挂号', d: '提前电话/线上预约，减少等候' },
      { n: '4', t: '就诊陪同', d: '全程翻译，协助与医生准确沟通病情' },
      { n: '5', t: '后续跟进', d: '整理病历、开药说明、复诊提醒' },
    ],
    hero: '🏥', heroBg: 'linear-gradient(135deg, #E0F2F1 0%, #B2DFDB 100%)',
  },
  vip: {
    icon: '👑', color: '#B8860B', bg: '#FFF8E1',
    name: '高端通道服务', price: '面议', unit: '',
    tags: ['VIP专属', '保密协议', '一对一'],
    desc: '针对高净值客户、企业高管、特殊需求用户，提供高度私密的专属服务。涵盖复杂审批加速、资源精准对接、高端定制全程服务，全程签订保密协议。',
    includes: ['专属顾问1对1', '全程保密协议', '复杂问题特办', '资源精准对接', '7×24小时响应', '服务结果兜底'],
    steps: [
      { n: '1', t: '需求评估', d: '深入了解您的核心需求，评估可操作性' },
      { n: '2', t: '方案制定', d: '量身定制专属解决方案' },
      { n: '3', t: '资源对接', d: '精准对接本地核心资源与渠道' },
      { n: '4', t: '全程跟进', d: '专属顾问7×24小时在线，全程督办' },
      { n: '5', t: '交付验收', d: '服务完成后交付结果，满意度回访' },
    ],
    hero: '👑', heroBg: 'linear-gradient(135deg, #FFF8E1 0%, #FFE082 100%)',
  },
}

const loadService = (id) => {
  loading.value = true
  const data = SVCS_MAP[id]
  if (data) {
    serviceData.value = data
  } else {
    // fallback: unknown service
    serviceData.value = {
      icon: '📋', color: '#1565C0', bg: '#E3F0FF',
      name: '服务详情', price: '面议', unit: '',
      tags: [],
      desc: '该服务详情正在准备中，请联系管家获取更多信息。',
      includes: [],
      steps: [],
      hero: '📋', heroBg: 'linear-gradient(135deg, #E3F0FF 0%, #C5D9F5 100%)',
    }
  }
  loading.value = false
}

const goBack = () => {
  uni.navigateBack()
}

const contactManager = () => {
  uni.switchTab({ url: '/pages/profile/index' })
  uni.showToast({ title: '请联系管家预约服务', icon: 'none' })
}
</script>

<template>
  <view class="page" style="background:#F5F8FC;min-height:100vh;padding-bottom:80px;">

    <!-- ── Hero 图 ── -->
    <view :style="{
      height: '220px',
      background: serviceData ? serviceData.heroBg : 'linear-gradient(135deg, #E3F0FF, #C5D9F5)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      position: 'relative',
    }">
      <!-- 返回按钮 -->
      <view
        @click="goBack"
        style="position:absolute;top:14px;left:14px;width:36px;height:36px;border-radius:50%;background:rgba(255,255,255,.9);display:flex;align-items:center;justify-content:center;box-shadow:0 2px 8px rgba(0,0,0,.12);border:1px solid #E2E8F4;">
        <text style="font-size:18px;">‹</text>
      </view>
      <!-- 分享/收藏 -->
      <view style="position:absolute;top:14px;right:14px;display:flex;gap:8px;">
        <view style="width:36px;height:36px;border-radius:50%;background:rgba(255,255,255,.9);display:flex;align-items:center;justify-content:center;box-shadow:0 2px 8px rgba(0,0,0,.12);border:1px solid #E2E8F4;">
          <text style="font-size:14px;">☆</text>
        </view>
      </view>
      <!-- 服务图标 -->
      <text style="font-size:88px;">{{ serviceData ? serviceData.hero : '📋' }}</text>
    </view>

    <!-- ── 服务标题区 ── -->
    <view style="background:#fff;padding:16px 16px 14px;box-shadow:0 1px 8px rgba(0,0,0,.05);margin-bottom:8px;">
      <view style="display:flex;align-items:flex-start;gap:12px;margin-bottom:10px;">
        <view :style="{
          width: '50px', height: '50px', borderRadius: '12px',
          background: serviceData ? serviceData.bg : '#E3F0FF',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontSize: '24px', flexShrink: 0
        }">
          {{ serviceData ? serviceData.icon : '📋' }}
        </view>
        <view style="flex:1;">
          <view style="font-size:18px;font-weight:700;color:#102A55;margin-bottom:4px;">
            {{ serviceData ? serviceData.name : '服务详情' }}
          </view>
          <view style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;">
            <text
              v-for="tag in (serviceData ? serviceData.tags : [])"
              :key="tag"
              style="font-size:10px;padding:2px 8px;border-radius:10px;background:#F0F6FF;color:#0D47A1;font-weight:600;margin-bottom:2px;">
              {{ tag }}
            </text>
          </view>
        </view>
      </view>
      <!-- 价格 -->
      <view style="display:flex;align-items:baseline;gap:4px;margin-bottom:12px;">
        <text style="font-size:22px;font-weight:800;color:#0D47A1;">
          {{ serviceData ? serviceData.price : '' }}
        </text>
        <text style="font-size:13px;color:#9AA3B5;">{{ serviceData ? serviceData.unit : '' }}</text>
      </view>
      <!-- 简介 -->
      <view style="font-size:13px;color:#4A5568;line-height:1.8;">
        {{ serviceData ? serviceData.desc : '' }}
      </view>
    </view>

    <!-- ── 服务内容 ── -->
    <view style="background:#fff;margin:0 0 8px;" class="sec-card">
      <view class="sec-hd">
        <view class="sec-t">
          <view class="sec-bar"></view>
          包含服务
        </view>
      </view>
      <view style="padding:0 16px 14px;">
        <view
          v-for="(item, i) in (serviceData ? serviceData.includes : [])"
          :key="i"
          style="display:flex;align-items:center;gap:10px;padding:9px 0;border-bottom:1px solid #F0F6FF;font-size:13px;color:#4A5568;">
          <text style="color:#2E7D32;font-size:15px;">✓</text>
          <text>{{ item }}</text>
        </view>
      </view>
    </view>

    <!-- ── 办理流程 ── -->
    <view style="background:#fff;margin:0 0 8px;" class="sec-card">
      <view class="sec-hd">
        <view class="sec-t">
          <view class="sec-bar"></view>
          办理流程
        </view>
      </view>
      <view style="padding:0 16px 14px;">
        <view
          v-for="(step, i) in (serviceData ? serviceData.steps : [])"
          :key="i"
          style="display:flex;gap:12px;padding:11px 0;border-bottom:1px solid #F0F6FF;position:relative;">
          <view
            :style="{
              width: '26px', height: '26px', borderRadius: '50%', flexShrink: 0,
              background: '#0D47A1', color: '#fff',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: '11px', fontWeight: '700', marginTop: '1px'
            }">
            {{ step.n }}
          </view>
          <view style="flex:1;">
            <view style="font-size:13px;font-weight:600;color:#102A55;margin-bottom:2px;">{{ step.t }}</view>
            <view style="font-size:12px;color:#7A8799;line-height:1.6;">{{ step.d }}</view>
          </view>
        </view>
      </view>
    </view>

    <!-- ── 温馨提示 ── -->
    <view style="background:#fff;margin:0 0 8px;" class="sec-card">
      <view class="sec-hd">
        <view class="sec-t">
          <view class="sec-bar"></view>
          温馨提示
        </view>
      </view>
      <view style="padding:0 16px 14px;">
        <view class="cb-tip">
          具体服务价格可能根据服务内容、服务时长、所在城市等因素有所调整，请以管家最终报价为准。
        </view>
        <view class="cb-warn">
          部分服务（如公司注册、银行开户）需提前准备完整材料，建议提前 3-5 个工作日联系管家预约。
        </view>
      </view>
    </view>

    <!-- ── 底部固定按钮 ── -->
    <view style="position:fixed;bottom:0;left:0;right:0;background:#fff;border-top:1px solid #E2E8F4;padding:10px 16px 12px;z-index:100;box-shadow:0 -2px 12px rgba(13,71,161,.08);">
      <view style="display:flex;gap:10px;">
        <!-- 咨询管家按钮 -->
        <view
          @click="contactManager"
          style="flex:1;padding:13px;border-radius:24px;background:#fff;border:1.5px solid #0D47A1;display:flex;align-items:center;justify-content:center;gap:6px;cursor:pointer;">
          <text style="font-size:16px;">💬</text>
          <text style="font-size:14px;font-weight:700;color:#0D47A1;">咨询管家</text>
        </view>
        <!-- 立即预约按钮 -->
        <view
          @click="contactManager"
          style="flex:2;padding:13px;border-radius:24px;background:linear-gradient(135deg, #0D47A1, #1565C0);display:flex;align-items:center;justify-content:center;gap:6px;cursor:pointer;box-shadow:0 4px 14px rgba(13,71,161,.35);">
          <text style="font-size:16px;">📅</text>
          <text style="font-size:14px;font-weight:700;color:#fff;">立即预约服务</text>
        </view>
      </view>
    </view>

  </view>
</template>

<style scoped>
</style>
