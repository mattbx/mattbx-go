// First-party progressive enhancement. Every page works without this file;
// effects added here hook into the "js" class so visitors without JavaScript
// get plain, readable HTML.
document.documentElement.classList.add("js");

// --- Metadata readouts (internal/ui/meta.templ) -----------------------------
// Each ticks independently and no-ops if its element isn't on the current
// page, so these run safely on every route regardless of which pieces (if
// any) a given page uses.

// formatDuration renders seconds as "Nd HH:MM:SS", matching the shape the
// server renders initially (internal/ui/meta.go's uptimeLabel), so there's no
// visible jump when JS takes over ticking.
function formatDuration(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds));
  const days = Math.floor(s / 86400);
  const hours = String(Math.floor((s % 86400) / 3600)).padStart(2, "0");
  const mins = String(Math.floor((s % 3600) / 60)).padStart(2, "0");
  const secs = String(s % 60).padStart(2, "0");
  return `${days}d ${hours}:${mins}:${secs}`;
}

// tickClocks keeps [data-meta="clock"] elements on the time zone named in
// their own data-tz, not the visitor's — the site's clock is Sydney's,
// wherever the page is being read from.
function tickClocks() {
  const els = document.querySelectorAll('[data-meta="clock"]');
  if (els.length === 0) return;
  const formatters = new Map();
  const update = () => {
    const now = new Date();
    els.forEach((el) => {
      const tz = el.dataset.tz;
      if (!tz) return;
      let fmt = formatters.get(tz);
      if (!fmt) {
        fmt = new Intl.DateTimeFormat("en-GB", { timeZone: tz, hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false });
        formatters.set(tz, fmt);
      }
      el.textContent = fmt.format(now);
    });
  };
  update();
  setInterval(update, 1000);
}

// tickUptime keeps [data-meta="uptime"] elements counting up from the process
// start time the server rendered into data-started (unix seconds).
function tickUptime() {
  const els = document.querySelectorAll('[data-meta="uptime"]');
  if (els.length === 0) return;
  const update = () => {
    const now = Date.now() / 1000;
    els.forEach((el) => {
      const started = Number(el.dataset.started);
      if (!started) return;
      el.textContent = formatDuration(now - started);
    });
  };
  update();
  setInterval(update, 1000);
}

// tickScreenTime counts up from zero for however long this tab has been
// open. Purely client-side — nothing is sent anywhere or stored.
function tickScreenTime() {
  const els = document.querySelectorAll('[data-meta="screentime"]');
  if (els.length === 0) return;
  const start = Date.now() / 1000;
  const update = () => {
    const elapsed = Date.now() / 1000 - start;
    const mins = String(Math.floor(elapsed / 60)).padStart(2, "0");
    const secs = String(Math.floor(elapsed % 60)).padStart(2, "0");
    els.forEach((el) => { el.textContent = `${mins}:${secs}`; });
  };
  update();
  setInterval(update, 1000);
}

// --- Letter scramble --------------------------------------------------------
// Cycles random glyphs across a short label, settling left-to-right into the
// real text on hover/focus. Scoped to mono, uppercase, short labels via
// [data-scramble]. Defaults match the Jean Dawson teardown (step 4, 15fps).

const SCRAMBLE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789#%&*+/<>";

function scramble(el, { step = 4, fps = 15 } = {}) {
  // Prefer an inner [data-label] so magnetic wrappers keep their DOM.
  const target = el.querySelector("[data-label]") || el;
  const original = target.dataset.scrambleText || target.textContent;
  target.dataset.scrambleText = original;
  const len = original.length;
  let frame = 0;
  clearInterval(target._scrambleTimer);
  target._scrambleTimer = setInterval(() => {
    frame++;
    // Chars left of (frame - step) are final — matches shuffle-letters shape.
    const settled = frame - step;
    let out = "";
    for (let i = 0; i < len; i++) {
      const ch = original[i];
      out += ch === " " || i < settled
        ? ch
        : SCRAMBLE_CHARS[(Math.random() * SCRAMBLE_CHARS.length) | 0];
    }
    target.textContent = out;
    if (settled >= len) {
      target.textContent = original;
      clearInterval(target._scrambleTimer);
    }
  }, 1000 / fps);
}

function initScramble() {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  const els = document.querySelectorAll("[data-scramble]");
  if (els.length === 0) return;
  els.forEach((el) => {
    const run = () => scramble(el);
    el.addEventListener("pointerenter", run);
    el.addEventListener("focus", run);
  });
}

// --- Magnetic pull ----------------------------------------------------------
// rAF lerp toward (pointer - centre) * force. Loop stops when settled.
// Rect measured on pointerenter so scroll never makes the centre stale.
// Gated to fine pointers; reduced-motion skips entirely.

function bindMagnetic(el) {
  const ease = Number(el.dataset.ease) || 0.2;
  const force = Number(el.dataset.force) || 0.25;
  const labelStrength = 1 / 3;
  const inner = el.querySelector("[data-label]");
  const areaSelector = el.dataset.area;
  const area = (areaSelector && document.querySelector(areaSelector)) || el;

  let x = 0, y = 0, tx = 0, ty = 0, cx = 0, cy = 0, raf = 0;

  const tick = () => {
    x += (tx - x) * ease;
    y += (ty - y) * ease;
    el.style.transform = `translate3d(${x}px, ${y}px, 0)`;
    if (inner) {
      inner.style.transform = `translate3d(${-x * labelStrength}px, ${-y * labelStrength}px, 0)`;
    }
    raf = Math.abs(tx - x) + Math.abs(ty - y) > 0.05
      ? requestAnimationFrame(tick)
      : 0;
  };
  const kick = () => { if (!raf) raf = requestAnimationFrame(tick); };

  const measure = () => {
    const r = el.getBoundingClientRect();
    cx = r.left + r.width / 2 - x;
    cy = r.top + r.height / 2 - y;
  };
  const move = (e) => {
    tx = (e.clientX - cx) * force;
    ty = (e.clientY - cy) * force;
    kick();
  };
  const reset = () => { tx = 0; ty = 0; kick(); };

  area.addEventListener("pointerenter", measure);
  area.addEventListener("pointermove", move);
  area.addEventListener("pointerleave", reset);
  // Focus parity for keyboard users — mild nudge toward centre is a no-op;
  // measure still keeps hover ready after tabbing.
  el.addEventListener("focus", measure);
  el.addEventListener("blur", reset);
}

function initMagnetic() {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  if (!window.matchMedia("(hover: hover) and (pointer: fine)").matches) return;
  document.querySelectorAll("[data-magnetic]").forEach(bindMagnetic);
}

tickClocks();
tickUptime();
tickScreenTime();
initScramble();
initMagnetic();
