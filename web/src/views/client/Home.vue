<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useClientStore } from '../../store/client'

const router = useRouter()
const client = useClientStore()
const orders = computed(() => client.orders || [])

const STATUS_CONFIG = {
  pending:      { i: '📋', l: '等待审核',    b: 'stb-bl', p: 10 },
  requirement_submitted: { i: '📝', l: '需求已提交', b: 'stb-bl', p: 20 },
  processing:   { i: '⚙️', l: '服务进行中',  b: 'stb-gd', p: 45 },
  supplementing:{ i: '⚠️', l: '需要补充材料', b: 'stb-rd', p: 65 },
  completed:    { i: '✅', l: '服务已完成',  b: 'stb-gr', p: 100 },
}
const getStatus = (sk) => STATUS_CONFIG[sk] || STATUS_CONFIG.pending
const getOrderClass = (sk) => {
  if (sk === 'supplementing') return 'oc-w'
  if (sk === 'completed') return 'oc-d'
  return 'oc-a'
}
</script>

<template>
  <!-- HERO BANNER: 427:357 aspect ratio, flex layout -->
  <div class="hero" style="background-image: url('/img.png'); background-size: cover; background-position: center; aspect-ratio: 427/357; min-height: 350px; display: flex; flex-direction: column; justify-content: space-between; border: none;">
    <!-- Empty top area — image carries the branding text -->
    <!-- Bottom overlay: search bar -->
    <div style="padding:0 16px 16px;">
      <div style="background:#fff;border-radius:22px;padding:9px 12px;display:flex;align-items:center;gap:8px;box-shadow:0 2px 12px rgba(0,0,0,.15);cursor:pointer;" @click="router.push('/service')">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9AA3B5" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <span style="flex:1;font-size:12px;color:#9AA3B5;">搜索签证、接机、公司注册…</span>
        <button style="padding:6px 12px;border-radius:18px;background:var(--bl);color:#fff;font-size:11px;font-weight:700;border:none;cursor:pointer;">搜索</button>
      </div>
    </div>
  </div>

  <!-- Quick Categories -->
  <div style="background:#fff;padding:14px 16px 4px;">
    <div style="display:flex;justify-content:space-between;gap:4px;">
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:22px;">✈️</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">签证旅游</div>
      </div>
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:22px;">🏢</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">房产投资</div>
      </div>
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:22px;">📋</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">公司注册</div>
      </div>
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:22px;">🚗</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">接送出行</div>
      </div>
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:22px;">🗣️</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">翻译服务</div>
      </div>
      <div @click="router.push('/service')" style="display:flex;flex-direction:column;align-items:center;gap:5px;cursor:pointer;width:54px;">
        <div style="width:48px;height:48px;border-radius:50%;background:#F0F6FF;display:flex;align-items:center;justify-content:center;font-size:20px;color:#9AA3B5;font-weight:700;">···</div>
        <div style="font-size:11px;color:#1A2340;font-weight:500;text-align:center;">更多服务</div>
      </div>
    </div>
  </div>

  <!-- HOT SERVICES -->
  <div style="background:#fff;margin-bottom:14px;padding:14px 16px;box-shadow:0 1px 8px rgba(0,0,0,.04);">
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
      <div style="font-size:16px;font-weight:700;color:#1A2340;">热门服务</div>
      <div @click="router.push('/service')" style="font-size:12px;color:#9AA3B5;cursor:pointer;display:flex;align-items:center;gap:2px;">全部服务 <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#9AA3B5" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg></div>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px;">
      <div @click="router.push('/service')" style="border-radius:12px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 2px 8px rgba(13,71,161,.08);">
        <div style="height:72px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);position:relative;overflow:hidden;display:flex;align-items:flex-end;justify-content:flex-start;padding:8px 8px 6px;">
          <div style="position:absolute;right:-8px;top:-8px;font-size:52px;opacity:.4;">🏙️</div>
          <div style="font-size:13px;font-weight:700;color:#1A2340;line-height:1.3;position:relative;z-index:1;">签证代办<br><span style="font-size:10px;font-weight:500;color:#4A5568;">快速出签</span></div>
        </div>
        <div style="padding:7px 8px;background:#fff;">
          <button style="width:100%;padding:5px;border-radius:16px;background:var(--bl);color:#fff;font-size:11px;font-weight:600;border:none;cursor:pointer;">去咨询</button>
        </div>
      </div>
      <div @click="router.push('/service')" style="border-radius:12px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 2px 8px rgba(13,71,161,.08);">
        <div style="height:72px;background:linear-gradient(135deg,#E8F5E9,#C8E6C9);position:relative;overflow:hidden;display:flex;align-items:flex-end;padding:8px 8px 6px;">
          <div style="position:absolute;right:-8px;top:-4px;font-size:52px;opacity:.4;">🚗</div>
          <div style="font-size:13px;font-weight:700;color:#1A2340;line-height:1.3;position:relative;z-index:1;">机场接机<br><span style="font-size:10px;font-weight:500;color:#4A5568;">安全 · 准时</span></div>
        </div>
        <div style="padding:7px 8px;background:#fff;">
          <button style="width:100%;padding:5px;border-radius:16px;background:var(--bl);color:#fff;font-size:11px;font-weight:600;border:none;cursor:pointer;">去咨询</button>
        </div>
      </div>
      <div @click="router.push('/service')" style="border-radius:12px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 2px 8px rgba(13,71,161,.08);">
        <div style="height:72px;background:linear-gradient(135deg,#FFF8E1,#FFE082);position:relative;overflow:hidden;display:flex;align-items:flex-end;padding:8px 8px 6px;">
          <div style="position:absolute;right:-8px;top:-4px;font-size:52px;opacity:.35;">🏛️</div>
          <div style="font-size:13px;font-weight:700;color:#1A2340;line-height:1.3;position:relative;z-index:1;">公司注册<br><span style="font-size:10px;font-weight:500;color:#4A5568;">快速办理</span></div>
        </div>
        <div style="padding:7px 8px;background:#fff;">
          <button style="width:100%;padding:5px;border-radius:16px;background:var(--bl);color:#fff;font-size:11px;font-weight:600;border:none;cursor:pointer;">去咨询</button>
        </div>
      </div>
    </div>
  </div>

  <!-- 攻略精选 (资讯) -->
  <div style="background:#fff;margin-bottom:14px;padding:14px 16px;box-shadow:0 1px 8px rgba(0,0,0,.04);">
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
      <div style="font-size:16px;font-weight:700;color:#1A2340;">攻略精选</div>
      <div @click="router.push('/news')" style="font-size:12px;color:#9AA3B5;cursor:pointer;display:flex;align-items:center;gap:2px;">更多 <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#9AA3B5" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg></div>
    </div>
    <div style="display:flex;gap:6px;margin-bottom:12px;overflow-x:auto;scrollbar-width:none;">
      <div style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:var(--bl);color:#fff;font-size:12px;font-weight:600;cursor:pointer;">全部</div>
      <div style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:#F0F6FF;color:#4A5568;font-size:12px;cursor:pointer;">签证</div>
      <div style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:#F0F6FF;color:#4A5568;font-size:12px;cursor:pointer;">生活</div>
      <div style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:#F0F6FF;color:#4A5568;font-size:12px;cursor:pointer;">商务</div>
      <div style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:#F0F6FF;color:#4A5568;font-size:12px;cursor:pointer;">公司</div>
      <div @click="router.push('/news')" style="flex-shrink:0;padding:5px 16px;border-radius:20px;background:#F0F6FF;color:#9AA3B5;font-size:12px;cursor:pointer;">更多 ›</div>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px;">
      <div @click="router.push('/news')" style="border-radius:10px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 1px 6px rgba(0,0,0,.06);">
        <div style="height:90px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);display:flex;align-items:center;justify-content:center;font-size:44px;">🇻🇳</div>
        <div style="padding:7px 8px;background:#fff;">
          <div style="font-size:11px;font-weight:600;color:#1A2340;line-height:1.4;margin-bottom:4px;">2026越南签证最新政策</div>
          <div style="font-size:9px;color:#9AA3B5;">👁 1.2k · ⭐ 4.9</div>
        </div>
      </div>
      <div @click="router.push('/news')" style="border-radius:10px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 1px 6px rgba(0,0,0,.06);">
        <div style="height:90px;background:linear-gradient(135deg,#FFF8E1,#FFE082);display:flex;align-items:center;justify-content:center;font-size:44px;">🏢</div>
        <div style="padding:7px 8px;background:#fff;">
          <div style="font-size:11px;font-weight:600;color:#1A2340;line-height:1.4;margin-bottom:4px;">外资公司注册避坑指南</div>
          <div style="font-size:9px;color:#9AA3B5;">👁 856 · ⭐ 4.8</div>
        </div>
      </div>
      <div @click="router.push('/news')" style="border-radius:10px;overflow:hidden;border:1px solid #E2E8F4;cursor:pointer;box-shadow:0 1px 6px rgba(0,0,0,.06);">
        <div style="height:90px;background:linear-gradient(135deg,#E8F5E9,#C8E6C9);display:flex;align-items:center;justify-content:center;font-size:44px;">🏙️</div>
        <div style="padding:7px 8px;background:#fff;">
          <div style="font-size:11px;font-weight:600;color:#1A2340;line-height:1.4;margin-bottom:4px;">越南生活成本全解析</div>
          <div style="font-size:9px;color:#9AA3B5;">👁 620 · ⭐ 4.7</div>
        </div>
      </div>
    </div>
  </div>

  <!-- PLATFORM GUARANTEE -->
  <div style="background:#fff;margin-bottom:14px;padding:14px 16px;box-shadow:0 1px 8px rgba(0,0,0,.04);">
    <div style="font-size:16px;font-weight:700;color:#1A2340;margin-bottom:14px;">平台保障</div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr 1fr;gap:6px;">
      <div style="text-align:center;">
        <div style="width:40px;height:40px;border-radius:50%;background:#F0F6FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--bl)" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        </div>
        <div style="font-size:11px;font-weight:600;color:#1A2340;">专业服务</div>
        <div style="font-size:9px;color:#9AA3B5;margin-top:2px;">专属管家团队</div>
      </div>
      <div style="text-align:center;">
        <div style="width:40px;height:40px;border-radius:50%;background:#F0F6FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--bl)" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        </div>
        <div style="font-size:11px;font-weight:600;color:#1A2340;">安全可靠</div>
        <div style="font-size:9px;color:#9AA3B5;margin-top:2px;">信息保密保障</div>
      </div>
      <div style="text-align:center;">
        <div style="width:40px;height:40px;border-radius:50%;background:#F0F6FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--bl)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        </div>
        <div style="font-size:11px;font-weight:600;color:#1A2340;">价格透明</div>
        <div style="font-size:9px;color:#9AA3B5;margin-top:2px;">无隐形消费</div>
      </div>
      <div style="text-align:center;">
        <div style="width:40px;height:40px;border-radius:50%;background:#F0F6FF;margin:0 auto 6px;display:flex;align-items:center;justify-content:center;">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--bl)" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div style="font-size:11px;font-weight:600;color:#1A2340;">24小时服务</div>
        <div style="font-size:9px;color:#9AA3B5;margin-top:2px;">随时为您服务</div>
      </div>
    </div>
  </div>

  <!-- Orders section -->
  <div style="border-top:1px solid var(--br);padding-top:4px;">
    <div style="padding:14px 16px 10px;">
      <div style="font-size:16px;font-weight:700;color:#1A2340;">📋 订单动态</div>
    </div>
    <div class="ord-list">
      <template v-if="orders.length === 0">
        <div class="empty">
          <div class="empty-ic">📭</div>
          <div class="empty-ti">暂无订单</div>
        </div>
      </template>
      <template v-else>
        <div
          v-for="order in orders"
          :key="order.id"
          :class="['ord-card', getOrderClass(order.sk)]"
        >
          <div class="oc-top">
            <div>
              <div class="oc-svc">{{ order.icon || '📋' }} {{ order.serviceName || order.svc }}</div>
              <div class="oc-no">订单号 {{ order.id }}</div>
            </div>
            <span :class="['stb', getStatus(order.status || order.sk).b]">
              {{ getStatus(order.status || order.sk).l }}
            </span>
          </div>
          <div style="margin:9px 0;">
            <div class="oc-pbar">
              <div class="oc-pfill" :style="{ width: getStatus(order.status || order.sk).p + '%' }"></div>
            </div>
            <div class="oc-plbl">
              <span>{{ getStatus(order.status || order.sk).l }}</span>
              <span>{{ getStatus(order.status || order.sk).p }}% 完成</span>
            </div>
          </div>
          <div v-if="order.status === 'supplementing'" class="oc-warn">
            <div class="oc-wdot"></div>
            需要操作：请补充材料
          </div>
          <div class="oc-bot">
            <div class="oc-mgr">
              <div class="oc-mav">{{ (order.managerName || order.mg || '管')[0] }}</div>
              {{ order.managerName || order.mg || '专属管家' }}
            </div>
            <div style="font-size:11px;color:var(--tx3);">{{ order.price || order.pr || '—' }}</div>
          </div>
        </div>
      </template>
    </div>
  </div>

  <div style="height:12px;"></div>
</template>

<style scoped>
/* 全局样式已由 style.css 接管，此处不留任何样式 */
</style>
