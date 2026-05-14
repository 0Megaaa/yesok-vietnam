<template>
  <div id="app">
    <div id="scroll">
      <RouterView />
    </div>

    <nav id="botnav">
      <RouterLink to="/" class="bni" active-class="on" exact>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>
        <div class="bni-lb">首页</div>
        <div class="bni-dot"></div>
      </RouterLink>
      <RouterLink to="/news" class="bni" active-class="on">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
        <div class="bni-lb">资讯</div>
        <div class="bni-dot"></div>
      </RouterLink>
      <RouterLink to="/services" class="bni" active-class="on">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14M4.93 4.93a10 10 0 0 0 0 14.14"/></svg>
        <div class="bni-lb">服务</div>
        <div class="bni-dot"></div>
      </RouterLink>
      <RouterLink to="/profile" class="bni" active-class="on">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        <div class="bni-lb">我的</div>
        <div class="bni-dot"></div>
      </RouterLink>
    </nav>
  </div>
</template>
