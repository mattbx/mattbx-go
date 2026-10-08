// Custom cursor: instant crosshairs + 100ms-sampled tapered trail.
// Trial feature — removable with this file, its CSS block, and one script tag.
// Native cursor stays visible. Pointer-fine devices only.

(function () {
  if (!window.matchMedia("(any-hover: hover) and (pointer: fine)").matches) return;

  const COUNT = 10;
  const INTERVAL = 100;
  const HEAD = 1.4;
  const IDLE_MS = 2000;
  const reduce = window.matchMedia("(prefers-reduced-motion: reduce)");

  const color =
    getComputedStyle(document.documentElement).getPropertyValue("--cross-color").trim() ||
    "rgb(5 5 5 / 0.72)";

  const ruleH = document.createElement("div");
  ruleH.className = "cursor-rule cursor-rule--h";
  ruleH.setAttribute("aria-hidden", "true");
  const ruleV = document.createElement("div");
  ruleV.className = "cursor-rule cursor-rule--v";
  ruleV.setAttribute("aria-hidden", "true");
  document.body.append(ruleH, ruleV);

  let canvas = null;
  let ctx = null;
  if (!reduce.matches) {
    canvas = document.createElement("canvas");
    canvas.className = "cursor-trail";
    canvas.setAttribute("aria-hidden", "true");
    ctx = canvas.getContext("2d");
    if (ctx) document.body.append(canvas);
    else canvas = null;
  }

  const raw = { x: 0, y: 0 };
  const points = [];
  for (let i = 0; i < COUNT; i++) points.push({ x: 0, y: 0 });

  let raf = 0;
  let last = 0;
  let acc = 0;
  let lastPointer = 0;
  let resizeTimer = 0;
  let running = false;
  let seeded = false;

  const moveCross = (x, y) => {
    ruleH.style.transform = `translateY(${y}px) translateZ(0)`;
    ruleV.style.transform = `translateX(${x}px) translateZ(0)`;
  };

  const resize = () => {
    if (!canvas || !ctx) return;
    const root = document.documentElement;
    const w = root.clientWidth;
    const h = root.clientHeight;
    const dpr = Math.min(2, window.devicePixelRatio || 1);
    canvas.width = Math.floor(w * dpr);
    canvas.height = Math.floor(h * dpr);
    canvas.style.width = `${w}px`;
    canvas.style.height = `${h}px`;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  };

  const samplesEqual = () => {
    const a = points[0];
    for (let i = 1; i < COUNT; i++) {
      if (points[i].x !== a.x || points[i].y !== a.y) return false;
    }
    return true;
  };

  const drawRibbon = () => {
    if (!ctx || !canvas) return;
    const w = document.documentElement.clientWidth;
    const h = document.documentElement.clientHeight;
    ctx.clearRect(0, 0, w, h);
    if (samplesEqual()) return;

    const n = COUNT;
    const left = [];
    const right = [];
    for (let i = 0; i < n; i++) {
      const prev = points[Math.max(0, i - 1)];
      const next = points[Math.min(n - 1, i + 1)];
      let dx = next.x - prev.x;
      let dy = next.y - prev.y;
      const len = Math.hypot(dx, dy);
      if (len < 1e-4) {
        dx = 1;
        dy = 0;
      } else {
        dx /= len;
        dy /= len;
      }
      // Perpendicular normal.
      const nx = -dy;
      const ny = dx;
      const width = HEAD * (i / (n - 1));
      const hx = nx * (width / 2);
      const hy = ny * (width / 2);
      const p = points[i];
      left.push({ x: p.x + hx, y: p.y + hy });
      right.push({ x: p.x - hx, y: p.y - hy });
    }

    ctx.beginPath();
    ctx.moveTo(left[0].x, left[0].y);
    for (let i = 1; i < n; i++) ctx.lineTo(left[i].x, left[i].y);
    for (let i = n - 1; i >= 0; i--) ctx.lineTo(right[i].x, right[i].y);
    ctx.closePath();
    ctx.fillStyle = color;
    ctx.fill();
  };

  const tick = (now) => {
    raf = 0;
    if (!ctx) return;

    acc += Math.min(now - last, 250);
    last = now;
    let guard = 0;
    while (acc >= INTERVAL && guard++ < 8) {
      points.shift();
      points.push({ x: raw.x, y: raw.y });
      acc -= INTERVAL;
    }

    drawRibbon();

    if (samplesEqual() && now - lastPointer > IDLE_MS) {
      ctx.clearRect(0, 0, document.documentElement.clientWidth, document.documentElement.clientHeight);
      running = false;
      return;
    }

    raf = requestAnimationFrame(tick);
  };

  const start = () => {
    if (!ctx || running || reduce.matches) return;
    running = true;
    last = performance.now();
    acc = 0;
    raf = requestAnimationFrame(tick);
  };

  const onPointer = (e) => {
    raw.x = e.clientX;
    raw.y = e.clientY;
    lastPointer = performance.now();
    moveCross(raw.x, raw.y);
    if (!seeded) {
      for (let i = 0; i < COUNT; i++) points[i] = { x: raw.x, y: raw.y };
      seeded = true;
    }
    start();
  };

  const onResize = () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
      resize();
      if (running) drawRibbon();
    }, 100);
  };

  // Pause the loop only — keep DOM + listeners so bfcache restore works.
  const pause = () => {
    if (raf) cancelAnimationFrame(raf);
    raf = 0;
    running = false;
    clearTimeout(resizeTimer);
    if (ctx && canvas) {
      ctx.clearRect(0, 0, document.documentElement.clientWidth, document.documentElement.clientHeight);
    }
  };

  const onPageShow = (e) => {
    if (!e.persisted) return;
    resize();
    moveCross(raw.x, raw.y);
  };

  resize();
  window.addEventListener("pointermove", onPointer, { passive: true });
  window.addEventListener("resize", onResize);
  window.addEventListener("pagehide", pause);
  window.addEventListener("pageshow", onPageShow);

  reduce.addEventListener("change", () => {
    // Full reload of trail presence is simplest; crosshairs stay.
    if (reduce.matches) {
      if (raf) cancelAnimationFrame(raf);
      raf = 0;
      running = false;
      if (canvas && ctx) {
        ctx.clearRect(0, 0, document.documentElement.clientWidth, document.documentElement.clientHeight);
        canvas.remove();
        canvas = null;
        ctx = null;
      }
    } else if (!canvas) {
      canvas = document.createElement("canvas");
      canvas.className = "cursor-trail";
      canvas.setAttribute("aria-hidden", "true");
      ctx = canvas.getContext("2d");
      if (ctx) {
        document.body.append(canvas);
        resize();
      } else {
        canvas = null;
      }
    }
  });
})();
