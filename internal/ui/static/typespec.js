// /admin/type only: fills each role row with its live computed type values,
// lists every type token the browser resolves, and flags type-* recipes that
// exist in the CSS but have no row. Loaded via Page.Scripts, never site-wide.
(() => {
  const round = (n) => Math.round(n * 100) / 100;
  const round3 = (n) => Math.round(n * 1000) / 1000;

  const describe = (el) => {
    const cs = getComputedStyle(el);
    const size = parseFloat(cs.fontSize);
    const family = cs.fontFamily.split(",")[0].replace(/["']/g, "").trim();
    let lh = cs.lineHeight;
    if (lh.endsWith("px")) lh = `${round(parseFloat(lh))}px (${round3(parseFloat(lh) / size)})`;
    let ls = cs.letterSpacing;
    if (ls.endsWith("px")) ls = `${round(parseFloat(ls))}px (${round3(parseFloat(ls) / size)}em)`;
    return [
      family,
      `${round(size)}px`,
      `line ${lh}`,
      `tracking ${ls}`,
      `weight ${cs.fontWeight}`,
      `case ${cs.textTransform}`,
    ].join(" · ");
  };

  const fillRows = () => {
    for (const row of document.querySelectorAll("[data-type-row]")) {
      const out = row.querySelector("[data-type-values]");
      const el = row.querySelector(`[data-type-short] ${row.dataset.typeTarget}`);
      if (!out) continue;
      if (!el) {
        out.textContent = `no element matches "${row.dataset.typeTarget}"`;
        continue;
      }
      out.textContent = describe(el);
      out.title = getComputedStyle(el).fontFamily;
    }
  };

  // Walk every readable stylesheet, including rules nested in @layer/@media.
  const eachRule = (fn) => {
    const walk = (rules) => {
      for (const rule of rules) {
        fn(rule);
        if (rule.cssRules) walk(rule.cssRules);
      }
    };
    for (const sheet of document.styleSheets) {
      try {
        walk(sheet.cssRules);
      } catch {
        // Cross-origin sheet: nothing of ours in it.
      }
    }
  };

  // Theme colour roles share the --text- prefix; they aren't type.
  const notType = new Set(["--text-primary", "--text-secondary", "--text-muted", "--text-navigation"]);
  const isTypeToken = (name) =>
    /^--(font|text|tracking|leading|case)-/.test(name) && !notType.has(name);

  const listTokens = () => {
    const box = document.querySelector("[data-type-tokens]");
    if (!box) return;
    const names = [];
    eachRule((rule) => {
      if (!rule.style || !/(^|,)\s*:root\b/.test(rule.selectorText || "")) return;
      for (const name of rule.style) {
        if (isTypeToken(name) && !names.includes(name)) names.push(name);
      }
    });
    const root = getComputedStyle(document.documentElement);
    const values = names.map((n) => [n, root.getPropertyValue(n).trim()]);
    const byValue = new Map();
    for (const [n, v] of values) byValue.set(v, [...(byValue.get(v) || []), n]);

    box.replaceChildren(
      ...values.map(([n, v]) => {
        const line = document.createElement("div");
        const name = document.createElement("span");
        name.textContent = `${n}: `;
        const value = document.createElement("span");
        value.className = "text-dim";
        const same = (byValue.get(v) || []).filter((o) => o !== n && !n.startsWith("--font-"));
        value.textContent = v + (same.length && v.length < 40 ? `   same as ${same.join(", ")}` : "");
        line.append(name, value);
        return line;
      }),
    );
  };

  const flagMissing = () => {
    const out = document.querySelector("[data-type-missing]");
    if (!out) return;
    const listed = new Set(
      [...document.querySelectorAll("[data-type-row]")].map((r) => r.dataset.typeUse),
    );
    const found = new Set();
    eachRule((rule) => {
      const m = /^\.(type-[\w-]+)$/.exec(rule.selectorText || "");
      if (m) found.add(m[1]);
    });
    const missing = [...found].filter((c) => !listed.has(c));
    if (missing.length) {
      out.textContent = `Recipes in the CSS with no row here: ${missing.join(", ")} - add them to typeGroups in typespec.templ.`;
      out.hidden = false;
    }
  };

  let queued = false;
  const refresh = () => {
    if (queued) return;
    queued = true;
    requestAnimationFrame(() => {
      queued = false;
      fillRows();
    });
  };

  const start = () => {
    fillRows();
    listTokens();
    flagMissing();
    // Display size tracks the column width; web fonts can swap in late.
    window.addEventListener("resize", refresh);
    document.fonts?.ready.then(refresh);
  };

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", start);
  else start();
})();
