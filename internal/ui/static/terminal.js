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

tickClocks();
tickUptime();
tickScreenTime();
