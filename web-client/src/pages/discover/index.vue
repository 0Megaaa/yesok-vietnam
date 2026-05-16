<template>
  <view style="background:#F5F8FC;min-height:100vh;">

    <!-- ── 顶部Banner ── -->
    <view style="position:relative;height:160px;overflow:hidden;flex-shrink:0;">
      <img
        src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0naHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmcnIHZpZXdCb3g9JzAgMCA0MzAgMTYwJyB3aWR0aD0nNDMwJyBoZWlnaHQ9JzE2MCc+CiAgPGRlZnM+CiAgICA8bGluZWFyR3JhZGllbnQgaWQ9J3NreScgeDE9JzAnIHkxPScwJyB4Mj0nMCcgeTI9JzEnPgogICAgICA8c3RvcCBvZmZzZXQ9JzAlJyBzdG9wLWNvbG9yPScjMTU2NUMwJy8+PHN0b3Agb2Zmc2V0PScxMDAlJyBzdG9wLWNvbG9yPScjNDJBNUY1Jy8+CiAgICA8L2xpbmVhckdyYWRpZW50PgogICAgPGxpbmVhckdyYWRpZW50IGlkPSd3YXRlcicgeDE9JzAnIHkxPScwJyB4Mj0nMCcgeTI9JzEnPgogICAgICA8c3RvcCBvZmZzZXQ9JzAlJyBzdG9wLWNvbG9yPScjMTk3NkQyJyBzdG9wLW9wYWNpdHk9Jy42Jy8+PHN0b3Agb2Zmc2V0PScxMDAlJyBzdG9wLWNvbG9yPScjMEQ0N0ExJyBzdG9wLW9wYWNpdHk9Jy44Jy8+CiAgICA8L2xpbmVhckdyYWRpZW50PgogICAgPGxpbmVhckdyYWRpZW50IGlkPSdvdmVyJyB4MT0nMCcgeTE9JzAnIHgyPScwJyB5Mj0nMSc+CiAgICAgIDxzdG9wIG9mZnNldD0nMCUnIHN0b3AtY29sb3I9JyMwQjJBNkInIHN0b3Atb3BhY2l0eT0nLjc1Jy8+PHN0b3Agb2Zmc2V0PSc3MCUnIHN0b3AtY29sb3I9JyMwQjJBNkInIHN0b3Atb3BhY2l0eT0nLjInLz48c3RvcCBvZmZzZXQ9JzEwMCUnIHN0b3AtY29sb3I9JyMwQjJBNkInIHN0b3Atb3BhY2l0eT0nLjA1Jy8+CiAgICA8L2xpbmVhckdyYWRpZW50PgogIDwvZGVmcz4KICA8cmVjdCB3aWR0aD0nNDMwJyBoZWlnaHQ9JzE2MCcgZmlsbD0ndXJsKCNza3kpJy8+CiAgPCEtLSB3YXRlciAtLT4KICA8cmVjdCB5PScxMTgnIHdpZHRoPSc0MzAnIGhlaWdodD0nNDInIGZpbGw9J3VybCgjd2F0ZXIpJy8+CiAgPCEtLSBidWlsZGluZ3MgYmcgLS0+CiAgPHJlY3QgeD0nMCcgeT0nNzAnIHdpZHRoPSc0MCcgaGVpZ2h0PSc1MCcgZmlsbD0nIzFBNEU4QScgb3BhY2l0eT0nLjcnLz4KICA8cmVjdCB4PSczOCcgeT0nNjAnIHdpZHRoPSczMCcgaGVpZ2h0PSc2MCcgZmlsbD0nIzFFNTc5OScgb3BhY2l0eT0nLjc1Jy8+CiAgPHJlY3QgeD0nNjYnIHk9Jzc1JyB3aWR0aD0nMjUnIGhlaWdodD0nNDUnIGZpbGw9JyMxQTRFOEEnIG9wYWNpdHk9Jy42NScvPgogIDxyZWN0IHg9Jzg5JyB5PSc1NScgd2lkdGg9JzM1JyBoZWlnaHQ9JzY1JyBmaWxsPScjMUU1Nzk5JyBvcGFjaXR5PScuNycvPgogIDxyZWN0IHg9JzEyMicgeT0nNjUnIHdpZHRoPScyOCcgaGVpZ2h0PSc1NScgZmlsbD0nIzFBNEU4QScgb3BhY2l0eT0nLjY1Jy8+CiAgPCEtLSBMYW5kbWFyayA4MSAtIHRhbGxlc3QgYnVpbGRpbmcgY2VudGVyLXJpZ2h0IC0tPgogIDxyZWN0IHg9JzIzOCcgeT0nMTUnIHdpZHRoPScyMicgaGVpZ2h0PScxMDUnIGZpbGw9JyMxNTY1QzAnLz4KICA8cG9seWdvbiBwb2ludHM9JzI0OSw4IDI0MSwxOCAyNTcsMTgnIGZpbGw9JyMwRDQ3QTEnLz4KICA8cmVjdCB4PScyNDAnIHk9JzE4JyB3aWR0aD0nMTgnIGhlaWdodD0nMTAyJyBmaWxsPScjMTk3NkQyJyBvcGFjaXR5PScuOScvPgogIDwhLS0gTGFuZG1hcmsgODEgd2luZG93cyAtLT4KICA8cmVjdCB4PScyNDInIHk9JzIyJyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMycvPgogIDxyZWN0IHg9JzI0OCcgeT0nMjInIHdpZHRoPSc0JyBoZWlnaHQ9JzMnIGZpbGw9JyNmZmYnIG9wYWNpdHk9Jy4zJy8+CiAgPHJlY3QgeD0nMjU0JyB5PScyMicgd2lkdGg9JzQnIGhlaWdodD0nMycgZmlsbD0nI2ZmZicgb3BhY2l0eT0nLjMnLz4KICA8cmVjdCB4PScyNDInIHk9JzI5JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMycvPgogIDxyZWN0IHg9JzI0OCcgeT0nMjknIHdpZHRoPSc0JyBoZWlnaHQ9JzMnIGZpbGw9JyNmZmYnIG9wYWNpdHk9Jy4zJy8+CiAgPHJlY3QgeD0nMjU0JyB5PScyOScgd2lkdGg9JzQnIGhlaWdodD0nMycgZmlsbD0nI2ZmZicgb3BhY2l0eT0nLjMnLz4KICA8cmVjdCB4PScyNDInIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMjUnLz4KICA8cmVjdCB4PScyNDgnIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMjUnLz4KICA8cmVjdCB4PScyNTQnIHk9JzM2JyB3aWR0aD0nNCcgaGVpZ2h0PSczJyBmaWxsPScjZmZmJyBvcGFjaXR5PScuMjUnLz4KICA8IS0tIG90aGVyIGJ1aWxkaW5ncyByaWdodCBzaWRlIC0tPgogIDxyZWN0IHg9JzI2NScgeT0nNDUnIHdpZHRoPSczMCcgaGVpZ2h0PSc3NScgZmlsbD0nIzE1NjVDMCcgb3BhY2l0eT0nLjc1Jy8+CiAgPHJlY3QgeD0nMjkzJyB5PSc1NScgd2lkdGg9JzI1JyBoZWlnaHQ9JzY1JyBmaWxsPScjMUE0RThBJyBvcGFjaXR5PScuNycvPgogIDxyZWN0IHg9JzMxNicgeT0nNTAnIHdpZHRoPSczNScgaGVpZ2h0PSc3MCcgZmlsbD0nIzFFNTc5OScgb3BhY2l0eT0nLjY1Jy8+CiAgPHJlY3QgeD0nMzQ5JyB5PSc2MCcgd2lkdGg9JzI4JyBoZWlnaHQ9JzYwJyBmaWxsPScjMUE0RThBJyBvcGFjaXR5PScuNjUnLz4KICA8cmVjdCB4PSczNzUnIHk9JzcwJyB3aWR0aD0nMzAnIGhlaWdodD0nNTAnIGZpbGw9JyMxNTY1QzAnIG9wYWNpdHk9Jy42Jy8+CiAgPHJlY3QgeD0nNDAzJyB5PSc2NScgd2lkdGg9JzI3JyBoZWlnaHQ9JzU1JyBmaWxsPScjMUU1Nzk5JyBvcGFjaXR5PScuNjUnLz4KICA8IS0tIHRyZWVzIC0tPgogIDxyZWN0IHg9JzE1MCcgeT0nMTAwJyB3aWR0aD0nNicgaGVpZ2h0PScyMCcgZmlsbD0nIzFBNEU4QScgb3BhY2l0eT0nLjUnLz4KICA8ZWxsaXBzZSBjeD0nMTUzJyBjeT0nOTcnIHJ4PScxMCcgcnk9JzgnIGZpbGw9JyMyRTdEMzInIG9wYWNpdHk9Jy42Jy8+CiAgPHJlY3QgeD0nMTc1JyB5PScxMDInIHdpZHRoPSc1JyBoZWlnaHQ9JzE4JyBmaWxsPScjMUE0RThBJyBvcGFjaXR5PScuNScvPgogIDxlbGxpcHNlIGN4PScxNzcnIGN5PSc5OScgcng9JzknIHJ5PSc3JyBmaWxsPScjMkU3RDMyJyBvcGFjaXR5PScuNTUnLz4KICA8cmVjdCB4PScyMDAnIHk9JzEwMycgd2lkdGg9JzUnIGhlaWdodD0nMTcnIGZpbGw9JyMxQTRFOEEnIG9wYWNpdHk9Jy41Jy8+CiAgPGVsbGlwc2UgY3g9JzIwMicgY3k9JzEwMCcgcng9JzgnIHJ5PSc3JyBmaWxsPScjMkU3RDMyJyBvcGFjaXR5PScuNScvPgogIDwhLS0gYnJpZGdlIGFyYyAtLT4KICA8cGF0aCBkPSdNNjAgMTE4IFExMjAgOTUgMTgwIDExOCcgc3Ryb2tlPScjMTU2NUMwJyBzdHJva2Utd2lkdGg9JzMnIGZpbGw9J25vbmUnIG9wYWNpdHk9Jy41Jy8+CiAgPGxpbmUgeDE9JzkwJyB5MT0nMTA4JyB4Mj0nOTAnIHkyPScxMTgnIHN0cm9rZT0nIzE1NjVDMCcgc3Ryb2tlLXdpZHRoPScxLjUnIG9wYWNpdHk9Jy41Jy8+CiAgPGxpbmUgeDE9JzEyMCcgeTE9JzEwMycgeDI9JzEyMCcgeTI9JzExOCcgc3Ryb2tlPScjMTU2NUMwJyBzdHJva2Utd2lkdGg9JzEuNScgb3BhY2l0eT0nLjUnLz4KICA8bGluZSB4MT0nMTUwJyB5MT0nMTA4JyB4Mj0nMTUwJyB5Mj0nMTE4JyBzdHJva2U9JyMxNTY1QzAnIHN0cm9rZS13aWR0aD0nMS41JyBvcGFjaXR5PScuNScvPgogIDwhLS0gZ3JhZGllbnQgb3ZlcmxheSAtLT4KICA8cmVjdCB3aWR0aD0nNDMwJyBoZWlnaHQ9JzE2MCcgZmlsbD0ndXJsKCNvdmVyKScvPgo8L3N2Zz4="
        style="width:100%;height:160px;object-fit:cover;" alt="">
      <!-- text overlay -->
      <view
        style="position:absolute;inset:0;padding:24px 16px 16px;display:flex;flex-direction:column;justify-content:flex-end;">
        <view
          style="font-size:24px;font-weight:800;color:#fff;line-height:1.2;text-shadow:0 1px 4px rgba(0,0,0,.3);">
          越南资讯</view>
        <view
          style="font-size:13px;color:rgba(255,255,255,.85);margin-top:4px;text-shadow:0 1px 3px rgba(0,0,0,.3);">
          在越南，这点事都帮你盯着</view>
      </view>
      <!-- search btn -->
      <view
        style="position:absolute;top:14px;right:14px;width:36px;height:36px;border-radius:50%;background:rgba(255,255,255,.9);display:flex;align-items:center;justify-content:center;cursor:pointer;box-shadow:0 2px 8px rgba(0,0,0,.15);">
        <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="#102A55" stroke-width="2.5">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
      </view>
    </view>

    <!-- ── 分类导航 ── -->
    <view
      style="background:#fff;padding:12px 16px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;border-bottom:1px solid #E6ECF3;flex-shrink:0;">
      <view class="nc-btn nc-on">全部</view>
      <view class="nc-btn">
        新闻早知道
        <text
          style="background:#E60012;color:#fff;font-size:8px;font-weight:700;padding:1px 4px;border-radius:4px;margin-left:3px;vertical-align:middle;">NEW</text>
      </view>
      <view class="nc-btn">攻略指南</view>
      <view class="nc-btn">签证政策</view>
      <view class="nc-btn">投资商业</view>
    </view>

    <!-- ── 今日精选 ── -->
    <view
      style="background:#fff;margin:10px 12px;border-radius:16px;padding:14px;box-shadow:0 6px 18px rgba(15,47,92,.08);">
      <view style="display:flex;align-items:center;gap:6px;margin-bottom:12px;">
        <text style="font-size:16px;">🔥</text>
        <text style="font-size:15px;font-weight:700;color:#102A55;">今日精选</text>
        <view style="flex:1;"></view>
        <text
          style="font-size:12px;color:#0B4FE8;display:flex;align-items:center;gap:2px;">查看更多 <svg
            width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#0B4FE8" stroke-width="2.5">
            <polyline points="9 18 15 12 9 6" />
          </svg></text>
      </view>
      <view style="display:grid;grid-template-columns:1fr 1fr;gap:10px;min-height:160px;">
        <view
          style="border-radius:12px;overflow:hidden;box-shadow:0 3px 10px rgba(15,47,92,.1);background:#fff;">
          <view
            style="height:90px;background:linear-gradient(135deg,#1565C0,#4285F4);display:flex;align-items:center;justify-content:center;font-size:40px;">
            🏢</view>
          <view style="padding:10px 10px 12px;">
            <text
              style="background:#0B4FE820;color:#0B4FE8;font-size:9px;font-weight:700;padding:2px 7px;border-radius:6px;">公司注册</text>
            <view style="font-size:12px;font-weight:700;color:#102A55;margin-top:4px;line-height:1.3;">
              越南外资公司注册全流程指南（2025）</view>
            <view style="font-size:10px;color:#9AA3B5;margin-top:3px;">👁 2.3万 · 越南投资报</view>
          </view>
        </view>
        <view
          style="border-radius:12px;overflow:hidden;box-shadow:0 3px 10px rgba(15,47,92,.1);background:#fff;">
          <view
            style="height:90px;background:linear-gradient(135deg,#00897B,#26A69A);display:flex;align-items:center;justify-content:center;font-size:40px;">
            🛂</view>
          <view style="padding:10px 10px 12px;">
            <text
              style="background:#00897B20;color:#00897B;font-size:9px;font-weight:700;padding:2px 7px;border-radius:6px;">签证政策</text>
            <view style="font-size:12px;font-weight:700;color:#102A55;margin-top:4px;line-height:1.3;">
              2025年越南签证最新政策解读</view>
            <view style="font-size:10px;color:#9AA3B5;margin-top:3px;">👁 1.8万 · 越南生活通</view>
          </view>
        </view>
        <view
          style="border-radius:12px;overflow:hidden;box-shadow:0 3px 10px rgba(15,47,92,.1);background:#fff;">
          <view
            style="height:90px;background:linear-gradient(135deg,#F57C00,#FF9800);display:flex;align-items:center;justify-content:center;font-size:40px;">
            💰</view>
          <view style="padding:10px 10px 12px;">
            <text
              style="background:#F57C0020;color:#F57C00;font-size:9px;font-weight:700;padding:2px 7px;border-radius:6px;">经济</text>
            <view style="font-size:12px;font-weight:700;color:#102A55;margin-top:4px;line-height:1.3;">
              越南盾汇率波动，投资机会与风险分析</view>
            <view style="font-size:10px;color:#9AA3B5;margin-top:3px;">👁 1.2万 · 经济观察</view>
          </view>
        </view>
        <view
          style="border-radius:12px;overflow:hidden;box-shadow:0 3px 10px rgba(15,47,92,.1);background:#fff;">
          <view
            style="height:90px;background:linear-gradient(135deg,#7B1FA2,#AB47BC);display:flex;align-items:center;justify-content:center;font-size:40px;">
            🏠</view>
          <view style="padding:10px 10px 12px;">
            <text
              style="background:#7B1FA220;color:#7B1FA2;font-size:9px;font-weight:700;padding:2px 7px;border-radius:6px;">房产</text>
            <view style="font-size:12px;font-weight:700;color:#102A55;margin-top:4px;line-height:1.3;">
              胡志明市租房全攻略：避坑+省钱技巧</view>
            <view style="font-size:10px;color:#9AA3B5;margin-top:3px;">👁 9800 · 越南房产网</view>
          </view>
        </view>
      </view>
    </view>

    <!-- ── 内容Tab ── -->
    <view style="background:#fff;padding:0 16px;display:flex;gap:0;border-bottom:1px solid #E6ECF3;flex-shrink:0;">
      <view
        style="padding:12px 0;margin-right:24px;font-size:14px;font-weight:700;color:#0B4FE8;border-bottom:2px solid #0B4FE8;cursor:pointer;">
        最新资讯</view>
      <view
        style="padding:12px 0;font-size:14px;font-weight:600;color:#7A8799;border-bottom:2px solid transparent;cursor:pointer;">
        实用攻略</view>
    </view>

    <!-- ── 最新资讯列表 ── -->
    <view style="background:#F5F8FC;">
      <view
        style="background:#fff;margin:8px 12px;border-radius:16px;padding:14px;display:flex;gap:12px;cursor:pointer;box-shadow:0 6px 18px rgba(15,47,92,.08);">
        <view
          style="width:80px;height:80px;border-radius:12px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);display:flex;align-items:center;justify-content:center;font-size:32px;flex-shrink:0;">
          ✈️</view>
        <view style="flex:1;min-width:0;">
          <view style="margin-bottom:5px;"><text
              style="background:#0B4FE820;color:#0B4FE8;font-size:10px;font-weight:700;padding:2px 8px;border-radius:8px;">政策</text>
          </view>
          <view style="font-size:14px;font-weight:600;color:#102A55;line-height:1.4;margin-bottom:4px;">
            越南加大吸引外资力度，简化外企注册流程</view>
          <view
            style="font-size:12px;color:#7A8799;line-height:1.5;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin-bottom:6px;">
            新规将企业注册时间缩短30%，进一步优化投资环境。</view>
          <view style="display:flex;align-items:center;justify-content:space-between;">
            <text style="font-size:11px;color:#7A8799;">4小时前 · 越南投资报</text>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#C0C8D8" stroke-width="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
          </view>
        </view>
      </view>
      <view
        style="background:#fff;margin:8px 12px;border-radius:16px;padding:14px;display:flex;gap:12px;cursor:pointer;box-shadow:0 6px 18px rgba(15,47,92,.08);">
        <view
          style="width:80px;height:80px;border-radius:12px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);display:flex;align-items:center;justify-content:center;font-size:32px;flex-shrink:0;">
          🛫</view>
        <view style="flex:1;min-width:0;">
          <view style="margin-bottom:5px;"><text
              style="background:#00897B20;color:#00897B;font-size:10px;font-weight:700;padding:2px 8px;border-radius:8px;">交通</text>
          </view>
          <view style="font-size:14px;font-weight:600;color:#102A55;line-height:1.4;margin-bottom:4px;">
            胡志明市新山一机场T3航站楼2025年底竣工</view>
          <view
            style="font-size:12px;color:#7A8799;line-height:1.5;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin-bottom:6px;">
            新航站楼投用后，机场年旅客吞吐能力将大幅提升。</view>
          <view style="display:flex;align-items:center;justify-content:space-between;">
            <text style="font-size:11px;color:#7A8799;">昨天 · 越南交通部</text>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#C0C8D8" stroke-width="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
          </view>
        </view>
      </view>
      <view
        style="background:#fff;margin:8px 12px;border-radius:16px;padding:14px;display:flex;gap:12px;cursor:pointer;box-shadow:0 6px 18px rgba(15,47,92,.08);">
        <view
          style="width:80px;height:80px;border-radius:12px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);display:flex;align-items:center;justify-content:center;font-size:32px;flex-shrink:0;">
          💰</view>
        <view style="flex:1;min-width:0;">
          <view style="margin-bottom:5px;"><text
              style="background:#F57C0020;color:#F57C00;font-size:10px;font-weight:700;padding:2px 8px;border-radius:8px;">经济</text>
          </view>
          <view style="font-size:14px;font-weight:600;color:#102A55;line-height:1.4;margin-bottom:4px;">
            越南盾汇率小幅波动，人民币兑越盾约3550</view>
          <view
            style="font-size:12px;color:#7A8799;line-height:1.5;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin-bottom:6px;">
            市场保持稳定，专家预计短期内波动有限。</view>
          <view style="display:flex;align-items:center;justify-content:space-between;">
            <text style="font-size:11px;color:#7A8799;">昨天 · 越南央行</text>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#C0C8D8" stroke-width="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
          </view>
        </view>
      </view>
      <view
        style="background:#fff;margin:8px 12px;border-radius:16px;padding:14px;display:flex;gap:12px;cursor:pointer;box-shadow:0 6px 18px rgba(15,47,92,.08);">
        <view
          style="width:80px;height:80px;border-radius:12px;background:linear-gradient(135deg,#E3F0FF,#C5D9F5);display:flex;align-items:center;justify-content:center;font-size:32px;flex-shrink:0;">
          🏠</view>
        <view style="flex:1;min-width:0;">
          <view style="margin-bottom:5px;"><text
              style="background:#7B1FA220;color:#7B1FA2;font-size:10px;font-weight:700;padding:2px 8px;border-radius:8px;">房产</text>
          </view>
          <view style="font-size:14px;font-weight:600;color:#102A55;line-height:1.4;margin-bottom:4px;">
            胡志明市2区高端住宅价格逆势上涨12%</view>
          <view
            style="font-size:12px;color:#7A8799;line-height:1.5;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin-bottom:6px;">
            地段稀缺+外资需求增长，推动高端房产热度上升。</view>
          <view style="display:flex;align-items:center;justify-content:space-between;">
            <text style="font-size:11px;color:#7A8799;">2天前 · 越南房产协会</text>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#C0C8D8" stroke-width="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
          </view>
        </view>
      </view>
      <view
        style="margin:8px 12px 14px;padding:13px;border-radius:12px;background:#fff;border:1px solid #E6ECF3;display:flex;align-items:center;justify-content:center;gap:5px;cursor:pointer;">
        <text style="font-size:13px;font-weight:600;color:#0B4FE8;">查看更多资讯</text>
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#0B4FE8" stroke-width="2.5">
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </view>
    </view>

    <!-- ── 实用攻略列表 ── -->
    <view style="background:#F5F8FC;display:none;">
      <view style="padding:0;"></view>
    </view>
  </view>
</template>

<script setup>
</script>

<style scoped>
.nc-btn {
  padding: 5px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  color: #7A8799;
  background: #F5F7FA;
  border: 1px solid #E2E8F4;
  white-space: nowrap;
  cursor: pointer;
  flex-shrink: 0;
}
.nc-btn.nc-on {
  background: #0B4FE8;
  color: #fff;
  border-color: #0B4FE8;
}
</style>
