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
      return Promise.resolve({
        mode: 'contact_me',
        corpId: payload.corp_id,
        contactWayConfigId: payload.contact_way_config_id,
        butlerName: payload.butler_name || '',
        needsContactMeButton: true,
      })
    }

    uni.showToast({ title: '管家联系配置缺失', icon: 'none' })
    return Promise.reject(new Error('missing contact_way_config_id'))
  }

  uni.showToast({ title: '管家联系配置异常', icon: 'none' })
  return Promise.reject(new Error('invalid wecom contact mode'))
}
