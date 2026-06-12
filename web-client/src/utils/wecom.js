export function openWecomCustomerService({ corpId, url } = {}) {
  if (!corpId || !url) {
    uni.showToast({ title: '管家客服暂未配置', icon: 'none' })
    return Promise.reject(new Error('missing wecom config'))
  }

  if (typeof wx === 'undefined' || !wx.openCustomerServiceChat) {
    uni.showToast({ title: '当前微信版本暂不支持客服会话', icon: 'none' })
    return Promise.reject(new Error('openCustomerServiceChat unsupported'))
  }

  return new Promise((resolve, reject) => {
    wx.openCustomerServiceChat({
      extInfo: {
        url,
      },
      corpId,
      success: resolve,
      fail: (error) => {
        console.error('[WeCom] openCustomerServiceChat failed:', error)
        uni.showToast({ title: '打开管家客服失败', icon: 'none' })
        reject(error)
      },
    })
  })
}
