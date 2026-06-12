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

export function openWecomContact(payload = {}) {
  const mode = payload.contact_mode || 'customer_service'

  if (mode === 'customer_service') {
    return openWecomCustomerService({
      corpId: payload.corp_id,
      url: payload.service_url,
    })
  }

  if (mode === 'contact_me') {
    if (payload.service_url) {
      return openWecomCustomerService({
        corpId: payload.corp_id,
        url: payload.service_url,
      })
    }

    if (payload.contact_way_config_id) {
      console.warn('[WeCom] contact_me config_id:', payload.contact_way_config_id)

      uni.showModal({
        title: '已通知专属管家',
        content: '管家已收到您的订单联系提醒，请等待管家主动联系您。',
        showCancel: false,
        confirmText: '我知道了',
      })

      return Promise.resolve({
        mode: 'contact_me',
        contactWayConfigId: payload.contact_way_config_id,
        notified: true,
      })
    }

    uni.showToast({ title: '管家联系配置缺失', icon: 'none' })
    return Promise.reject(new Error('missing contact_way_config_id'))
  }

  uni.showToast({ title: '管家联系配置异常', icon: 'none' })
  return Promise.reject(new Error('invalid wecom contact mode'))
}
