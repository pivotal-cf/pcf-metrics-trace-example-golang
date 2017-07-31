package middlewares

import (
	"fmt"
	"net/http"
	"time"
	"math/rand"
)

var (
	traceIdHeader = "X-B3-TraceId"
	spanIdHeader  = "X-B3-SpanId"
	parentSpanIdHeader = "X-B3-ParentSpanId"
)

type proxyMiddleware struct {
	path      string
	targetUrl string
	next      http.Handler
}

func NewProxyMiddleware(path, targetUrl string, next http.Handler) http.Handler {
	return proxyMiddleware{
		path:      path,
		targetUrl: targetUrl,
		next:      next,
	}
}

func (p proxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := p.proxy(r)
	if err != nil {
		fmt.Printf("Failed to forward request %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to forward request"))
		return
	}

	p.next.ServeHTTP(w, r)
}

func (p proxyMiddleware) proxy(proxyReq *http.Request) error {
	time.Sleep(time.Duration(500) * time.Millisecond)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", p.targetUrl, p.path), nil)
	if err != nil {
		fmt.Printf("Cannot create proxy request %s", err.Error())
		return err
	}

	traceId := proxyReq.Header.Get(traceIdHeader)
	previousSpanId := proxyReq.Header.Get(spanIdHeader)

	req.Header.Add(traceIdHeader, traceId)
	req.Header.Add(parentSpanIdHeader, previousSpanId)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	req.Header.Add(spanIdHeader, fmt.Sprintf("%x", r1.Uint64()))

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	_, err = client.Do(req)

	return err
}
