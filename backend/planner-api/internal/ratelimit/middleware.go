package ratelimit

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultCleanupInterval = 5 * time.Minute

type Config struct {
	RequestsPerWindow int
	Window            time.Duration
	CleanupInterval   time.Duration
}

type Store struct {
	mu              sync.Mutex
	clients         map[string]*clientWindow
	requests        int
	window          time.Duration
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

type clientWindow struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

func NewStore(cfg Config) *Store {
	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = defaultCleanupInterval
	}

	return &Store{
		clients:         make(map[string]*clientWindow),
		requests:        cfg.RequestsPerWindow,
		window:          cfg.Window,
		cleanupInterval: cleanupInterval,
		lastCleanup:     time.Now().UTC(),
	}
}

func (s *Store) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if s == nil || s.requests <= 0 || s.window <= 0 || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			allowed, retryAfter := s.allow(clientKey(r), time.Now().UTC())
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(retryAfter)))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (s *Store) allow(key string, now time.Time) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupStaleEntries(now)

	client, ok := s.clients[key]
	if !ok || now.Sub(client.windowStart) >= s.window {
		s.clients[key] = &clientWindow{
			count:       1,
			windowStart: now,
			lastSeen:    now,
		}
		return true, s.window
	}

	client.lastSeen = now
	if client.count < s.requests {
		client.count++
		return true, client.windowStart.Add(s.window).Sub(now)
	}

	return false, client.windowStart.Add(s.window).Sub(now)
}

func (s *Store) cleanupStaleEntries(now time.Time) {
	if now.Sub(s.lastCleanup) < s.cleanupInterval {
		return
	}

	for key, client := range s.clients {
		if now.Sub(client.lastSeen) >= s.window+s.cleanupInterval {
			delete(s.clients, key)
		}
	}

	s.lastCleanup = now
}

func clientKey(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if ip != "" {
				return ip
			}
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-Ip")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	if remoteAddr == "" {
		return "unknown"
	}

	return remoteAddr
}

func retryAfterSeconds(retryAfter time.Duration) int {
	if retryAfter <= 0 {
		return 1
	}

	return int(math.Ceil(retryAfter.Seconds()))
}
