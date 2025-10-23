package handler

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func (h *accrualHandler) rateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := h.extractIP(r)

			h.mutex.Lock()
			limiter, exists := h.rateLimiters[ip]
			if !exists {
				limit := rate.Limit(h.config.RateLimitPerMin) / 60.0
				limiter = rate.NewLimiter(limit, h.config.RateLimitPerMin) 
				h.rateLimiters[ip] = limiter
				h.logger.Debug("New rate limiter created",
					zap.String("ip", ip),
					zap.Int("total_limiters", len(h.rateLimiters)))
			}
			h.mutex.Unlock()

			allowed := limiter.Allow()
			h.logger.Debug("Rate limit check",
				zap.String("ip", ip),
				zap.Bool("allowed", allowed),
				zap.Float64("limit", float64(limiter.Limit())))

			if !allowed {
				h.logger.Warn("Rate limit exceeded", zap.String("ip", ip))
				w.Header().Set("Retry-After", fmt.Sprintf("%d", h.config.RetryAfterSec))
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, "No more than %d requests per minute allowed", h.config.RateLimitPerMin)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (h *accrualHandler) extractIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
