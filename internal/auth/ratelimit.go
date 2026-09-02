package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter throttles login attempts per client. The portfolio password gets
// shared around, so the login form will be probed; this blunts brute force
// without adding a dependency or a store.
type Limiter struct {
	mu      sync.Mutex
	seen    map[string]*window
	max     int
	period  time.Duration
	nowFunc func() time.Time // overridable in tests
}

type window struct {
	count   int
	resetAt time.Time
}

func NewLimiter(max int, period time.Duration) *Limiter {
	return &Limiter{
		seen:    make(map[string]*window),
		max:     max,
		period:  period,
		nowFunc: time.Now,
	}
}

// Allow records an attempt for key and reports whether it may proceed.
func (l *Limiter) Allow(key string) bool {
	now := l.nowFunc()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Sweep expired entries occasionally so a flood of unique IPs can't grow
	// the map without bound.
	if len(l.seen) > 1024 {
		for k, w := range l.seen {
			if now.After(w.resetAt) {
				delete(l.seen, k)
			}
		}
	}

	w, ok := l.seen[key]
	if !ok || now.After(w.resetAt) {
		l.seen[key] = &window{count: 1, resetAt: now.Add(l.period)}
		return true
	}
	if w.count >= l.max {
		return false
	}
	w.count++
	return true
}

// Reset clears a client's attempt count, called after a successful login so a
// legitimate user isn't penalised for earlier typos.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.seen, key)
}

// RetryAfter reports how long key must wait, or 0 if it may attempt now.
func (l *Limiter) RetryAfter(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	w, ok := l.seen[key]
	if !ok || w.count < l.max {
		return 0
	}
	if d := w.resetAt.Sub(l.nowFunc()); d > 0 {
		return d
	}
	return 0
}

// ClientIP identifies the caller for rate-limiting purposes.
//
// In production every request arrives through Disco's reverse proxy, so
// RemoteAddr is always the proxy and would collapse all visitors into one
// bucket — letting a single attacker lock out everyone. The proxy appends the
// peer address it actually saw to X-Forwarded-For, so with exactly one trusted
// hop the rightmost entry is the real client. A forged header only ever adds
// entries to the left of it.
//
// trustProxy must be false when the server is exposed directly, otherwise a
// client could spoof the header and sidestep the limit entirely.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
