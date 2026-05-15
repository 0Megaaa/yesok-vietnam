<script>
export default {
  onLaunch() {
    // 1.意图 -> 启动应用时保留轻量日志，不再注入前端 Mock 状态。
    // 2.步骤 -> 由各页面自行请求真实 Go API 与数据库数据。
    // 3.返回 -> 无返回值。
    console.log('Yesok Vietnam real data app launched')
  },
  onShow() {
    // 1.意图 -> 记录应用回到前台的生命周期，方便沙盒联调。
    // 2.步骤 -> 输出前台日志，不改变业务状态。
    // 3.返回 -> 无返回值。
    console.log('Yesok Vietnam app show')
  },
  onHide() {
    // 1.意图 -> 记录应用进入后台的生命周期，避免隐藏副作用。
    // 2.步骤 -> 输出后台日志，不清理鉴权数据。
    // 3.返回 -> 无返回值。
    console.log('Yesok Vietnam app hide')
  },
}
</script>

<style>
@import './style.css';
</style>
