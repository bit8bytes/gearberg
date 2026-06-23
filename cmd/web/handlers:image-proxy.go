package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bit8bytes/gearberg/internal/image"
)

var privateRanges []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"100.64.0.0/10",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	} {
		_, network, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, network)
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, network := range privateRanges {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// getImageProxy fetches an image from an external https:// URL and streams it back.
// It blocks requests to private/loopback IPs to prevent SSRF.
func (app *application) getImageProxy(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "Missing url parameter.", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		http.Error(w, "Only https:// URLs are allowed.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := validateProxyHost(ctx, u.Hostname()); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	result, status, err := fetchAndProcessImage(ctx, rawURL, w)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Data)
}

func validateProxyHost(ctx context.Context, hostname string) error {
	resolver := &net.Resolver{}
	addrs, err := resolver.LookupHost(ctx, hostname) //nolint:gosec // SSRF mitigated: HTTPS-only + private-IP block below
	if err != nil {
		return fmt.Errorf("cannot resolve host")
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil || isPrivateIP(ip) {
			return fmt.Errorf("host %q is not allowed", hostname)
		}
	}
	return nil
}

func fetchAndProcessImage(ctx context.Context, rawURL string, w http.ResponseWriter) (*image.ProcessResult, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil) //nolint:gosec // SSRF mitigated by validateProxyHost
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed to build request")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req) //nolint:gosec // SSRF mitigated by validateProxyHost
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("failed to fetch image")
	}
	defer resp.Body.Close() //nolint:errcheck // response body close errors are not actionable

	const maxSize = 20 * 1024 * 1024
	result, err := image.Process(http.MaxBytesReader(w, resp.Body, maxSize))
	if err != nil {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("image could not be processed: unsupported format or too large")
	}
	return result, http.StatusOK, nil
}
