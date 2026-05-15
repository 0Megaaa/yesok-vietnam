import { onShareAppMessage, onShareTimeline } from '@dcloudio/uni-app'

const DEFAULT_SHARE_TITLE = 'YesOK越南管家｜越南一站式中文服务平台'
const DEFAULT_SHARE_PATH = '/pages/home/index'
const DEFAULT_SHARE_IMAGE = '/static/img.png'

// useGlobalShare 为页面注入统一分享配置。
// 实现步骤：
// 1. 合并默认分享标题、路径和图片，避免每个页面重复硬编码。
// 2. 注册微信小程序好友分享与朋友圈分享钩子。
// 3. H5/TG 环境暂不强依赖分享 SDK，只保留同一份配置供后续扩展。
export function useGlobalShare(options = {}) {
  const shareConfig = {
    title: options.title || DEFAULT_SHARE_TITLE,
    path: options.path || DEFAULT_SHARE_PATH,
    imageUrl: options.imageUrl || DEFAULT_SHARE_IMAGE,
  }

  onShareAppMessage(() => ({
    title: shareConfig.title,
    path: shareConfig.path,
    imageUrl: shareConfig.imageUrl,
  }))

  onShareTimeline(() => ({
    title: shareConfig.title,
    query: shareConfig.path.includes('?') ? shareConfig.path.split('?')[1] : '',
    imageUrl: shareConfig.imageUrl,
  }))

  return shareConfig
}
