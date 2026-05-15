<script setup>
import { ref, computed } from "vue";
import { useClientStore } from "@/store/client";

const client = useClientStore();
const orders = computed(() => client.orders || []);

// 热门服务数据
const SVCS_DATA = ref([
  {
    id: "company",
    icon: "🏢",
    color: "#1565C0",
    bg: "#E3F0FF",
    name: "公司注册",
    sub: "外资/内资 · 一站式办理",
    price: "¥3000",
    unit: "起",
    tags: ["合规", "全程代办"],
  },
  {
    id: "visa",
    icon: "🛂",
    color: "#00897B",
    bg: "#E0F2F1",
    name: "签证办理",
    sub: "旅游/商务/工作签证",
    price: "¥600",
    unit: "起",
    tags: ["加急可办", "拒签退费"],
  },
  {
    id: "bank",
    icon: "🏦",
    color: "#F57C00",
    bg: "#FFF3E0",
    name: "银行开户",
    sub: "企业账户 · 快速开通",
    price: "¥800",
    unit: "起",
    tags: ["本地银行", "高效安全"],
  },
  {
    id: "airport",
    icon: "✈️",
    color: "#1565C0",
    bg: "#E3F0FF",
    name: "机场接机",
    sub: "落地无忧 · 中文司机",
    price: "¥280",
    unit: "起",
    tags: ["24h服务", "中文"],
  },
  {
    id: "rent",
    icon: "🏠",
    color: "#7B1FA2",
    bg: "#F3E5F5",
    name: "租房找房",
    sub: "真实房源 · 精准匹配",
    price: "¥500",
    unit: "起",
    tags: ["真实房源", "1对1带看"],
  },
  {
    id: "translate",
    icon: "🗣️",
    color: "#00695C",
    bg: "#E0F2F1",
    name: "翻译陪同",
    sub: "商务/生活翻译服务",
    price: "¥200",
    unit: "/时",
    tags: ["越南语", "商务"],
  },
  {
    id: "car",
    icon: "🚗",
    color: "#E53935",
    bg: "#FFEBEE",
    name: "商务用车",
    sub: "专车服务 · 安全舒适",
    price: "¥350",
    unit: "起",
    tags: ["7座", "专人服务"],
  },
  {
    id: "medical",
    icon: "🏥",
    color: "#00897B",
    bg: "#E0F2F1",
    name: "医疗陪诊",
    sub: "就医陪同 · 专业协助",
    price: "¥500",
    unit: "/天",
    tags: ["正规机构", "翻译"],
  },
  {
    id: "vip",
    icon: "👑",
    color: "#B8860B",
    bg: "#FFF8E1",
    name: "高端通道",
    sub: "私密·专属·高效",
    price: "面议",
    unit: "",
    tags: ["VIP专属", "保密"],
  },
]);

// 攻略精选数据
const activeNewsTab = ref("全部");
const NEWS_CATS = ["全部", "签证政策", "房产投资", "生活指南", "商务资讯"];
const DAILY_NEWS = ref([
  {
    id: "n1",
    title: "越南签证最新政策解读：商务签与旅游签的本质区别",
    tag: "签证政策",
    source: "官方发布",
    date: "05-02",
    top: true,
  },
  {
    id: "n2",
    title: "胡志明市高端公寓租房避坑指南，这5个小区最受华人欢迎",
    tag: "房产投资",
    source: "管家精选",
    date: "05-01",
    top: true,
  },
  {
    id: "n3",
    title: "在越南开办外资公司的全流程解析及税务注意事项",
    tag: "商务资讯",
    source: "政策解读",
    date: "04-28",
  },
  {
    id: "n4",
    title: "初到胡志明市：交通出行与防坑全攻略",
    tag: "生活指南",
    source: "用户分享",
    date: "04-25",
  },
]);
const filteredNews = computed(() => {
  if (activeNewsTab.value === "全部") return DAILY_NEWS.value;
  return DAILY_NEWS.value.filter((n) => n.tag === activeNewsTab.value);
});

