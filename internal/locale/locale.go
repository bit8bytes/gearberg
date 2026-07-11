// Package locale maps Accept-Language headers to org setting defaults.
package locale

import (
	"strings"
)

// Defaults holds the pre-selected currency and timezone for a locale.
type Defaults struct {
	Currency string
	Timezone string
}

// fallback is returned when no locale match is found.
var fallback = Defaults{Currency: "USD", Timezone: "UTC"}

// localeDefaults maps BCP-47 tags (region-specific first, language-only second)
// to sensible currency and timezone defaults.
// Only currencies from settings.PermittedCurrencies and timezones from
// settings.PermittedTimezones are used here.
var localeDefaults = map[string]Defaults{
	// German-speaking
	"de":    {Currency: "EUR", Timezone: "Europe/Berlin"},
	"de-de": {Currency: "EUR", Timezone: "Europe/Berlin"},
	"de-at": {Currency: "EUR", Timezone: "Europe/Vienna"},
	"de-ch": {Currency: "CHF", Timezone: "Europe/Zurich"},

	// English-speaking
	"en":    {Currency: "USD", Timezone: "America/New_York"},
	"en-us": {Currency: "USD", Timezone: "America/New_York"},
	"en-gb": {Currency: "GBP", Timezone: "Europe/London"},
	"en-au": {Currency: "AUD", Timezone: "Australia/Sydney"},
	"en-ca": {Currency: "CAD", Timezone: "America/Toronto"},
	"en-nz": {Currency: "NZD", Timezone: "Pacific/Auckland"},
	"en-ie": {Currency: "EUR", Timezone: "Europe/Dublin"},

	// French-speaking
	"fr":    {Currency: "EUR", Timezone: "Europe/Paris"},
	"fr-fr": {Currency: "EUR", Timezone: "Europe/Paris"},
	"fr-be": {Currency: "EUR", Timezone: "Europe/Brussels"},
	"fr-ch": {Currency: "CHF", Timezone: "Europe/Zurich"},
	"fr-ca": {Currency: "CAD", Timezone: "America/Toronto"},

	// Dutch-speaking
	"nl":    {Currency: "EUR", Timezone: "Europe/Amsterdam"},
	"nl-nl": {Currency: "EUR", Timezone: "Europe/Amsterdam"},
	"nl-be": {Currency: "EUR", Timezone: "Europe/Brussels"},

	// Italian
	"it":    {Currency: "EUR", Timezone: "Europe/Rome"},
	"it-it": {Currency: "EUR", Timezone: "Europe/Rome"},

	// Spanish-speaking
	"es":    {Currency: "EUR", Timezone: "Europe/Madrid"},
	"es-es": {Currency: "EUR", Timezone: "Europe/Madrid"},
	"es-mx": {Currency: "MXN", Timezone: "America/Mexico_City"},
	"es-ar": {Currency: "USD", Timezone: "America/Argentina/Buenos_Aires"},

	// Portuguese-speaking
	"pt":    {Currency: "EUR", Timezone: "Europe/Lisbon"},
	"pt-pt": {Currency: "EUR", Timezone: "Europe/Lisbon"},
	"pt-br": {Currency: "BRL", Timezone: "America/Sao_Paulo"},

	// Nordic
	"sv":    {Currency: "SEK", Timezone: "Europe/Stockholm"},
	"sv-se": {Currency: "SEK", Timezone: "Europe/Stockholm"},
	"nb":    {Currency: "NOK", Timezone: "Europe/Oslo"},
	"no":    {Currency: "NOK", Timezone: "Europe/Oslo"},
	"no-no": {Currency: "NOK", Timezone: "Europe/Oslo"},
	"da":    {Currency: "DKK", Timezone: "Europe/Oslo"},
	"da-dk": {Currency: "DKK", Timezone: "Europe/Oslo"},
	"fi":    {Currency: "EUR", Timezone: "Europe/Helsinki"},
	"fi-fi": {Currency: "EUR", Timezone: "Europe/Helsinki"},

	// Eastern European
	"pl":    {Currency: "PLN", Timezone: "Europe/Warsaw"},
	"pl-pl": {Currency: "PLN", Timezone: "Europe/Warsaw"},
	"cs":    {Currency: "CZK", Timezone: "Europe/Prague"},
	"cs-cz": {Currency: "CZK", Timezone: "Europe/Prague"},
	"hu":    {Currency: "HUF", Timezone: "Europe/Budapest"},
	"hu-hu": {Currency: "HUF", Timezone: "Europe/Budapest"},
	"ro":    {Currency: "RON", Timezone: "Europe/Athens"},
	"ro-ro": {Currency: "RON", Timezone: "Europe/Athens"},
	"el":    {Currency: "EUR", Timezone: "Europe/Athens"},
	"el-gr": {Currency: "EUR", Timezone: "Europe/Athens"},

	// Turkish
	"tr":    {Currency: "TRY", Timezone: "Europe/Istanbul"},
	"tr-tr": {Currency: "TRY", Timezone: "Europe/Istanbul"},

	// East Asian
	"ja":    {Currency: "JPY", Timezone: "Asia/Tokyo"},
	"ja-jp": {Currency: "JPY", Timezone: "Asia/Tokyo"},
	"zh":    {Currency: "CNY", Timezone: "Asia/Shanghai"},
	"zh-cn": {Currency: "CNY", Timezone: "Asia/Shanghai"},
	"zh-hk": {Currency: "HKD", Timezone: "Asia/Hong_Kong"},
	"zh-tw": {Currency: "USD", Timezone: "Asia/Tokyo"},
	"zh-sg": {Currency: "SGD", Timezone: "Asia/Singapore"},

	// South / Southeast Asian
	"hi":    {Currency: "INR", Timezone: "Asia/Kolkata"},
	"hi-in": {Currency: "INR", Timezone: "Asia/Kolkata"},
	"ms-sg": {Currency: "SGD", Timezone: "Asia/Singapore"},

	// Middle Eastern
	"ar-sa": {Currency: "SAR", Timezone: "Asia/Riyadh"},
	"ar-ae": {Currency: "AED", Timezone: "Asia/Dubai"},

	// South African
	"en-za": {Currency: "ZAR", Timezone: "Africa/Johannesburg"},
}

// FromAcceptLanguage parses the Accept-Language header and returns the best
// matching Defaults. It checks region-specific tags before language-only tags
// and falls back to USD/UTC when no match is found.
func FromAcceptLanguage(header string) Defaults {
	if header == "" {
		return fallback
	}

	for part := range strings.SplitSeq(header, ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if d, ok := localeDefaults[tag]; ok {
			return d
		}
		// Try language-only (e.g. "de" from "de-DE")
		if idx := strings.IndexByte(tag, '-'); idx > 0 {
			if d, ok := localeDefaults[tag[:idx]]; ok {
				return d
			}
		}
	}

	return fallback
}
