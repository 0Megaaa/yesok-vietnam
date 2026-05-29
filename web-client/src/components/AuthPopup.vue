<script setup>
import {computed} from 'vue'
import {useClientStore} from '@/store/client'

const client = useClientStore()
const visible = computed(() => client.loginSheetVisible)

// showSafeToast 安全展示轻提示。
// 意图：兼容微信小程序、H5 与普通浏览器预览环境。
// 实现步骤：
// 1. 优先使用 UniApp 的 showToast API。
// 2. 普通浏览器环境降级为控制台输出。
// 3. 保证登录异常不会阻断弹窗交互。
// 返回：无返回值，仅完成用户反馈。
const showSafeToast = (title) => {
  if (typeof uni !== 'undefined' && uni?.showToast) {
    uni.showToast({title, icon: 'none'})
    return
  }
  console.info('[Yesok Auth]', title)
}

// handleDemoLogin 执行演示登录动作。
// 意图：在不联调真实后端的前提下打通 C 端鉴权闭环。
// 实现步骤：
// 1. 调用客户端 store 的 Mock 登录方法。
// 2. 由 store 统一写入 token、用户资料和本地缓存。
// 3. 登录失败时展示中文轻提示，方便验收人员定位问题。
// 返回：Promise 登录结果，页面无需额外处理。
const handleDemoLogin = async () => {
  try {
    return await client.loginByDemo()
  } catch (error) {
    showSafeToast('登录失败，请稍后重试')
    return null
  }
}

// handleWechatAuthorize 处理微信小程序授权登录。
// 意图：调用微信 wx.login 获取 code，再通过后端换取真实 JWT token。
// 实现步骤：
// 1. 调用 store.loginByWechat(phoneCode) 执行完整的微信登录流程。
// 2. 登录失败时展示中文轻提示，方便用户定位问题。
// 返回：Promise 登录结果。
const handleWechatAuthorize = async (e) => {
  try {
    const phoneCode = e?.detail?.code || ''
    return await client.loginByWechat(phoneCode)
  } catch (error) {
    console.error('[WechatLogin] failed:', error)
    showSafeToast(error?.message || '微信登录失败，请稍后重试')
    return null
  }
}

// handleTelegramPlaceholder 预留 Telegram Mini App 登录入口。
// 意图：保留 TG 无感登录产品位，但本阶段严格不在前端信任 initData。
// 实现步骤：
// 1. 提示验收人员 TG 登录占位已完成。
// 2. 不调用真实 Telegram SDK，避免跨端预览报错。
// 3. 引导继续使用演示登录完成页面验收。
// 返回：无返回值。
const handleTelegramPlaceholder = () => {
  showSafeToast('TG 登录占位已保留，当前请用演示登录')
}
</script>

<template>
  <view v-if="visible" class="auth-mask" @click="client.closeLoginSheet">
    <view class="auth-popup" @click.stop>
      <view class="auth-handle"></view>
      <view class="auth-vip-mark">YESOK PASSPORT</view>
      <view class="auth-title">登录后{{ client.pendingActionText }}</view>
      <view class="auth-desc">
        Yesok 将为您建立专属越南管家档案，用于订单进度、材料提醒与节点验收。
<!--        当前演示版仅使用 Mock 数据，不会向真实后端提交个人信息。-->
      </view>

      <!-- #ifdef MP-WEIXIN -->
      <button class="auth-primary" open-type="getPhoneNumber" @getphonenumber="handleWechatAuthorize">
        微信一键授权
      </button>
<!--      <button class="auth-secondary" @click="handleDemoLogin">使用演示身份继续</button>-->
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

<!--      <view class="auth-tips">-->
<!--        <text>已隔离微信小程序、H5、Telegram Mini App 与未来 iOS/Android 登录入口。</text>-->
<!--      </view>-->
    </view>
  </view>
</template>

<style scoped>
/* 意图：创建全屏遮罩，让用户明确当前操作需要先完成授权。 */
/* 步骤：固定定位覆盖视口，底部对齐弹窗，并使用深绿色半透明蒙层强化高端质感。 */
/* 返回：一个跨端稳定的底部授权容器。 */
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

/* 意图：打造热带奢华风格的底部 Spring 弹窗。 */
/* 步骤：使用大圆角、香槟金描边、深绿色柔光和弹性关键帧动画。 */
/* 返回：点击咨询或下单时从底部弹性滑出的跨端登录面板。 */
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