const STATUS_CONFIG = {
  pending: { i: "📋", l: "等待审核", b: "stb-bl", p: 10 },
  requirement_submitted: { i: "📝", l: "需求已提交", b: "stb-bl", p: 20 },
  processing: { i: "⚙️", l: "服务进行中", b: "stb-gd", p: 45 },
  supplementing: { i: "⚠️", l: "需要补充材料", b: "stb-rd", p: 65 },
  completed: { i: "✅", l: "服务已完成", b: "stb-gr", p: 100 },
};
const getStatus = (sk) => STATUS_CONFIG[sk] || STATUS_CONFIG.pending;
const getOrderClass = (sk) => {
  if (sk === "supplementing") return "oc-w";
  if (sk === "completed") return "oc-d";
  return "oc-a";
};

// Navigation helpers
const goPage = (page) => {
  uni.switchTab({ url: `/pages/${page}/index` });
};
const openChat = (svc) => {
  uni.navigateTo({ url: `/pages/chat/index?svc=${encodeURIComponent(svc)}` });
};
const openSearch = () => {
  uni.navigateTo({ url: "/pages/search/index" });
};
const openLocationPicker = () => {
  uni.showToast({ title: "位置选择", icon: "none" });
};
const homeNewsFilter = (cat) => {
  uni.showToast({ title: cat, icon: "none" });
};
const toast = (msg) => {
  uni.showToast({ title: msg, icon: "none" });
};
</script>

