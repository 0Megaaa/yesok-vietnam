<script setup>
import { ref, onMounted } from 'vue'

const SVCS_DATA = ref([
  { id: 'company', icon: '🏢', color: '#1565C0', bg: '#E3F0FF', name: '公司注册', sub: '外资/内资 · 一站式办理', price: '¥3000', unit: '起', tags: ['合规', '全程代办'] },
  { id: 'visa', icon: '🛂', color: '#00897B', bg: '#E0F2F1', name: '签证办理', sub: '旅游/商务/工作签证', price: '¥600', unit: '起', tags: ['加急可办', '拒签退费'] },
  { id: 'bank', icon: '🏦', color: '#F57C00', bg: '#FFF3E0', name: '银行开户', sub: '企业账户 · 快速开通', price: '¥800', unit: '起', tags: ['本地银行', '高效安全'] },
  { id: 'airport', icon: '✈️', color: '#1565C0', bg: '#E3F0FF', name: '机场接机', sub: '落地无忧 · 中文司机', price: '¥280', unit: '起', tags: ['24h服务', '中文'] },
  { id: 'rent', icon: '🏠', color: '#7B1FA2', bg: '#F3E5F5', name: '租房找房', sub: '真实房源 · 精准匹配', price: '¥500', unit: '起', tags: ['真实房源', '1对1带看'] },
  { id: 'translate', icon: '🗣️', color: '#00695C', bg: '#E0F2F1', name: '翻译陪同', sub: '商务/生活翻译服务', price: '¥200', unit: '/时', tags: ['越南语', '商务'] },
  { id: 'car', icon: '🚗', color: '#E53935', bg: '#FFEBEE', name: '商务用车', sub: '专车服务 · 安全舒适', price: '¥350', unit: '起', tags: ['7座', '专人服务'] },
  { id: 'medical', icon: '🏥', color: '#00897B', bg: '#E0F2F1', name: '医疗陪诊', sub: '就医陪同 · 专业协助', price: '¥500', unit: '/天', tags: ['正规机构', '翻译'] },
  { id: 'vip', icon: '👑', color: '#B8860B', bg: '#FFF8E1', name: '高端通道', sub: '私密·专属·高效', price: '面议', unit: '', tags: ['VIP专属', '保密'] },
])

const hotServices = ref([])
const allServices = ref([])

// 用于「查看更多」锚点滚动
const svcAllSecRef = ref(null)

const scrollToAllServices = () => {
  if (svcAllSecRef.value) {
    uni.pageScrollTo({ selector: '#svc-all-sec', duration: 300, fail: () => {} })
  }
}

const goToDetail = (id) => {
  uni.navigateTo({ url: `/pages/service-detail/index?id=${id}` })
}

const placeOrder = (id) => {
  uni.navigateTo({ url: `/pages/service-detail/index?id=${id}` })
}

const goSearch = () => {
  uni.navigateTo({ url: '/pages/search/index' })
}

const contactManager = () => {
  uni.navigateTo({ url: '/pages/chat/index' })
}

onMounted(() => {
  hotServices.value = SVCS_DATA.value.slice(0, 3)
  allServices.value = SVCS_DATA.value.slice(3).filter(s => s.id !== 'vip')
})
</script>

