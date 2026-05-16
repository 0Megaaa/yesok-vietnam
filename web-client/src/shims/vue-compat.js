// 兼容说明：
// 1. 当前 UniApp 发行包会从 vue 中导入 isInSSRComponentSetup 私有字段。
// 2. 官方 Vue 运行时不稳定暴露该私有字段，pnpm 严格解析下会导致 H5 构建失败。
// 3. 这里使用本地适配层转发官方 Vue 全部能力，并补齐只读占位字段。
// 4. 该文件不承载业务逻辑，仅用于保证微信小程序、H5 与后续 App 构建链路稳定。
export * from '../../node_modules/vue/dist/vue.runtime.esm-bundler.js'

// 兼容步骤：
// 1. 保持默认 SSR 组件 setup 标识为 false。
// 2. 仅满足 UniApp 包的命名导入，避免业务层误用内部私有状态。
// 3. 后续升级 UniApp 到完全匹配的 Vue 版本后，可删除该兼容导出。
export const isInSSRComponentSetup = false

// 兼容步骤：
// 1. UniApp 生命周期封装会调用 Vue 内部 injectHook 注册 onLoad、onShow 等生命周期。
// 2. 标准 Vue 运行时未公开该私有方法，因此这里提供最小兼容实现。
// 3. 若当前组件实例不存在，则安全返回空函数；若存在，则把生命周期回调挂载到实例对应队列中。
// 4. 该实现只处理 UniApp 生命周期注册，不改变 Vue 官方生命周期行为。
export const injectHook = (lifecycleName, lifecycleHandler, targetInstance) => {
  if (!targetInstance || !lifecycleName || typeof lifecycleHandler !== 'function') {
    return () => {}
  }

  const lifecycleQueue = targetInstance[lifecycleName] || (targetInstance[lifecycleName] = [])
  const wrappedLifecycleHandler = (...args) => lifecycleHandler(...args)
  lifecycleQueue.push(wrappedLifecycleHandler)

  return wrappedLifecycleHandler
}

// logError 兼容导出：
// 1.意图 -> 满足当前 UniApp H5 运行时从 vue 导入 logError 的内部依赖。
// 2.步骤 -> 将错误透传到 console.error，避免吞掉真实异常。
// 3.返回 -> 无返回值，仅作为运行时错误记录函数。
export const logError = (err, type, context) => {
  console.error('[Yesok Vue Runtime]', type || 'runtime', context || '', err)
}

// 激活生命周期兼容导出：
// 1.意图 -> 满足 UniApp H5 对 Vue 内部 onBeforeActivate/onBeforeDeactivate 的命名导入。
// 2.步骤 -> 使用 injectHook 将回调安全注册到当前组件实例队列。
// 3.返回 -> 生命周期卸载函数或空函数。
export const onBeforeActivate = (hook, target) => injectHook('ba', hook, target)
export const onBeforeDeactivate = (hook, target) => injectHook('bda', hook, target)

// createVueApp 兼容导出：
// 1.意图 -> 对齐 UniApp H5 官方入口转换中 createVueApp as createSSRApp 的运行时约定。
// 2.步骤 -> 从 Vue 官方运行时按需引入 createApp，并以 createVueApp 名称重新导出。
// 3.返回 -> 浏览器端 Vue 应用创建函数。
export { createApp as createVueApp } from '../../node_modules/vue/dist/vue.runtime.esm-bundler.js'
