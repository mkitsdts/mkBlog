package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func normalizeServerURL(raw string, tlsEnabled bool, defaultPort int) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("site.server is empty")
	}

	scheme := "http"
	if tlsEnabled {
		scheme = "https"
	}

	var u *url.URL
	var err error
	if strings.Contains(raw, "://") {
		u, err = url.Parse(raw)
	} else {
		u, err = url.Parse("//" + raw)
	}
	if err != nil {
		return "", fmt.Errorf("invalid server %q: %w", raw, err)
	}

	host := u.Host
	if host == "" && u.Path != "" && !strings.Contains(u.Path, "/") {
		host = u.Path
	}
	if host == "" {
		return "", fmt.Errorf("invalid server %q: missing host", raw)
	}

	hostName, port := splitHostPort(host)
	if hostName == "" {
		return "", fmt.Errorf("invalid server %q: missing host", raw)
	}
	if port == "" {
		if defaultPort > 0 {
			port = fmt.Sprintf("%d", defaultPort)
		} else if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(hostName, port)), nil
}

func splitHostPort(host string) (string, string) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", ""
	}
	if h, p, err := net.SplitHostPort(host); err == nil {
		return trimIPv6Brackets(h), p
	}
	return trimIPv6Brackets(host), ""
}

func trimIPv6Brackets(h string) string {
	if strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(h, "["), "]")
	}
	return h
}
