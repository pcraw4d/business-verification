package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// IPBlocker blocks requests from abusive IPs based on response patterns
// Suspicious activity heuristic: many 4xx/401/403/429 within a short window.
type IPBlocker struct {
	enabled       bool
	threshold     int
	window        time.Duration
	blockDuration time.Duration
	whitelist     map[string]struct{}
	blacklist     map[string]struct{}
	logger        *observability.Logger

	mu       sync.RWMutex
	records  map[string]*ipRecord
	cleanupC chan struct{}
}

type ipRecord struct {
	failures    int
	firstSeen   time.Time
	blockedTill *time.Time
}

// NewIPBlocker constructs a new IP blocker middleware
func NewIPBlocker(enabled bool, threshold int, window, blockDuration time.Duration, whitelist, blacklist []string, logger *observability.Logger) *IPBlocker {
	b := &IPBlocker{
		enabled:       enabled,
		threshold:     threshold,
		window:        window,
		blockDuration: blockDuration,
		whitelist:     sliceToSet(whitelist),
		blacklist:     sliceToSet(blacklist),
		logger:        logger,
		records:       make(map[string]*ipRecord),
		cleanupC:      make(chan struct{}),
	}
	go b.cleanup()
	return b
}

// Middleware applies IP blocking to requests
func (b *IPBlocker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		if b.isWhitelisted(ip) {
			next.ServeHTTP(w, r)
			return
		}
		if b.isBlacklisted(ip) || b.isBlocked(ip) {
			b.logger.WithComponent("ip_block").Warn("Blocked request from IP", "client_ip", ip, "path", r.URL.Path)
			http.Error(w, "Access from your IP is temporarily blocked", http.StatusTooManyRequests)
			return
		}

		// Wrap ResponseWriter to inspect status code after handler runs
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)

		// Count suspicious responses
		if rec.status >= 400 {
			b.noteFailure(ip)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (b *IPBlocker) isWhitelisted(ip string) bool {
	_, ok := b.whitelist[ip]
	return ok
}

func (b *IPBlocker) isBlacklisted(ip string) bool {
	_, ok := b.blacklist[ip]
	return ok
}

func (b *IPBlocker) isBlocked(ip string) bool {
	b.mu.RLock()
	rec, ok := b.records[ip]
	b.mu.RUnlock()
	if !ok || rec.blockedTill == nil {
		return false
	}
	if time.Now().Before(*rec.blockedTill) {
		return true
	}
	// unblock expired
	b.mu.Lock()
	rec.blockedTill = nil
	b.mu.Unlock()
	return false
}

func (b *IPBlocker) noteFailure(ip string) {
	now := time.Now()

	b.mu.Lock()
	rec, ok := b.records[ip]
	if !ok {
		rec = &ipRecord{failures: 0, firstSeen: now}
		b.records[ip] = rec
	}

	// reset window if expired
	if now.Sub(rec.firstSeen) > b.window {
		rec.failures = 0
		rec.firstSeen = now
	}

	rec.failures++
	if rec.failures >= b.threshold {
		until := now.Add(b.blockDuration)
		rec.blockedTill = &until
		b.logger.WithComponent("ip_block").Warn("IP temporarily blocked due to suspicious activity",
			"client_ip", ip,
			"failures", rec.failures,
			"block_until", until.Format(time.RFC3339))
	}
	b.mu.Unlock()
}

func (b *IPBlocker) cleanup() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			b.gc()
		case <-b.cleanupC:
			return
		}
	}
}

func (b *IPBlocker) gc() {
	now := time.Now()
	b.mu.Lock()
	for ip, rec := range b.records {
		if rec.blockedTill == nil && now.Sub(rec.firstSeen) > 2*b.window {
			delete(b.records, ip)
		}
	}
	b.mu.Unlock()
}

func sliceToSet(values []string) map[string]struct{} {
	m := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}
	return m
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
