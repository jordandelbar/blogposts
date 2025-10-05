package handlers

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	limiters map[string]*ipLimiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

func NewIPRateLimiter(rps rate.Limit, burst int) *IPRateLimiter {
	irl := &IPRateLimiter{
		limiters: make(map[string]*ipLimiter),
		rps:      rps,
		burst:    burst,
	}

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			irl.cleanupOldEntries()
		}
	}()

	return irl
}

func (irl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	irl.mu.Lock()
	defer irl.mu.Unlock()

	limiter, exists := irl.limiters[ip]
	if !exists {
		limiter = &ipLimiter{
			limiter:  rate.NewLimiter(irl.rps, irl.burst),
			lastSeen: time.Now(),
		}
		irl.limiters[ip] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

func (irl *IPRateLimiter) cleanupOldEntries() {
	irl.mu.Lock()
	defer irl.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	for ip, limiter := range irl.limiters {
		if limiter.lastSeen.Before(cutoff) {
			delete(irl.limiters, ip)
		}
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if net.ParseIP(clientIP) != nil {
				return clientIP
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}