// 监听标签页更新（页面加载、切换）
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === "complete" && tab.url) {
    sendUrl(tab.url);
  }
});

// 监听标签页激活（切换标签）
chrome.tabs.onActivated.addListener(async (activeInfo) => {
  const tab = await chrome.tabs.get(activeInfo.tabId);
  if (tab.url) {
    sendUrl(tab.url);
  }
});

function sendUrl(url) {
  const payload = { url: url };
  console.log("即将发送 URL:", payload);

  fetch("http://127.0.0.1:7887/url", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  })
    .then(async (response) => {
      const text = await response.text();
      console.log("接口响应:", text);
    })
    .catch((error) => {
      console.log("请求失败:", error);
    });
}
