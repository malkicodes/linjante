package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"malki.codes/linjante/server/errorhandler"

	"golang.org/x/time/rate"
)

const (
	Red    = "\x1b[31m"
	Green  = "\x1b[32m"
	Yellow = "\x1b[33m"
	Blue   = "\x1b[34m"
	Gray   = "\x1b[90m"
	Reset  = "\x1b[0m"
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
		errorhandler.HandleServerError(w, err)
		return
	}

	limiter := rl.getLimiter(host)

	if !limiter.Allow() {
		errorhandler.HandleRateLimitError(w)
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

type WrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *WrappedResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		wrapped := &WrappedResponseWriter{
			ResponseWriter: w,
		}

		start := time.Now()

		next.ServeHTTP(wrapped, r)

		elapsed := time.Since(start)

		var statusCodeColor string

		switch {
		case wrapped.statusCode >= 500:
			statusCodeColor = Red
		case wrapped.statusCode >= 400:
			statusCodeColor = Yellow
		case wrapped.statusCode >= 200 && wrapped.statusCode < 300:
			statusCodeColor = Green
		default:
			statusCodeColor = Blue
		}

		log.Printf("[%s] %s%s%s %d%s %s", r.Method, Gray, r.URL.EscapedPath(), statusCodeColor, wrapped.statusCode, Reset, elapsed)
	})
}
