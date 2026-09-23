// Progressive enhancement. Pages work without this file; we only add the "js" class.
document.documentElement.classList.add("js");

// Meta readouts (meta.templ). Each no-ops when its nodes are absent.

// Matches meta.go uptimeLabel so the first tick doesn't jump.
function formatDuration(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds));
  const days = Math.floor(s / 86400);
  const hours = String(Math.floor((s % 86400) / 3600)).padStart(2, "0");
  const mins = String(Math.floor((s % 3600) / 60)).padStart(2, "0");
  const secs = String(s % 60).padStart(2, "0");
  return `${days}d ${hours}:${mins}:${secs}`;
}

// Clock uses data-tz (site zone), not the visitor's.
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

// Scramble: settle left-to-right on [data-scramble]. Prefer inner [data-label].
const SCRAMBLE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789#%&*+/<>";

function scramble(el, { step = 4, fps = 15 } = {}) {
  const target = el.querySelector("[data-label]") || el;
  const original = target.dataset.scrambleText || target.textContent;
  target.dataset.scrambleText = original;
  // Keep the control's accessible name stable while glyphs cycle.
  if (!el.getAttribute("aria-label")) {
    el.setAttribute("aria-label", original.trim());
  }
  const len = original.length;
  let frame = 0;
  clearInterval(target._scrambleTimer);
  target._scrambleTimer = setInterval(() => {
    frame++;
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

// Magnetic: rAF lerp; measure centre on pointerenter; stop when settled.
// Opt in with data-magnetic (+ optional [data-label] for counter-move).

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
  el.addEventListener("focus", measure);
  el.addEventListener("blur", reset);
}

function initMagnetic() {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  if (!window.matchMedia("(hover: hover) and (pointer: fine)").matches) return;
  document.querySelectorAll("[data-magnetic]").forEach(bindMagnetic);
}

// Arm fill-out after first enter so leave can wipe without a page-load flash.
function initHoverFillArm() {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  document.querySelectorAll(".hover-fill-group").forEach((el) => {
    const arm = () => el.classList.add("is-armed");
    el.addEventListener("pointerenter", arm, { once: true });
    el.addEventListener("focusin", arm, { once: true });
  });
}

tickClocks();
tickUptime();
tickScreenTime();
initScramble();
initMagnetic();
initHoverFillArm();
