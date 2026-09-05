// Package locale resolves a user's preferred language and vends pre-built
// message.Printers. It is transport-agnostic: cookie middleware, session
// middleware, and tests all call WithTag to store the resolved tag; handlers
// read it back with TagFrom or PrinterFrom.
package locale

import (
	"context"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Key for storing the resolved language tag.
type Key struct{}

// TagFrom returns the resolved tag stored in ctx, falling back to English.
func TagFrom(ctx context.Context) language.Tag {
	if tag, ok := ctx.Value(Key{}).(language.Tag); ok {
		return tag
	}
	return language.English
}

// PrinterFrom returns a Printer for the tag stored in ctx, falling back to English.
func PrinterFrom(ctx context.Context) *message.Printer {
	return message.NewPrinter(TagFrom(ctx))
}
