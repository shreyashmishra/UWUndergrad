package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStoreAllowResetsWindow(t *testing.T) {
	store := NewStore(Config{
		RequestsPerWindow: 2,
		Window:            time.Minute,
	})

	now := time.Unix(1_000, 0).UTC()

	allowed, _ := store.allow("127.0.0.1", now)
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}

	allowed, _ = store.allow("127.0.0.1", now.Add(10*time.Second))
	if !allowed {
		t.Fatal("expected second request to be allowed")
	}

	allowed, retryAfter := store.allow("127.0.0.1", now.Add(20*time.Second))
	if allowed {
		t.Fatal("expected third request to be rejected")
	}

	if retryAfter <= 0 {
		t.Fatal("expected retry-after duration for rejected request")
	}

	allowed, _ = store.allow("127.0.0.1", now.Add(61*time.Second))
	if !allowed {
		t.Fatal("expected request after window reset to be allowed")
	}
}

func TestMiddlewareSkipsOptionsRequests(t *testing.T) {
	store := NewStore(Config{
		RequestsPerWindow: 1,
		Window:            time.Minute,
	})

	handler := store.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/graphql", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected OPTIONS request to bypass limiter, got %d", res.Code)
	}
}

func TestMiddlewareRejectsRequestOverLimit(t *testing.T) {
	store := NewStore(Config{
		RequestsPerWindow: 1,
		Window:            time.Minute,
	})

	handler := store.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	firstReq := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	firstReq.RemoteAddr = "127.0.0.1:1234"
	firstRes := httptest.NewRecorder()
	handler.ServeHTTP(firstRes, firstReq)

	if firstRes.Code != http.StatusOK {
		t.Fatalf("expected first request to be allowed, got %d", firstRes.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	secondReq.RemoteAddr = "127.0.0.1:5678"
	secondRes := httptest.NewRecorder()
	handler.ServeHTTP(secondRes, secondReq)

	if secondRes.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be rejected, got %d", secondRes.Code)
	}

	if secondRes.Header().Get("Retry-After") == "" {
		t.Fatal("expected retry-after header on rejected request")
	}
}
