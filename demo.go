// Package plugindemo a demo plugin.
package plugindemo

import (
	"context"
	"net/http"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	Enabled bool `json:"enabled,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled: true,
	}
}

// Demo a Demo plugin.
type Demo struct {
	next    http.Handler
	name    string
	enabled bool
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Demo{
		next:    next,
		name:    name,
		enabled: config.Enabled,
	}, nil
}

func (a *Demo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	splitRemoteAddr := strings.Split(req.RemoteAddr, ":")

	// If the remote address is not in the expected format, fail gracefully and just pass the request
	if len(splitRemoteAddr) != 2 {
		a.next.ServeHTTP(rw, req)
		return
	}

	remoteIP := splitRemoteAddr[0]

	if req.Header.Get("X-Forwarded-For") == "" {
		req.Header.Set("X-Forwarded-For", remoteIP)
	}

	a.next.ServeHTTP(rw, req)
}
