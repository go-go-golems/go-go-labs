// Record a simple, serialisable snapshot for each mutation entry.
function serialiseMutation(m) {
  const ts = new Date().toISOString();
  switch (m.type) {
    case "childList":
      return [...m.addedNodes].map((n) => ({
        ts,
        type: "addedNode",
        nodeName: n.nodeName,
        snippet: n.outerHTML?.slice(0, 120) || n.textContent.slice(0, 120),
      }));
    case "attributes":
      return [
        {
          ts,
          type: "attributeChange",
          nodeName: m.target.nodeName,
          attr: m.attributeName,
          newValue: m.target.getAttribute(m.attributeName),
        },
      ];
    case "characterData":
      return [
        {
          ts,
          type: "textChange",
          nodeName: m.target.parentNode?.nodeName || "#text",
          newValue: m.target.data.slice(0, 120),
        },
      ];
    default:
      return [];
  }
}

const observer = new MutationObserver((mutations) => {
  console.log(
    `[DOM\u2011MUTATION\u2011LOGGER] observed ${mutations.length} raw mutations`
  );
  const records = mutations.flatMap(serialiseMutation);
  if (records.length) {
    console.table(records);
    browser.runtime.sendMessage({ type: "STORE_MUTATIONS", records });
  }
});

observer.observe(document.documentElement, {
  childList: true,
  subtree: true,
  attributes: true,
  characterData: true,
});
