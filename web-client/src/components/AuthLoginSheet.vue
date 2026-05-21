<script setup>
import { computed } from 'vue'
import { useClientStore } from '@/store/client'

const client = useClientStore()
const visible = computed(() => client.loginSheetVisible)

// handleDemoLogin 执行演示登录动作。
// 实现步骤：
// 1. 调用客户端 store 的 Mock 登录方法。
// 2. 由 store 统一写入 token、用户资料和本地缓存。
// 3. 登录失败时展示轻提示，避免用户停留在无反馈状态。
const handleDemoLogin = async () => {
  try {
    await client.loginByDemo()
  } catch (error) {
    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.showToast) {
      uniApi.showToast({ title: '登录失败，请稍后重试', icon: 'none' })
    }
  }
}

// handleTelegramPlaceholder 预留 Telegram Mini App 登录入口。
// 实现步骤：
// 1. 当前不联调真实 Telegram 无感登录，避免误信任前端 initData。
// 2. 保留按钮与注释，日后接入时在 api/client/auth.js 中补充真实校验。
// 3. 临时引导用户使用演示登录继续验收页面。
const handleTelegramPlaceholder = () => {
  const uniApi = typeof uni !== 'undefined' ? uni : null
  if (uniApi?.showToast) {
    uniApi.showToast({ title: 'TG 登录占位已保留，当前请用演示登录', icon: 'none' })
  }
}
</script>

<template>
  <view v-if="visible" class="auth-mask" @click="client.closeLoginSheet">
    <view class="auth-sheet" @click.stop>
      <view class="auth-handle"></view>
      <view class="auth-title">登录后{{ client.pendingActionText }}</view>
      <view class="auth-desc">
        为保障服务履约与订单进度通知，请先完成授权登录。当前演示版使用 Mock 数据，不会向真实后端提交个人信息。
      </view>

      <button class="auth-primary" @click="handleDemoLogin">一键演示登录</button>
      <button class="auth-secondary" @click="handleTelegramPlaceholder">Telegram 登录占位</button>

      <view class="auth-tips">
        <text>已预留微信小程序、Telegram Mini App 与未来 iOS/Android 原生登录扩展位。</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.auth-mask {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  z-index: 9999;
  display: flex;
  align-items: flex-end;
  background: rgba(13, 24, 54, 0.42);
}

.auth-sheet {
  width: 100%;
  padding: 10px 22px calc(24px + env(safe-area-inset-bottom));
  border-radius: 24px 24px 0 0;
  background: #fff;
  box-shadow: 0 -12px 36px rgba(13, 71, 161, 0.18);
}

.auth-handle {
  width: 44px;
  height: 4px;
  margin: 0 auto 18px;
  border-radius: 99px;
  background: #d9e2ef;
}

.auth-title {
  color: #16213e;
  font-size: 20px;
  font-weight: 800;
  text-align: center;
}

.auth-desc {
  margin: 12px 0 18px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.7;
  text-align: center;
}

.auth-primary,
.auth-secondary {
  height: 46px;
  margin: 0 0 10px;
  border: none;
  border-radius: 16px;
  font-size: 15px;
  font-weight: 700;
  line-height: 46px;
}

.auth-primary {
  color: #fff;
  background: linear-gradient(135deg, #0d47a1, #1976d2);
}

.auth-secondary {
  color: #0d47a1;
  background: #eef6ff;
}

.auth-tips {
  color: #98a2b3;
  font-size: 11px;
  line-height: 1.6;
  text-align: center;
}
</style>