<template>
  <!-- ============================================================ -->
  <!--  HOME PAGE (yesok-final.html pg-home, 1:1 移植)          -->
  <!-- ============================================================ -->

  <!-- ── HERO BANNER ── -->
  <view class="container" style="padding-top: 0; margin-top: 0; overflow-x: hidden;">

    <view class="hero" style="background-image: url('/static/img.png'); background-size: cover; background-position: center top; width: 100vw; height: 320px; padding-top: 60px; margin-top: -20px; border-radius: 0 0 24px 24px;">
      <view class="tb-main" style="padding: 0 20px; display: flex; justify-content: space-between; align-items: center;">
        <view class="logo" style="font-size: 22px; font-weight: 800; color: #fff;">Yesok <text style="color:#F6B000;">Vietnam</text></view>
        <view class="t-ico" style="background:#fff; border-radius:50%; width:32px; height:32px; display:flex; align-items:center; justify-content:center; box-shadow:var(--sh);">🇻🇳</view>
      </view>
    </view>

    <view style="padding: 0 20px; margin-top: -30px; position: relative; z-index: 10;">
      <view class="search-bar" style="background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 12px; padding: 12px 16px; display: flex; align-items: center; gap: 10px; box-shadow: 0 4px 20px rgba(13, 71, 161, 0.14);">
        <text>🔍</text>
        <input type="text" placeholder="搜索越南管家服务..." style="flex: 1; font-size: 14px; border: none; outline: none; background: transparent;" />
      </view>
    </view>

  </view>

  <!-- ── SERVICE CATEGORIES ── -->
  <view
    style="
      background: #fff;
      padding: 4px 0 10px;
      margin-bottom: 14px;
      box-shadow: 0 1px 8px rgba(0, 0, 0, 0.04);
    "
  >
    <view
      style="display: flex; justify-content: space-around; padding: 8px 8px 0"
    >
      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
          "
          >✈️</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >签证旅游</view
        >
      </view>

      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
          "
          >🏢</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >房产投资</view
        >
      </view>

      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
          "
          >📋</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >公司注册</view
        >
      </view>

      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
          "
          >🚗</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >接送出行</view
        >
      </view>

      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
          "
          >🗣️</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >翻译服务</view
        >
      </view>

      <view
        @click="goPage('services')"
        style="
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          cursor: pointer;
          width: 54px;
        "
      >
        <view
          style="
            width: 48px;
            height: 48px;
            border-radius: 50%;
            background: #f0f6ff;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            color: #9aa3b5;
            font-weight: 700;
          "
          >···</view
        >
        <view
          style="
            font-size: 11px;
            color: #1a2340;
            font-weight: 500;
            text-align: center;
          "
          >更多服务</view
        >
      </view>
    </view>
  </view>

  <!-- ── HOT SERVICES ── -->
  <!-- 横向滚动服务卡片，完整数据绑定 + hc-btn 去咨询按钮 -->
  <view
    style="
      background: #fff;
      margin-bottom: 14px;
      padding: 14px 16px 4px;
      box-shadow: 0 1px 8px rgba(0, 0, 0, 0.04);
    "
  >
    <view
      style="
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 12px;
      "
    >
      <view style="font-size: 16px; font-weight: 700; color: #1a2340"
        >热门服务</view
      >
      <view
        @click="goPage('services')"
        style="
          font-size: 12px;
          color: #9aa3b5;
          cursor: pointer;
          display: flex;
          align-items: center;
          gap: 2px;
        "
      >
        全部服务
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#9AA3B5"
          stroke-width="2"
        >
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </view>
    </view>

    <!-- 横向滚动卡片列表 -->
    <scroll-view
      class="hot-scroll"
      scroll-x
      style="white-space: nowrap; padding: 10px 16px"
    >
      <view
        v-for="svc in SVCS_DATA"
        :key="svc.id"
        class="hc-card"
        style="
          display: inline-block;
          width: 280px;
          margin-right: 12px;
          background: #fff;
          border-radius: 16px;
          padding: 16px;
          box-shadow: var(--sh);
          vertical-align: top;
        "
      >
        <view style="display: flex; gap: 12px">
          <!-- 左侧图标 -->
          <view
            class="hc-icon"
            :style="{ background: svc.bg }"
            style="
              width: 48px;
              height: 48px;
              border-radius: 12px;
              display: flex;
              align-items: center;
              justify-content: center;
              font-size: 24px;
              flex-shrink: 0;
            "
          >
            {{ svc.icon }}
          </view>

          <!-- 右侧信息 -->
          <view class="hc-info" style="flex: 1; min-width: 0">
            <view
              class="hc-name"
              :style="{ color: svc.color }"
              style="font-size: 15px; font-weight: 700; margin-bottom: 4px"
            >
              {{ svc.name }}
            </view>
            <view
              class="hc-sub"
              style="font-size: 12px; color: var(--tx2); margin-bottom: 8px"
            >
              {{ svc.sub }}
            </view>

            <!-- 标签 -->
            <view
              class="hc-tags"
              style="display: flex; gap: 6px; margin-bottom: 12px"
            >
              <text
                v-for="tag in svc.tags"
                :key="tag"
                class="hc-tag"
                style="
                  padding: 2px 6px;
                  background: var(--gy);
                  border-radius: 4px;
                  font-size: 10px;
                  color: var(--tx3);
                "
              >
                {{ tag }}
              </text>
            </view>

            <!-- 底部：价格 + 去咨询按钮 -->
            <view
              class="hc-bot"
              style="
                display: flex;
                justify-content: space-between;
                align-items: center;
                border-top: 1px dashed #eee;
                padding-top: 10px;
              "
            >
              <view
                class="hc-price"
                style="font-size: 16px; font-weight: 800; color: var(--tx)"
              >
                {{ svc.price
                }}<text
                  style="font-size: 10px; font-weight: 400; color: var(--tx3)"
                  >/{{ svc.unit }}</text
                >
              </view>
              <button
                class="hc-btn"
                :style="{ background: svc.color }"
                style="
                  margin: 0;
                  padding: 0 16px;
                  height: 28px;
                  line-height: 28px;
                  border-radius: 14px;
                  color: #fff;
                  font-size: 12px;
                  font-weight: 600;
                  border: none;
                "
                @click="openChat(svc.name)"
              >
                去咨询
              </button>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>

    <view style="padding: 24px 16px 12px; display: flex; justify-content: space-between; align-items: center;">
      <view style="font-size: 18px; font-weight: 800; color: var(--tx);">攻略精选</view>
      <view style="font-size: 12px; color: var(--tx3);">了解真实越南 ></view>
    </view>
    <scroll-view scroll-x class="nc-scroll" style="white-space: nowrap; padding: 0 16px 12px;">
      <view style="display: inline-flex; gap: 8px;">
        <text v-for="cat in NEWS_CATS" :key="cat"
              :class="['nc-btn', activeNewsTab === cat ? 'active' : '']"
              @click="activeNewsTab = cat"
              style="padding: 6px 14px; border-radius: 16px; font-size: 12px; font-weight: 600; cursor: pointer; transition: all 0.3s; background: var(--gy); color: var(--tx2);">
          {{ cat }}
        </text>
      </view>
    </scroll-view>
    <view style="padding: 0 16px; margin-bottom: 24px;">
      <view style="background: #fff; border-radius: 16px; box-shadow: var(--sh); overflow: hidden;">
        <view v-for="n in filteredNews" :key="n.id" class="sr-item" style="display: flex; gap: 12px; padding: 12px 16px; border-bottom: 1px solid var(--gy);">
          <view class="sr-type news" style="background: var(--gdl); width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px;">📰</view>
          <view style="flex: 1; min-width: 0;">
            <view class="sr-name" style="font-size: 13px; font-weight: 600; color: var(--tx); line-height: 1.4; margin-bottom: 4px;">
              <text v-if="n.top" style="color: var(--rd);">📌 </text>{{ n.title }}
            </view>
            <view class="sr-sub" style="font-size: 11px; color: var(--tx3);">{{ n.tag }} · {{ n.source }}</view>
          </view>
          <view class="sr-badge" style="background: var(--gy2); color: var(--tx3); margin-left: auto; padding: 2px 8px; border-radius: 10px; font-size: 10px; font-weight: 600; height: fit-content;">{{ n.date }}</view>
        </view>
      </view>
    </view>

    <view style="margin: 0 16px 20px; padding: 20px; background: #fff; border-radius: 16px; box-shadow: var(--sh);">
      <view style="font-size: 15px; font-weight: 700; margin-bottom: 16px; text-align: center; color: var(--tx);">为什么选择 Yesok？</view>
      <view style="display: flex; justify-content: space-around;">
        <view style="text-align: center;">
          <view style="font-size: 24px; margin-bottom: 8px;">🥇</view>
          <view style="font-size: 12px; font-weight: 600; color: var(--tx);">官方直营</view>
        </view>
        <view style="text-align: center;">
          <view style="font-size: 24px; margin-bottom: 8px;">⚡</view>
          <view style="font-size: 12px; font-weight: 600; color: var(--tx);">极速响应</view>
        </view>
        <view style="text-align: center;">
          <view style="font-size: 24px; margin-bottom: 8px;">🛡️</view>
          <view style="font-size: 12px; font-weight: 600; color: var(--tx);">资金担保</view>
        </view>
        <view style="text-align: center;">
          <view style="font-size: 24px; margin-bottom: 8px;">🇨🇳</view>
          <view style="font-size: 12px; font-weight: 600; color: var(--tx);">全中文服务</view>
        </view>
      </view>
    </view>

  <!-- ── 订单动态 ── -->
  <view style="border-top: 1px solid var(--br); padding-top: 4px">
    <view style="padding: 14px 16px 10px">
      <view style="font-size: 16px; font-weight: 700; color: #1a2340"
        >📋 订单动态</view
      >
    </view>

    <view class="ord-list">
      <template v-if="orders.length === 0">
        <view class="empty">
          <view class="empty-ic">📭</view>
          <view class="empty-ti">暂无订单</view>
        </view>
      </template>
      <template v-else>
        <view
          v-for="order in orders"
          :key="order.id"
          :class="['ord-card', getOrderClass(order.sk)]"
        >
          <view class="oc-top">
            <view>
              <view class="oc-svc"
                >{{ order.icon || "📋" }}
                {{ order.serviceName || order.svc }}</view
              >
              <view class="oc-no">订单号 {{ order.id }}</view>
            </view>
            <text :class="['stb', getStatus(order.status || order.sk).b]">
              {{ getStatus(order.status || order.sk).l }}
            </text>
          </view>
          <view style="margin: 9px 0">
            <view class="oc-pbar">
              <view
                class="oc-pfill"
                :style="{ width: getStatus(order.status || order.sk).p + '%' }"
              ></view>
            </view>
            <view class="oc-plbl">
              <text>{{ getStatus(order.status || order.sk).l }}</text>
              <text>{{ getStatus(order.status || order.sk).p }}% 完成</text>
            </view>
          </view>
          <view v-if="order.status === 'supplementing'" class="oc-warn">
            <view class="oc-wdot"></view>
            需要操作：请补充材料
          </view>
          <view class="oc-bot">
            <view class="oc-mgr">
              <view class="oc-mav">{{
                (order.managerName || order.mg || "管")[0]
              }}</view>
              {{ order.managerName || order.mg || "专属管家" }}
            </view>
            <view style="font-size: 11px; color: var(--tx3)">{{
              order.price || order.pr || "—"
            }}</view>
          </view>
        </view>
      </template>
    </view>
  </view>

  <view style="height: 12px"></view>
</template>

<style scoped>
/* 全局样式已由 style.css 接管，此处不留任何样式 */
</style>
