package test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter tracks rate limiters per IP address (copied for testing)
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// AddIP creates a new rate limiter for an IP address
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the given IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()
	return limiter
}

// GetIPAddress extracts the client IP address from the request
func GetIPAddress(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// Test_NewIPRateLimiter tests rate limiter creation
func Test_NewIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 5)

	if limiter == nil {
		t.Fatal("Expected non-nil rate limiter")
	}

	if limiter.r != rate.Every(time.Second) {
		t.Errorf("Expected rate %v, got %v", rate.Every(time.Second), limiter.r)
	}

	if limiter.b != 5 {
		t.Errorf("Expected burst size 5, got %d", limiter.b)
	}

	if limiter.ips == nil {
		t.Error("Expected initialized IP map")
	}
}

// Test_GetLimiter tests getting a limiter for an IP
func Test_GetLimiter(t *testing.T) {
	rateLimiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	ip := "192.168.1.1"

	limiter := rateLimiter.GetLimiter(ip)
	if limiter == nil {
		t.Fatal("Expected non-nil limiter for IP")
	}

	// Get again, should return the same limiter
	limiter2 := rateLimiter.GetLimiter(ip)
	if limiter != limiter2 {
		t.Error("Expected same limiter instance for same IP")
	}
}

// Test_RateLimitEnforcement tests that rate limiting actually works
func Test_RateLimitEnforcement(t *testing.T) {
	// Create a rate limiter: 5 requests per minute (burst of 1)
	rateLimiter := NewIPRateLimiter(rate.Every(time.Minute/5), 1)
	ip := "192.168.1.1"

	limiter := rateLimiter.GetLimiter(ip)

	// First request should be allowed
	if !limiter.Allow() {
		t.Error("Expected first request to be allowed")
	}

	// Subsequent rapid requests should be denied
	deniedCount := 0
	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			deniedCount++
		}
	}

	if deniedCount < 8 {
		t.Errorf("Expected at least 8 requests to be denied, got %d", deniedCount)
	}
}

// Test_MultipleIPAddresses tests that different IPs get different limiters
func Test_MultipleIPAddresses(t *testing.T) {
	rateLimiter := NewIPRateLimiter(rate.Every(time.Minute/5), 1)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	limiter1 := rateLimiter.GetLimiter(ip1)
	limiter2 := rateLimiter.GetLimiter(ip2)

	// Should be different limiters
	if limiter1 == limiter2 {
		t.Error("Expected different limiters for different IPs")
	}

	// First request from each IP should be allowed
	if !limiter1.Allow() {
		t.Error("Expected first request from IP1 to be allowed")
	}

	if !limiter2.Allow() {
		t.Error("Expected first request from IP2 to be allowed")
	}

	// Second rapid request from IP1 should be denied
	if limiter1.Allow() {
		t.Error("Expected second rapid request from IP1 to be denied")
	}

	// But IP2's second request should still be denied (it already used its token)
	if limiter2.Allow() {
		t.Error("Expected second rapid request from IP2 to be denied")
	}
}

// Test_ConcurrentAccess tests thread safety
func Test_ConcurrentAccess(t *testing.T) {
	rateLimiter := NewIPRateLimiter(rate.Every(time.Second), 10)
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	wg.Add(numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune(id%256))
			limiter := rateLimiter.GetLimiter(ip)
			_ = limiter.Allow()
		}(i)
	}
	
	wg.Wait()
	// If we get here without panicking, thread safety works
}

// Test_GetIPAddress_XForwardedFor tests X-Forwarded-For header extraction
func Test_GetIPAddress_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	ip := GetIPAddress(req)
	if ip != "203.0.113.1" {
		t.Errorf("Expected IP 203.0.113.1, got %s", ip)
	}
}

// Test_GetIPAddress_XRealIP tests X-Real-IP header extraction
func Test_GetIPAddress_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := GetIPAddress(req)
	if ip != "198.51.100.1" {
		t.Errorf("Expected IP 198.51.100.1, got %s", ip)
	}
}

// Test_GetIPAddress_RemoteAddr tests fallback to RemoteAddr
func Test_GetIPAddress_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	ip := GetIPAddress(req)
	if ip != "192.0.2.1:12345" {
		t.Errorf("Expected IP 192.0.2.1:12345, got %s", ip)
	}
}

// Test_GetIPAddress_Priority tests header priority
func Test_GetIPAddress_Priority(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "198.51.100.1")
	req.RemoteAddr = "192.0.2.1:12345"

	// X-Forwarded-For should take priority
	ip := GetIPAddress(req)
	if ip != "203.0.113.1" {
		t.Errorf("Expected X-Forwarded-For to take priority, got %s", ip)
	}
}

// Test_RateLimitRecovery tests that limits recover over time
func Test_RateLimitRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-dependent test in short mode")
	}

	// 1 request per second, burst of 1
	rateLimiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	ip := "192.168.1.1"

	limiter := rateLimiter.GetLimiter(ip)

	// Use up the initial token
	if !limiter.Allow() {
		t.Error("Expected first request to be allowed")
	}

	// Immediate second request should be denied
	if limiter.Allow() {
		t.Error("Expected immediate second request to be denied")
	}

	// Wait for token to replenish (1.1 seconds to be safe)
	time.Sleep(1100 * time.Millisecond)

	// Now it should be allowed again
	if !limiter.Allow() {
		t.Error("Expected request after waiting to be allowed")
	}
}

// Benchmark_GetLimiter benchmarks the performance of getting a limiter
func Benchmark_GetLimiter(b *testing.B) {
	rateLimiter := NewIPRateLimiter(rate.Every(time.Second), 10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := "192.168.1." + string(rune(i%256))
		rateLimiter.GetLimiter(ip)
	}
}

// Benchmark_Allow benchmarks the performance of checking rate limits
func Benchmark_Allow(b *testing.B) {
	rateLimiter := NewIPRateLimiter(rate.Every(time.Second), 1000)
	limiter := rateLimiter.GetLimiter("192.168.1.1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}
