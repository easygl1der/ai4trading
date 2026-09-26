(() => {
  const base = "/documents";
  let indexPromise;
  let overlay;

  const escapeHtml = value => String(value ?? "")
    .replaceAll("&", "&amp;").replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;").replaceAll('"', "&quot;");

  const loadIndex = () => indexPromise ||= fetch(`${base}/search-index.json`, {cache: "force-cache"})
    .then(response => response.ok ? response.json() : Promise.reject(new Error("search index unavailable")));

  const close = () => {
    overlay?.remove();
    overlay = undefined;
  };

  const score = (entry, terms) => terms.reduce((total, term) => {
    const title = entry.title.toLowerCase();
    const text = entry.text.toLowerCase();
    return total + (title.includes(term) ? 16 : 0) + (text.includes(term) ? 1 : 0);
  }, 0);

  const render = (input, results, status) => {
    const terms = input.value.trim().toLowerCase().split(/\s+/).filter(Boolean);
    const matches = terms.length
      ? results.map(entry => ({entry, score: score(entry, terms)})).filter(result => result.score).sort((a, b) => b.score - a.score).slice(0, 12).map(result => result.entry)
      : results.slice(0, 8);
    status.innerHTML = matches.length ? matches.map(entry => `<a href="${base}${entry.path}" class="polybot-search-result"><strong>${escapeHtml(entry.title)}</strong><span>${escapeHtml(entry.excerpt)}</span></a>`).join("") : "<p class=\"polybot-search-empty\">No matching research page.</p>";
  };

  const open = async () => {
    if (overlay) return;
    overlay = document.createElement("div");
    overlay.className = "polybot-search-overlay";
    overlay.innerHTML = `<section class="polybot-search-dialog" role="dialog" aria-modal="true" aria-label="Search Polybot research"><div class="polybot-search-heading"><span>Search research</span><button type="button" aria-label="Close search">Esc</button></div><input autofocus placeholder="Search reports, parameters, findings…"/><div class="polybot-search-results"><p class="polybot-search-empty">Loading research index…</p></div></section>`;
    document.body.append(overlay);
    const input = overlay.querySelector("input");
    const results = overlay.querySelector(".polybot-search-results");
    overlay.addEventListener("click", event => { if (event.target === overlay || event.target.closest("button")) close(); });
    input.focus();
    try {
      const entries = await loadIndex();
      render(input, entries, results);
      input.addEventListener("input", () => render(input, entries, results));
    } catch {
      results.innerHTML = "<p class=\"polybot-search-empty\">Search index could not be loaded. Refresh and try again.</p>";
    }
  };

  const style = document.createElement("style");
  style.textContent = `.polybot-search-overlay{position:fixed;inset:0;z-index:9999;background:rgba(14,22,18,.28);backdrop-filter:blur(5px);padding:10vh 18px}.polybot-search-dialog{max-width:720px;margin:auto;background:#fff;border:1px solid #d8ded9;border-radius:16px;box-shadow:0 25px 70px rgba(20,34,29,.22);overflow:hidden}.polybot-search-heading{display:flex;justify-content:space-between;align-items:center;padding:15px 18px;color:#16785f;font-weight:700}.polybot-search-heading button{border:0;background:#eff3ef;border-radius:6px;padding:5px 9px;color:#526158;cursor:pointer}.polybot-search-dialog input{width:100%;border:0;border-top:1px solid #e4e9e5;border-bottom:1px solid #e4e9e5;padding:17px 18px;font:16px inherit;outline:none}.polybot-search-results{max-height:55vh;overflow:auto;padding:8px}.polybot-search-result{display:block;padding:13px 12px;border-radius:9px;text-decoration:none;color:#17211d}.polybot-search-result:hover{background:#eff7f2}.polybot-search-result strong,.polybot-search-result span{display:block}.polybot-search-result span{margin-top:4px;color:#63716a;font-size:13px;line-height:1.45}.polybot-search-empty{padding:18px;color:#63716a}`;
  document.head.append(style);

  document.addEventListener("click", event => {
    if (event.target.closest("#search-bar-entry")) {
      event.preventDefault();
      event.stopPropagation();
      open();
    }
  }, true);
  document.addEventListener("keydown", event => {
    if (event.key === "Escape" && overlay) return close();
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
      event.preventDefault();
      open();
    }
  });
})();
