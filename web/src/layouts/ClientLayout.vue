<template>
  <div id="app">
    <div class="tb-main">
      <div class="logo-wrap">
        <div class="logo-ic">
          <svg viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
        </div>
        <div class="logo-tx-wrap">
          <div class="logo-tx"><span class="yes">Yes</span><span class="ok">ok</span></div>
        </div>
      </div>
      <div class="tb-ics">
        <div class="tb-btn">🇻🇳</div>
        <div class="tb-btn">🎧</div>
      </div>
    </div>

    <div id="scroll">
      <RouterView />
    </div>

    <div id="botnav">
      <RouterLink to="/" class="bni" active-class="on">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
        <div class="bni-lb">首页</div>
        <div class="bni-dot"></div>
      </RouterLink>
      <div class="bni">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
        <div class="bni-lb">服务</div>
        <div class="bni-dot"></div>
      </div>
      <div class="bni">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        <div class="bni-lb">发现</div>
        <div class="bni-dot"></div>
      </div>
      <RouterLink to="/profile" class="bni" active-class="on">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        <div class="bni-lb">我的</div>
        <div class="bni-dot"></div>
      </RouterLink>
    </div>
  </div>
</template>
