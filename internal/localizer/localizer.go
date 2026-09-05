// Package localizer resolves a user's preferred language from a cookie or
// Accept-Language header and makes a pre-built message.Printer available via
// the Localizer map. Keeping the Printer in the Localizer (not the context)
// makes the dependency explicit in handler signatures.
package localizer

import (
	"context"
	"net/http"

	_ "github.com/bit8bytes/gearberg/internal/translations"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type tagKey struct{}

// TagFrom returns the resolved tag stored by the middleware. Falls back to
// English when the middleware has not run.
func TagFrom(ctx context.Context) language.Tag {
	tag, ok := ctx.Value(tagKey{}).(language.Tag)
	if !ok {
		return language.English
	}
	return tag
}

// Localizer resolves language tags and vends pre-built message.Printers.
// Construct with New; the zero value is not usable.
type Localizer struct {
	cookieName string
	tags       []language.Tag
	printers   map[language.Tag]*message.Printer
}

// New builds a Localizer for the given supported tags. Panics on empty input
// because a localizer with no languages is always misconfigured.
func New(tags ...language.Tag) *Localizer {
	if len(tags) == 0 {
		panic("at least one language tag must be provided")
	}
	printers := make(map[language.Tag]*message.Printer)
	for _, tag := range tags {
		printers[tag] = message.NewPrinter(tag)
	}
	return &Localizer{cookieName: "locale", tags: tags, printers: printers}
}

// Printer returns the pre-built Printer for the given tag, falling back to
// English so callers never have to handle a nil printer.
func (l *Localizer) Printer(tag language.Tag) *message.Printer {
	if p, ok := l.printers[tag]; ok {
		return p
	}
	return l.printers[language.English]
}

// Handler is an http.Handler middleware that reads the locale cookie and
// Accept-Language header, resolves the best-matching tag, and stores only the
// tag in the request context. Handlers fetch the Printer explicitly via Printer(tag).
func (l *Localizer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var locale string
		if c, err := r.Cookie(l.cookieName); err == nil {
			locale = c.Value
		}
		tag := l.Resolve(locale, r.Header.Get("Accept-Language"))
		ctx := context.WithValue(r.Context(), tagKey{}, tag)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Resolve picks the best supported tag from an explicit locale string and an
// Accept-Language header value. Exposed so tests and non-HTTP code can resolve
// a tag without constructing a fake request.
func (l *Localizer) Resolve(locale, acceptLanguage string) language.Tag {
	matcher := language.NewMatcher(l.tags)
	tag, _ := language.MatchStrings(matcher, locale, acceptLanguage)
	return tag
}
