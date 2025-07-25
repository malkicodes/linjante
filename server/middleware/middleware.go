package middleware

import (
	"linjante/server/errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const RATELIMIT_PER rate.Limit = 5
const RATELIMIT_BURST int = 1
const RATELIMIT_AUTOCLEAR_INTERVAL time.Duration = 15 * time.Minute
const RATELIMIT_AUTOCLEAR_WHEN time.Duration = time.Minute

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*Visitor

	next http.Handler
}

type Visitor struct {
	limiter     *rate.Limiter
	lastRequest int64
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, prs := rl.visitors[ip]

	if prs {
		visitor.lastRequest = time.Now().Unix()
	} else {
		visitor = &Visitor{
			limiter:     rate.NewLimiter(RATELIMIT_PER, RATELIMIT_BURST),
			lastRequest: time.Now().Unix(),
		}
		rl.visitors[ip] = visitor
	}

	return visitor.limiter
}

func (rl *RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		errors.HandleServerError(w, err)
		return
	}

	limiter := rl.getLimiter(host)

	if !limiter.Allow() {
		errors.HandleRateLimitError(w)
		return
	}

	rl.next.ServeHTTP(w, r)
}

func RateLimitMiddleware(next http.Handler) *RateLimiter {
	rateLimiter := RateLimiter{
		next:     next,
		visitors: make(map[string]*Visitor),
	}

	ticker := time.NewTicker(RATELIMIT_AUTOCLEAR_INTERVAL)

	go func() {
		for {
			// Blocks until rate limit autoclear interval
			<-ticker.C

			rateLimiter.mu.Lock()

			now := time.Now().Unix()
			amountCleared := uint(0)

			for ip, v := range rateLimiter.visitors {
				if now-v.lastRequest > int64(RATELIMIT_AUTOCLEAR_WHEN.Seconds()) {
					delete(rateLimiter.visitors, ip)
					amountCleared++
				}
			}

			rateLimiter.mu.Unlock()
		}
	}()

	return &rateLimiter
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s", r.Method, r.URL.EscapedPath())
	})
}
