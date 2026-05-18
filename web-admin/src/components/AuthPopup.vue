<script setup>
import { computed } from 'vue'
import { useClientStore } from '@/store/client'
import { ElMessage } from 'element-plus'

const client = useClientStore()
const visible = computed(() => client.loginSheetVisible)

const showToast = (title, type = 'info') => {
  ElMessage({ message: title, type })
}

const handleDemoLogin = async () => {
  try {
    return await client.loginByDemo()
  } catch (error) {
    showToast('登录失败，请稍后重试', 'error')
    return null
  }
}

const handleWechatAuthorize = async () => handleDemoLogin()

const handleTelegramPlaceholder = () => {
  showToast('TG 登录占位已保留，当前请用演示登录', 'info')
}
</script>

<template>
  <div v-if="visible" class="auth-mask" @click="client.closeLoginSheet">
    <div class="auth-popup" @click.stop>
      <div class="auth-handle"></div>
      <div class="auth-vip-mark">YESOK PASSPORT</div>
      <div class="auth-title">登录后{{ client.pendingActionText }}</div>
      <div class="auth-desc">
        Yesok 将为你建立专属越南管家档案，用于订单进度、材料提醒与节点验收。当前演示版仅使用 Mock 数据，不会向真实后端提交个人信息。
      </div>
      <button class="auth-primary" @click="handleDemoLogin">H5 一键演示登录</button>
      <button class="auth-secondary" @click="handleTelegramPlaceholder">Telegram Mini App 登录占位</button>
      <div class="auth-tips">
        <span>已隔离微信小程序、H5、Telegram Mini App 与未来 iOS/Android 登录入口。</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-mask {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 9999;
  display: flex;
  align-items: flex-end;
  background: rgba(0, 31, 27, 0.46);
  backdrop-filter: blur(8px);
}

.auth-popup {
  width: 100%;
  padding: 12px 24px calc(28px + env(safe-area-inset-bottom));
  border: 1px solid rgba(197, 160, 89, 0.2);
  border-radius: 32px 32px 0 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), #f2f6f5 100%);
  box-shadow: 0 -26px 70px rgba(0, 77, 64, 0.2);
  animation: springSlideUp 560ms cubic-bezier(0.18, 1.34, 0.42, 1) both;
}

.auth-handle {
  width: 46px;
  height: 5px;
  margin: 0 auto 18px;
  border-radius: 99px;
  background: rgba(0, 77, 64, 0.18);
}

.auth-vip-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10px;
  padding: 5px 12px;
  border-radius: 999px;
  color: #7a5a21;
  background: rgba(197, 160, 89, 0.16);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 1.2px;
}

.auth-title {
  color: #12312c;
  font-size: 21px;
  font-weight: 900;
  line-height: 1.35;
  text-align: center;
}

.auth-desc {
  margin: 12px 0 20px;
  color: #6b7c78;
  font-size: 13px;
  line-height: 1.75;
  text-align: center;
}

.auth-primary,
.auth-secondary {
  display: block;
  width: 100%;
  height: 48px;
  margin: 0 0 12px;
  border: none;
  border-radius: 24px;
  font-size: 15px;
  font-weight: 800;
  line-height: 48px;
  cursor: pointer;
}

.auth-primary {
  color: #fff;
  background: linear-gradient(135deg, #004d40, #00695c);
  box-shadow: 0 14px 30px rgba(0, 77, 64, 0.22);
}

.auth-secondary {
  color: #004d40;
  background: rgba(0, 77, 64, 0.08);
}

.auth-tips {
  color: rgba(18, 49, 44, 0.5);
  font-size: 11px;
  line-height: 1.65;
  text-align: center;
}

@keyframes springSlideUp {
  0% { opacity: 0; transform: translateY(110%); }
  68% { opacity: 1; transform: translateY(-10px); }
  84% { transform: translateY(4px); }
  100% { opacity: 1; transform: translateY(0); }
}
</style>
