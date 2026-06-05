<script setup>
import { computed } from 'vue'
import { ORIGIN_URL } from '@/api/request'
import { useClientStore } from '@/store/client'

const client = useClientStore()
const visible = computed(() => client.loginSheetVisible)
const profileVisible = computed(() => client.profileSheetVisible)
const profileForm = computed(() => client.profileForm || {})
const toAvatarUrl = (url) => {
  if (!url) return ''
  const value = String(url || '').trim()
  if (!value) return ''

  if (
    value.startsWith('wxfile://') ||
    value.startsWith('file://') ||
    value.startsWith('http://tmp/') ||
    value.includes('/tmp/') ||
    value.includes('/temp/')
  ) {
    return value
  }

  if (/^https?:\/\//.test(value)) return value

  if (value.startsWith('/static/') || value.startsWith('static/')) {
    return value.startsWith('/') ? value : `/${value}`
  }

  if (
    value.startsWith('/uploads/') ||
    value.startsWith('uploads/') ||
    value.startsWith('/material/') ||
    value.startsWith('material/')
  ) {
    const path = value.startsWith('/') ? value : `/${value}`
    return `${ORIGIN_URL}${path}`
  }

  return `${ORIGIN_URL}${value.startsWith('/') ? '' : '/'}${value}`
}
const profileAvatarPreview = computed(() => toAvatarUrl(profileForm.value.avatar_url))

const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({ title, icon: 'none' })
    return
  }
  console.info('[Yesok Auth]', title)
}

const handleDemoLogin = async () => {
  try {
    return await client.loginByDemo()
  } catch (error) {
    showSafeToast('登录失败，请稍后重试')
    return null
  }
}

const handleWechatAuthorize = async () => {
  try {
    return await client.loginByWechat('')
  } catch (error) {
    console.error('[WechatLogin] failed:', error)
    showSafeToast(error?.message || '微信登录失败，请稍后重试')
    return null
  }
}

const handleTelegramPlaceholder = () => {
  showSafeToast('TG 登录占位已保留，当前请用演示登录')
}

const handleChooseAvatar = (event) => {
  const avatarUrl = event?.detail?.avatarUrl || ''
  if (!avatarUrl) return
  client.setProfileForm({ avatar_url: avatarUrl })
}

const handleNicknameInput = (event) => {
  const value = event?.detail?.value || ''
  client.setProfileForm({ nickname: value })
}

const handleSaveProfile = async () => {
  try {
    await client.completeProfile(profileForm.value)
  } catch (error) {
    console.error('[ProfileComplete] failed:', error)
    showSafeToast(error?.message || '资料保存失败，请稍后重试')
  }
}
</script>

<template>
  <view>
    <view v-if="visible" class="auth-mask" @click="client.closeLoginSheet">
      <view class="auth-popup" @click.stop>
        <view class="auth-handle"></view>
        <view class="auth-vip-mark">YESOK PASSPORT</view>
        <view class="auth-title">登录后{{ client.pendingActionText }}</view>
        <view class="auth-desc">
          Yesok 将为您建立专属越南管家档案，用于订单进度、材料提醒与节点验收。
        </view>

        <!-- #ifdef MP-WEIXIN -->
        <button class="auth-primary" @click="handleWechatAuthorize">
          微信授权登录
        </button>
        <!-- #endif -->

        <!-- #ifdef H5 -->
        <button class="auth-primary" @click="handleDemoLogin">H5 一键演示登录</button>
        <button class="auth-secondary" @click="handleTelegramPlaceholder">Telegram Mini App 登录占位</button>
        <!-- #endif -->

        <!-- #ifndef MP-WEIXIN -->
        <!-- #ifndef H5 -->
        <button class="auth-primary" @click="handleDemoLogin">一键演示登录</button>
        <!-- #endif -->
        <!-- #endif -->
      </view>
    </view>

    <view v-if="profileVisible" class="auth-mask" @click="client.closeProfileSheet">
      <view class="auth-popup" @click.stop>
        <view class="auth-handle"></view>
        <view class="auth-vip-mark">PROFILE</view>
        <view class="auth-title">完善头像昵称</view>
        <view class="auth-desc">
          用于订单联系人展示和管家服务识别，可随时在个人中心修改。
        </view>

        <!-- #ifdef MP-WEIXIN -->
        <button class="avatar-picker" open-type="chooseAvatar" @chooseavatar="handleChooseAvatar">
          <image
            v-if="profileAvatarPreview"
            class="avatar-preview"
            :src="profileAvatarPreview"
            mode="aspectFill"
          />
          <text v-else>选择头像</text>
        </button>

        <input
          class="nickname-input"
          type="nickname"
          :value="profileForm.nickname"
          placeholder="请输入昵称"
          @input="handleNicknameInput"
          @blur="handleNicknameInput"
        />
        <!-- #endif -->

        <!-- #ifndef MP-WEIXIN -->
        <input
          class="nickname-input"
          :value="profileForm.nickname"
          placeholder="请输入昵称"
          @input="handleNicknameInput"
          @blur="handleNicknameInput"
        />
        <!-- #endif -->

        <button class="auth-primary" @click="handleSaveProfile">保存资料</button>
        <button class="auth-secondary" @click="client.closeProfileSheet">稍后再说</button>
      </view>
    </view>
  </view>
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
  height: 48px;
  margin: 0 0 12px;
  border: none;
  border-radius: 24px;
  font-size: 15px;
  font-weight: 800;
  line-height: 48px;
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

.avatar-picker {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 88px;
  height: 88px;
  margin: 0 auto 16px;
  padding: 0;
  border: none;
  border-radius: 50%;
  overflow: hidden;
  background: rgba(0, 77, 64, 0.08);
  color: #004d40;
  font-size: 13px;
  font-weight: 800;
  line-height: 88px;
}

.avatar-preview {
  width: 88px;
  height: 88px;
  border-radius: 50%;
}

.nickname-input {
  height: 48px;
  margin-bottom: 14px;
  padding: 0 16px;
  border-radius: 24px;
  background: rgba(0, 77, 64, 0.06);
  color: #12312c;
  font-size: 15px;
  line-height: 48px;
}

.auth-tips {
  color: rgba(18, 49, 44, 0.5);
  font-size: 11px;
  line-height: 1.65;
  text-align: center;
}

@keyframes springSlideUp {
  0% {
    opacity: 0;
    transform: translateY(110%);
  }

  68% {
    opacity: 1;
    transform: translateY(-10px);
  }

  84% {
    transform: translateY(4px);
  }

  100% {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
