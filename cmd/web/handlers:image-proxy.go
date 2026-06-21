package main

import (
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

	addrs, err := net.LookupHost(u.Hostname())
	if err != nil {
		http.Error(w, "Cannot resolve host.", http.StatusBadRequest)
		return
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil || isPrivateIP(ip) {
			http.Error(w, fmt.Sprintf("Host %q is not allowed.", u.Hostname()), http.StatusForbidden)
			return
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		http.Error(w, "Failed to fetch image.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	const maxSize = 20 * 1024 * 1024
	result, err := image.Process(http.MaxBytesReader(w, resp.Body, maxSize))
	if err != nil {
		http.Error(w, "Image could not be processed: unsupported format or too large.", http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", result.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Data)
}
