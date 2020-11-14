package middleware

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/logger"
)

func NewHealthCheck(paths, userAgents []string, store sessions.SessionStore) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return healthCheck(paths, userAgents, store, next)
	}
}

func healthCheck(paths, userAgents []string, store sessions.SessionStore, next http.Handler) http.Handler {
	// Use a map as a set to check health check paths
	pathSet := make(map[string]struct{})
	for _, path := range paths {
		if len(path) > 0 {
			pathSet[path] = struct{}{}
		}
	}

	// Use a map as a set to check health check paths
	userAgentSet := make(map[string]struct{})
	for _, userAgent := range userAgents {
		if len(userAgent) > 0 {
			userAgentSet[userAgent] = struct{}{}
		}
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if isHealthCheckRequest(pathSet, userAgentSet, req) {
			err := store.Healthcheck(req)
			if err != nil {
				logger.Errorf("session store healthcheck failed: %v", err)
				rw.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			rw.WriteHeader(http.StatusOK)
			fmt.Fprintf(rw, "OK")
			return
		}

		next.ServeHTTP(rw, req)
	})
}

func isHealthCheckRequest(paths, userAgents map[string]struct{}, req *http.Request) bool {
	if _, ok := paths[req.URL.EscapedPath()]; ok {
		return true
	}
	if _, ok := userAgents[req.Header.Get("User-Agent")]; ok {
		return true
	}
	return false
}
