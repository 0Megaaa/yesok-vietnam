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

onMounted(() => {
  hotServices.value = SVCS_DATA.value.slice(0, 3)
  allServices.value = SVCS_DATA.value.slice(3).filter(s => s.id !== 'vip')
})

const goToDetail = (id) => {
  uni.navigateTo({ url: `/pages/service-detail/index?id=${id}` })
}

const placeOrder = (id) => {
  uni.navigateTo({ url: `/pages/service-detail/index?id=${id}` })
}
</script>

<template>
  <div class="page" style="background:#F5F8FC;">

    <!-- ── Banner ── -->
    <div style="position:relative;overflow:hidden;flex-shrink:0;">
      <img
        src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0naHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmcnIHZpZXdCb3g9JzAgMCA0MzAgMTgwJyB3aWR0aD0nNDMwJyBoZWlnaHQ9JzE4MCc+CiAgPGRlZnM+CiAgICA8bGluZWFyR3JhZGllbnQgaWQ9J3Nza3knIHgxPScwJyB5MT0nMCcgeDI9JzAnIHkyPScxJz4KICAgICAgPHN0b3Agb2Zmc2V0PScwJScgc3RvcC1jb2xvcj0nIzBEMkI2QicvPjxzdG9wIG9mZnNldD0nMTAwJScgc3RvcC1jb2xvcj0nIzE1NjVDMCcvPgogICAgPC9saW5lYXJHcmFkaWVudD4KICAgIDxsaW5lYXJHcmFkaWVudCBpZD0nc292ZXInIHgxPScwJyB5MT0nMCcgeDI9JzAnIHkyPScxJz4KICAgICAgPHN0b3Agb2Zmc2V0PScwJScgc3RvcC1jb2xvcj0nIzA2MTY0MCcgc3RvcC1vcGFjaXR5PScuNicvPjxzdG9wIG9mZnNldD0nNjAlJyBzdG9wLWNvbG9yPScjMDYxNjQwJyBzdG9wLW9wYWNpdHk9Jy4xNScvPjxzdG9wIG9mZnNldD0nMTAwJScgc3RvcC1jb2xvcj0nIzA2MTY0MCcgc3RvcC1vcGFjaXR5PScwJy8+CiAgICA8L2xpbmVhckdyYWRpZW50PgogIDwvZGVmcz4KICA8cmVjdCB3aWR0aD0nNDMwJyBoZWlnaHQ9JzE4MCcgZmlsbD0ndXJsKCNzc2t5KScvPgogIDwhLS0gY2l0eSBnbG93IC0tPgogIDxlbGxpcHNlIGN4PSczMjAnIGN5PScxNjAnIHJ4PScxMjAnIHJ5PSc0MCcgZmlsbD0nI0Y2QjAwMCcgb3BhY2l0eT0nLjEyJy8+CiAgPCEtLSBMYW5kbWFyayA4MSAtLT4KICA8cmVjdCB4PScyODAnIHk9JzMwJyB3aWR0aD0nMjQnIGhlaWdodD0nMTMwJyBmaWxsPScjMTk3NkQyJyBvcGFjaXR5PScuOScvPgogIDxwb2x5Z29uIHBvaW50cz0nMjkyLDIyIDI4MiwzMiAzMDIsMzInIGZpbGw9JyMwRDQ3QTEnLz4KICA8cmVjdCB4PScyODInIHk9JzMyJyB3aWR0aD0nMjAnIGhlaWdodD0nMTI4JyBmaWxsPScjMjE5NkYzJyBvcGFjaXR5PScuOCcvPgogIDwhLS0gd2luZG93cyBMYW5kbWFyayA4MSAtLT4KICA8cmVjdCB4PScyODQnIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMzUnLz4KICA8cmVjdCB4PScyOTAnIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMzUnLz4KICA8cmVjdCB4PScyOTYnIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMzUnLz4KICA8cmVjdCB4PScyODQnIHk9JzQzJyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMycvPgogIDxyZWN0IHg9JzI5MCcgeT0nNDMnIHdpZHRoPSc0JyBoZWlnaHQ9JzMnIGZpbGw9JyNmZmYnIG9wYWNpdHk9Jy4zJy8+CiAgPHJlY3QgeD0nMjk2JyB5PSc0Mycgd2lkdGg9JzQnIGhlaWdodD0nMycgZmlsbD0nI2ZmZicgb3BhY2l0eT0nLjMnLz4KICA8cmVjdCB4PScyODQnIHk9JzUwJyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMjUnLz4KICA8cmVjdCB4PScyOTAnIHk9JzUwJyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMjUnLz4KICA8IS0tIGJ1aWxkaW5ncyByaWdodCAtLT4KICA8cmVjdCB4PSczMDgnIHk9JzYwJyB3aWR0aD0nMjgnIGhlaWdodD0nMTAwJyBmaWxsPScjMTU2NUMwJyBvcGFjaXR5PScuOCcvPgogIDxyZWN0IHg9JzMzNCcgeT0nNzAnIHdpZHRoPScyMicgaGVpZ2h0PSc5MCcgZmlsbD0nIzFBNEU4QScgb3BhY2l0eT0nLjc1Jy8+CiAgPHJlY3QgeD0nMzU0JyB5PSc1NScgd2lkdGg9JzMwJyBoZWlnaHQ9JzEwNScgZmlsbD0nIzE5NzZEMicgb3BhY2l0eT0nLjcrJz4KICA8cmVjdCB4PSczODInIHk9Jzc1JyB3aWR0aD0nMjUnIGhlaWdodD0nODUnIGZpbGw9JyMxNTY1QzAnIG9wYWNpdHk9Jy42NScvPgogIDxyZWN0IHg9JzQwNScgeT0nNjUnIHdpZHRoPScyNScgaGVpZ2h0PSc5NScgZmlsbD0nIzFBNEU4QScgb3BhY2l0eT0nLjcnLz4KICA8IS0tIGJ1aWxkaW5ncyBsZWZ0IC0tPgogIDxyZWN0IHg9JzAnIHk9JzgwJyB3aWR0aD0nMzAnIGhlaWdodD0nODAnIGZpbGw9JyMxNTY1QzAnIG9wYWNpdHk9Jy42NScvPgogIDxyZWN0IHg9JzI4JyB5PSc2NScgd2lkdGg9JzI4JyBoZWlnaHQ9Jzk1JyBmaWxsPScjMUE0RThBJyBvcGFjaXR5PScuNycvPgogIDxyZWN0IHg9JzU0JyB5PSc3NScgd2lkdGg9JzIyJyBoZWlnaHQ9Jzg1JyBmaWxsPScjMTk3NkQyJyBvcGFjaXR5PScuNjUnLz4KICA8cmVjdCB4PSc3NCcgeT0nNjAnIHdpZHRoPSczMCcgaGVpZ2h0PScxMDAnIGZpbGw9JyMxNTY1QzAnIG9wYWNpdHk9Jy43Jy8+CiAgPHJlY3QgeD0nMTAyJyB5PSc3MCcgd2lkdGg9JzI0JyBoZWlnaHQ9JzkwJyBmaWxsPScjMUE0RThBJyBvcGFjaXR5PScuNjUnLz4KICA8IS0tIHJpdmVyL2dyb3VuZCAtLT4KICA8cmVjdCB5PScxNTUnIHdpZHRoPSc0MzAnIGhlaWdodD0nMjUnIGZpbGw9JyMwRDQ3QTEnIG9wYWNpdHk9Jy41Jy8+CiAgPCEtLSBncmFkaWVudCBvdmVybGF5IC0tPgogIDxyZWN0IHdpZHRoPSc0MzAnIGhlaWdodD0nMTgwJyBmaWxsPSd1cmwoI3NvdmVyKScvPgo8L3N2Zz4="
        style="width:100%;height:180px;object-fit:cover;display:block;" alt="">
      <div style="position:absolute;inset:0;padding:20px 16px 16px;display:flex;flex-direction:column;justify-content:flex-end;">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
          <div style="font-size:22px;font-weight:800;color:#fff;text-shadow:0 1px 4px rgba(0,0,0,.4);">专业服务</div>
          <span style="font-size:16px;">⭐</span>
        </div>
        <div style="font-size:13px;color:rgba(255,255,255,.85);margin-bottom:10px;text-shadow:0 1px 3px rgba(0,0,0,.3);">在越南，这些事我们帮你搞定</div>
        <div style="display:flex;gap:16px;">
          <div style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
            </svg>
            <span style="font-size:11px;color:rgba(255,255,255,.9);">本地团队</span>
          </div>
          <div style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <span style="font-size:11px;color:rgba(255,255,255,.9);">价格透明</span>
          </div>
          <div style="display:flex;align-items:center;gap:4px;">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.9)" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span style="font-size:11px;color:rgba(255,255,255,.9);">品质保障</span>
          </div>
        </div>
      </div>
      <!-- search -->
      <div
        style="position:absolute;top:14px;right:14px;width:36px;height:36px;border-radius:50%;background:rgba(255,255,255,.9);display:flex;align-items:center;justify-content:center;cursor:pointer;box-shadow:0 2px 8px rgba(0,0,0,.15);">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#102A55" stroke-width="2.5">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
      </div>
    </div>

    <!-- ── 专属管家横幅 ── -->
    <div style="margin:12px;background:#fff;border-radius:16px;padding:14px 14px 14px 10px;box-shadow:0 4px 16px rgba(15,47,92,.1);border:1px solid #E8EFFE;display:flex;align-items:center;gap:12px;">
      <div style="position:relative;flex-shrink:0;">
        <img
          src="data:image/png;base64,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"
          style="width:64px;height:64px;border-radius:50%;object-fit:cover;border:2px solid #E3F0FF;" alt="">
        <div style="position:absolute;bottom:1px;right:1px;width:12px;height:12px;border-radius:50%;background:#4CAF50;border:2px solid #fff;"></div>
      </div>
      <div style="flex:1;min-width:0;">
        <div style="display:flex;align-items:center;gap:6px;margin-bottom:3px;">
          <span style="font-size:13px;font-weight:700;color:#102A55;">专属管家</span>
          <span style="font-size:10px;background:#E3F0FF;color:#0D47A1;padding:1px 6px;border-radius:10px;font-weight:600;">一对一服务</span>
        </div>
        <div style="font-size:13px;font-weight:700;color:#102A55;margin-bottom:3px;">7×24小时在线 · 全程中文陪同</div>
        <div style="font-size:11px;color:#7A8799;">从咨询到完成，全程为您保驾护航</div>
        <div style="display:flex;gap:8px;margin-top:6px;">
          <span style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;"><span style="font-size:12px;">⚡</span>快速响应</span>
          <span style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;"><span style="font-size:12px;">💰</span>价格透明</span>
          <span style="font-size:10px;color:#0D47A1;display:flex;align-items:center;gap:2px;"><span style="font-size:12px;">🔒</span>隐私保障</span>
        </div>
      </div>
      <div style="display:flex;flex-direction:column;align-items:flex-end;gap:6px;flex-shrink:0;">
        <button style="padding:10px 16px;border-radius:22px;background:#0D47A1;color:#fff;font-size:12px;font-weight:700;border:none;cursor:pointer;white-space:nowrap;box-shadow:0 2px 8px rgba(13,71,161,.3);display:flex;align-items:center;gap:4px;">立即咨询 💬</button>
        <div style="font-size:9px;color:#9AA3B5;text-align:right;">已服务 3000+ 华人用户<br>平均响应 3 分钟</div>
      </div>
    </div>

    <!-- ── 热门服务 (Top 3) ── -->
    <div style="background:#fff;margin:0 0 8px;padding:14px 16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
        <div style="display:flex;align-items:center;gap:6px;">
          <span style="font-size:15px;">🔥</span>
          <span style="font-size:15px;font-weight:700;color:#102A55;">热门服务</span>
        </div>
        <span onclick="document.getElementById('svc-all-sec').scrollIntoView({behavior:'smooth'})" style="font-size:12px;color:#0D47A1;cursor:pointer;">查看更多 ›</span>
      </div>
      <div class="svc-hot-grid" style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px;">
        <div
          v-for="s in hotServices"
          :key="s.id"
          style="background:#fff;border-radius:14px;overflow:hidden;cursor:pointer;box-shadow:0 4px 14px rgba(15,47,92,.08);"
          @click="goToDetail(s.id)">
          <div :style="{ background: s.bg, padding: '14px 14px 10px' }">
            <div style="font-size:32px;margin-bottom:6px;">{{ s.icon }}</div>
          </div>
          <div style="padding:10px 14px 12px;">
            <div style="margin-bottom:8px;">
              <div style="font-size:14px;font-weight:700;color:#102A55;margin-bottom:4px;">{{ s.name }}</div>
              <div style="font-size:11px;color:#7A8799;margin-bottom:8px;">{{ s.sub }}</div>
              <div style="margin-bottom:8px;display:flex;gap:4px;flex-wrap:wrap;">
                <span
                  v-for="tag in s.tags.slice(0, 2)"
                  :key="tag"
                  :style="{ fontSize: '10px', background: s.color, color: '#fff', padding: '2px 8px', borderRadius: '8px', fontWeight: '600', marginRight: '4px' }">
                  {{ tag }}
                </span>
              </div>
              <div style="display:flex;align-items:center;justify-content:space-between;">
                <span :style="{ fontSize: '15px', fontWeight: '800', color: s.color }">{{ s.price }}{{ s.unit }}</span>
                <span :style="{ fontSize: '11px', color: s.color, fontWeight: '600' }">了解详情 ›</span>
              </div>
            </div>
            <button
              @click.stop="placeOrder(s.id)"
              style="margin-top:8px;width:100%;padding:8px;border-radius:12px;"
              :style="{ background: s.color, color: '#fff', fontSize: '11px', fontWeight: '700', border: 'none', cursor: 'pointer' }">
              立即下单
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── 全部服务 ── -->
    <div id="svc-all-sec" style="background:#fff;margin-bottom:8px;padding:14px 16px;">
      <div style="font-size:15px;font-weight:700;color:#102A55;margin-bottom:12px;">全部服务</div>
      <div class="svc-all-grid" style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">
        <div
          v-for="s in allServices"
          :key="s.id"
          style="background:#fff;border-radius:14px;overflow:hidden;cursor:pointer;box-shadow:0 4px 14px rgba(15,47,92,.08);border:1px solid #EEF2F8;"
          @click="goToDetail(s.id)">
          <div style="padding:14px;">
            <div style="display:flex;align-items:center;gap:10px;margin-bottom:10px;">
              <div :style="{ width: '44px', height: '44px', borderRadius: '12px', background: s.bg, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '22px', flexShrink: '0' }">
                {{ s.icon }}
              </div>
              <div>
                <div style="font-size:14px;font-weight:700;color:#102A55;">{{ s.name }}</div>
                <div style="font-size:11px;color:#7A8799;margin-top:1px;">{{ s.sub }}</div>
              </div>
            </div>
            <div style="display:flex;gap:4px;flex-wrap:wrap;margin-bottom:8px;">
              <span
                v-for="tag in s.tags.slice(0, 2)"
                :key="tag"
                :style="{ fontSize: '9px', background: s.bg, color: s.color, padding: '2px 7px', borderRadius: '8px', fontWeight: '600' }">
                {{ tag }}
              </span>
            </div>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span :style="{ fontSize: '14px', fontWeight: '800', color: s.color }">{{ s.price }}{{ s.unit }}</span>
              <div :style="{ width: '24px', height: '24px', borderRadius: '50%', background: s.bg, display: 'flex', alignItems: 'center', justifyContent: 'center' }">›</div>
            </div>
            <button
              @click.stop="placeOrder(s.id)"
              style="margin-top:8px;width:100%;padding:8px;border-radius:12px;"
              :style="{ background: s.color, color: '#fff', fontSize: '11px', fontWeight: '700', border: 'none', cursor: 'pointer' }">
              立即下单
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── 高端通道 (VIP) ── -->
    <div style="margin:0 12px 14px;background:linear-gradient(135deg,#0A1628,#1A2A50);border-radius:16px;padding:16px;box-shadow:0 6px 20px rgba(10,22,40,.4);">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px;">
        <div>
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
            <span style="font-size:18px;">👑</span>
            <span style="font-size:16px;font-weight:700;color:#F6B000;">高端通道服务</span>
            <span style="font-size:10px;background:rgba(246,176,0,.15);color:#F6B000;border:1px solid rgba(246,176,0,.3);padding:1px 8px;border-radius:10px;">私密·专属·高效</span>
          </div>
          <div style="font-size:11px;color:rgba(255,255,255,.55);">解决复杂问题，一对一资源对接，全程保密</div>
        </div>
        <button
          @click="goToDetail('vip')"
          style="padding:8px 14px;border-radius:20px;background:#F6B000;color:#0A1628;font-size:11px;font-weight:700;border:none;cursor:pointer;white-space:nowrap;">
          专属咨询 🔒
        </button>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:8px;margin-bottom:12px;">
        <div @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <div style="font-size:20px;margin-bottom:5px;">📋</div>
          <div style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;">企业审批<br>加速</div>
        </div>
        <div @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <div style="font-size:20px;margin-bottom:5px;">🤝</div>
          <div style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;">本地资源<br>对接</div>
        </div>
        <div @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <div style="font-size:20px;margin-bottom:5px;">⚖️</div>
          <div style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;">商务问题<br>解决</div>
        </div>
        <div @click="goToDetail('vip')" style="background:rgba(255,255,255,.07);border:1px solid rgba(246,176,0,.2);border-radius:10px;padding:10px 6px;text-align:center;cursor:pointer;">
          <div style="font-size:20px;margin-bottom:5px;">🛡️</div>
          <div style="font-size:10px;font-weight:600;color:#F6B000;line-height:1.3;">高端定制<br>服务</div>
        </div>
      </div>
      <div style="display:flex;justify-content:space-around;border-top:1px solid rgba(255,255,255,.08);padding-top:10px;">
        <span style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
          </svg>
          专属顾问全程跟进
        </span>
        <span style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          严格保密协议
        </span>
        <span style="font-size:10px;color:rgba(255,255,255,.5);display:flex;align-items:center;gap:3px;">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,.4)" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          高效解决方案
        </span>
      </div>
    </div>

    <div style="height:8px;"></div>
  </div>
</template>