<template>
  <view class="page" style="background:#F5F8FC;">

    <!-- ── Banner ── -->
    <view style="position:relative;overflow:hidden;flex-shrink:0;">
      <!-- SVG 背景图用 CSS gradient 代替，避免 base64 过大 -->
      <view style="width:100%;height:180px;background:linear-gradient(135deg,#0D2B6B 0%,#1565C0 60%,#1976D2 100%);display:block;"></view>

      <!-- Banner 叠加文字 -->
      <view style="position:absolute;inset:0;padding:20px 16px 16px;display:flex;flex-direction:column;justify-content:flex-end;">
        <view style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
          <text style="font-size:22px;font-weight:800;color:#fff;text-shadow:0 1px 4px rgba(0,0,0,.4);">专业服务</text>
          <text style="font-size:16px;">⭐</text>
        </view>
        <text style="font-size:13px;color:rgba(255,255,255,.85);margin-bottom:10px;text-shadow:0 1px 3px rgba(0,0,0,.3);">在越南，这些事我们帮你搞定</text>
        <view style="display:flex;gap:16px;">
          <view style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
            </svg>
            <text style="font-size:11px;color:rgba(255,255,255,.9);">本地团队</text>
          </view>
          <view style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <text style="font-size:11px;color:rgba(255,255,255,.9);">价格透明</text>
          </view>
          <view style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <text style="font-size:11px;color:rgba(255,255,255,.9);">品质保障</text>
          </view>
        </view>
      </view>

      <!-- 搜索按钮 -->
      <view
        @click="goSearch()"
        style="position:absolute;top:14px;right:14px;width:36px;height:36px;border-radius:50%;background:rgba(255,255,255,.9);display:flex;align-items:center;justify-content:center;cursor:pointer;box-shadow:0 2px 8px rgba(0,0,0,.15);">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#102A55" stroke-width="2.5">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
      </view>
    </view>

    <!-- ── 专属管家横幅 ── -->
    <view style="margin:12px;background:#fff;border-radius:16px;padding:14px 14px 14px 10px;box-shadow:0 4px 16px rgba(15,47,92,.1);border:1px solid #E8EFFE;display:flex;align-items:center;gap:12px;">
      <view style="position:relative;flex-shrink:0;">
        <image
          src="/static/manager-avatar.png"
          style="width:64px;height:64px;border-radius:50%;object-fit:cover;border:2px solid #E3F0FF;"
          mode="aspectFill"
          @error="(e) => { e.target.src = '/static/default-avatar.png' }"
        />
        <view style="position:absolute;bottom:1px;right:1px;width:12px;height:12px;border-radius:50%;background:#4CAF50;border:2px solid #fff;"></view>
      </view>
      <view style="flex:1;min-width:0;">
        <view style="display:flex;align-items:center;gap:6px;margin-bottom:3px;">
          <text style="font-size:13px;font-weight:700;color:#102A55;">专属管家</text>
          <text style="font-size:10px;background:#E3F0FF;color:#0D47A1;padding:1px 6px;border-radius:10px;font-weight:600;">一对一服务</text>
        </view>
        <text style="font-size:13px;font-weight:700;color:#102A55;display:block;margin-bottom:3px;">7×24小时在线 · 全程中文陪同</text>
        <text style="font-size:11px;color:#7A8799;display:block;margin-bottom:6px;">从咨询到完成，全程为您保驾护航</text>
        <view style="display:flex;gap:8px;">
          <text style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;">⚡快速响应</text>
          <text style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;">💰价格透明</text>
          <text style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;">🔒隐私保障</text>
        </view>
      </view>
      <view style="display:flex;flex-direction:column;align-items:flex-end;gap:6px;flex-shrink:0;">
        <button @click="contactManager()" style="padding:10px 16px;border-radius:22px;background:#0D47A1;color:#fff;font-size:12px;font-weight:700;border:none;cursor:pointer;white-space:nowrap;box-shadow:0 2px 8px rgba(13,71,161,.3);display:flex;align-items:center;gap:4px;">立即咨询 💬</button>
        <text style="font-size:9px;color:#9AA3B5;text-align:right;display:block;">已服务 3000+ 华人用户<br>平均响应 3 分钟</text>
      </view>
    </view>

    <!-- ── 热门服务 (Top 3) ── -->
    <view style="background:#fff;margin:0 0 8px;padding:14px 16px;">
      <view style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
        <view style="display:flex;align-items:center;gap:6px;">
          <text style="font-size:15px;">🔥</text>
          <text style="font-size:15px;font-weight:700;color:#102A55;">热门服务</text>
        </view>
        <text
          @click="scrollToAllServices()"
          style="font-size:12px;color:#0D47A1;cursor:pointer;">
          查看更多 ›
        </text>
      </view>
      <view style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px;">
        <view
          v-for="s in hotServices"
          :key="s.id"
          style="background:#fff;border-radius:14px;overflow:hidden;cursor:pointer;box-shadow:0 4px 14px rgba(15,47,92,.08);"
          @click="goToDetail(s.id)">
          <view :style="{ background: s.bg, padding: '14px 14px 10px' }">
            <text style="font-size:32px;display:block;margin-bottom:6px;">{{ s.icon }}</text>
          </view>
          <view style="padding:10px 14px 12px;">
            <view style="margin-bottom:8px;">
              <text style="font-size:14px;font-weight:700;color:#102A55;display:block;margin-bottom:4px;">{{ s.name }}</text>
              <text style="font-size:11px;color:#7A8799;display:block;margin-bottom:8px;">{{ s.sub }}</text>
              <view style="margin-bottom:8px;display:flex;gap:4px;flex-wrap:wrap;">
                <text
                  v-for="tag in s.tags.slice(0, 2)"
                  :key="tag"
                  :style="{ fontSize: '10px', background: s.color, color: '#fff', padding: '2px 8px', borderRadius: '8px', fontWeight: '600', marginRight: '4px', display: 'inline-block' }">
                  {{ tag }}
                </text>
              </view>
              <view style="display:flex;align-items:center;justify-content:space-between;">
                <text :style="{ fontSize: '15px', fontWeight: '800', color: s.color }">{{ s.price }}{{ s.unit }}</text>
                <text :style="{ fontSize: '11px', color: s.color, fontWeight: '600' }">了解详情 ›</text>
              </view>
            </view>
            <button
              @click.stop="placeOrder(s.id)"
              style="margin-top:8px;width:100%;padding:8px;border-radius:12px;"
              :style="{ background: s.color, color: '#fff', fontSize: '11px', fontWeight: '700', border: 'none', cursor: 'pointer' }">
              立即下单
            </button>
          </view>
        </view>
      </view>
    </view>

    <!-- ── 全部服务 ── -->
    <view id="svc-all-sec" ref="svcAllSecRef" style="background:#fff;margin-bottom:8px;padding:14px 16px;">
      <text style="font-size:15px;font-weight:700;color:#102A55;display:block;margin-bottom:12px;">全部服务</text>
      <view style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">
        <view
          v-for="s in allServices"
          :key="s.id"
          style="background:#fff;border-radius:14px;overflow:hidden;cursor:pointer;box-shadow:0 4px 14px rgba(15,47,92,.08);border:1px solid #EEF2F8;"
          @click="goToDetail(s.id)">
          <view style="padding:14px;">
            <view style="display:flex;align-items:center;gap:10px;margin-bottom:10px;">
              <view :style="{ width: '44px', height: '44px', borderRadius: '12px', background: s.bg, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '22px', flexShrink: '0' }">
                <text style="font-size:22px;">{{ s.icon }}</text>
              </view>
              <view>
                <text style="font-size:14px;font-weight:700;color:#102A55;display:block;">{{ s.name }}</text>
                <text style="font-size:11px;color:#7A8799;display:block;margin-top:1px;">{{ s.sub }}</text>
              </view>
            </view>
            <view style="display:flex;gap:4px;flex-wrap:wrap;margin-bottom:8px;">
              <text
                v-for="tag in s.tags.slice(0, 2)"
                :key="tag"
                :style="{ fontSize: '9px', background: s.bg, color: s.color, padding: '2px 7px', borderRadius: '8px', fontWeight: '600', display: 'inline-block' }">
                {{ tag }}
              </text>
            </view>
            <view style="display:flex;align-items:center;justify-content:space-between;">
              <text :style="{ fontSize: '14px', fontWeight: '800', color: s.color }">{{ s.price }}{{ s.unit }}</text>
              <view :style="{ width: '24px', height: '24px', borderRadius: '50%', background: s.bg, display: 'flex', alignItems: 'center', justifyContent: 'center' }">
                <text style="font-size:12px;color:#9AA3B5;">›</text>
              </view>
            </view>
            <button
              @click.stop="placeOrder(s.id)"
              style="margin-top:8px;width:100%;padding:8px;border-radius:12px;"
              :style="{ background: s.color, color: '#fff', fontSize: '11px', fontWeight: '700', border: 'none', cursor: 'pointer' }">
              立即下单
            </button>
          </view>
        </view>
      </view>
    </view>

    <!-- ── 高端通道 (VIP) ── -->
    <view style="margin:0 12px 14px;background:linear-gradient(135deg,#0A1628,#1A2A50);border-radius:16px;padding:16px;box-shadow:0 6px 20px rgba(10,22,40,.4);">
      <view style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px;">
        <view>
          <view style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
            <text style="font-size:18px;">👑</text>
            <text style="font-size:16px;font-weight:700;color:#F6B000;">高端通道服务</text>
            <text style="font-size:10px;background:rgba(246,176,0,.15);color:#F6B000;border:1px solid rgba(246,176,0,.3);padding:1px 8px;border-radius:10px;">私密·专属·高效</text>
          </view>
          <text style="font-size:11px;color:rgba(255,255,255,.55);display:block;">解决复杂问题，一对一资源对接，全程保密</text>
        </view>
        <button
          @click="goToDetail('vip')"
          style="padding:8px 14px;border-radius:20px;background:#F6B000;color:#0A1628;font-size:11px;font-weight:700;border:none;cursor:pointer;white-space:nowrap;">
          专属咨询 🔒
        </button>
      </view>
      <view style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:8px;margin-bottom:12px;">
        <view @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <text style="font-size:20px;display:block;margin-bottom:5px;">📋</text>
          <text style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;display:block;">企业审批<br>加速</text>
        </view>
        <view @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <text style="font-size:20px;display:block;margin-bottom:5px;">🤝</text>
          <text style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;display:block;">本地资源<br>对接</text>
        </view>
        <view @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <text style="font-size:20px;display:block;margin-bottom:5px;">⚖️</text>
          <text style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;display:block;">商务问题<br>解决</text>
        </view>
        <view @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <text style="font-size:20px;display:block;margin-bottom:5px;">🛡️</text>
          <text style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;display:block;">高端定制<br>服务</text>
        </view>
      </view>
      <view style="display:flex;justify-content:space-around;border-top:1px solid rgba(255,255,255,.08);padding-top:10px;">
        <text style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
          </svg>
          专属顾问全程跟进
        </text>
        <text style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          严格保密协议
        </text>
        <text style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          高效解决方案
        </text>
      </view>
    </view>

    <view style="height:8px;"></view>
  </view>
</template>

<style scoped>
/* 全局样式已由 style.css 接管，此处不留任何样式 */
</style>
