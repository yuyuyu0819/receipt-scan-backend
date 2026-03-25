package http

import (
	"net"
	stdhttp "net/http"
	"sync"

	"golang.org/x/time/rate"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.limiters[ip] = lim
	return lim
}

// loginLimiter: 1IPあたり毎分5回まで
var loginLimiter = newIPRateLimiter(rate.Every(60e9/5), 5)

// LoginRateLimitMiddleware はログイン・登録エンドポイントへのブルートフォース攻撃を防ぎます。
func LoginRateLimitMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !loginLimiter.getLimiter(ip).Allow() {
			stdhttp.Error(w, "too many requests", stdhttp.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
