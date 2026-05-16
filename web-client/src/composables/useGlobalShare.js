import { onShareAppMessage, onShareTimeline } from '@dcloudio/uni-app'

const DEFAULT_SHARE_TITLE = 'YesOK越南管家｜越南一站式中文服务平台'
const DEFAULT_SHARE_PATH = '/pages/home/index'
const DEFAULT_SHARE_IMAGE = '/static/img.png'

// useGlobalShare 为页面注入统一分享配置。
// 实现步骤：
// 1. 合并默认分享标题、路径和图片，避免每个页面重复硬编码。
// 2. 在微信小程序等支持环境中注册好友分享与朋友圈分享钩子。
// 3. 在普通 H5/TG 预览环境中只返回配置，不强行注册平台生命周期，避免浏览器白屏。
export function useGlobalShare(options = {}) {
  const shareConfig = {
    title: options.title || DEFAULT_SHARE_TITLE,
    path: options.path || DEFAULT_SHARE_PATH,
    imageUrl: options.imageUrl || DEFAULT_SHARE_IMAGE,
  }

  // registerShareLifecycle 安全注册平台分享生命周期。
  // 实现步骤：
  // 1. 判断生命周期函数是否存在，避免 H5 兜底预览中调用空实现抛错。
  // 2. 分别注册好友分享和朋友圈分享。
  // 3. 捕获平台差异异常，保证分享能力不影响主页面渲染。
  const registerShareLifecycle = () => {
    try {
      if (typeof onShareAppMessage === 'function') {
        onShareAppMessage(() => ({
          title: shareConfig.title,
          path: shareConfig.path,
          imageUrl: shareConfig.imageUrl,
        }))
      }

      if (typeof onShareTimeline === 'function') {
        onShareTimeline(() => ({
          title: shareConfig.title,
          query: shareConfig.path.includes('?') ? shareConfig.path.split('?')[1] : '',
          imageUrl: shareConfig.imageUrl,
        }))
      }
    } catch (error) {
      console.warn('[useGlobalShare] 当前运行环境不支持平台分享生命周期，已降级为 H5 配置占位。', error)
    }
  }

  registerShareLifecycle()

  return shareConfig
}
