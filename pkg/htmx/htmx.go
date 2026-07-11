// Package htmx provides helpers for working with HTMX HTTP headers.
package htmx

import (
	"net/http"
	urlpkg "net/url"
)

// Redirect sends an HTMX client-side redirect by setting the HX-Redirect header.
// The url may be a path relative to the request path and will be resolved accordingly.
//
// The provided code should be in the 3xx range and is usually [StatusMovedPermanently], [StatusFound], [StatusSeeOther], or [StatusOK].
//
// This function is specifically designed for HTMX requests. For standard HTTP redirects, use http.Redirect instead.
func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	// Let http.Redirect do the URL normalization
	if u, err := urlpkg.Parse(url); err == nil {
		if u.Scheme == "" && u.Host == "" {
			// Use stdlib's URL resolution
			url = r.URL.ResolveReference(u).String()
		}
	}

	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(code)
}

// StatusOK writes a 200 OK status code to the response.
// This is useful for HTMX requests that don't require a response body.
func StatusOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// IsRequest checks if the request is an HTMX request
// by looking for the "HX-IsRequest" header.
func IsRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// SetTrigger sets the HX-Trigger response header, instructing HTMX to fire
// a client-side event with the given name after the response is processed.
func SetTrigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}
