package settings

import (
	"testing"
)

func TestValidate_PermittedCurrencyAndTimezone(t *testing.T) {
	cases := []struct {
		name     string
		currency string
		vatRate  string
		timezone string
		valid    bool
	}{
		{"valid", "EUR", "19", "Europe/Berlin", true},
		{"vat rate zero", "EUR", "0", "Europe/Berlin", true},
		{"vat rate one", "EUR", "100", "Europe/Berlin", true},
		{"vat rate below range", "EUR", "-1", "Europe/Berlin", false},
		{"vat rate above range", "EUR", "101", "Europe/Berlin", false},
		{"invalid currency", "XYZ", "19", "Europe/Berlin", false},
		{"blank currency", "", "19", "Europe/Berlin", false},
		{"invalid timezone", "EUR", "19", "Mars/Olympus", false},
		{"blank timezone", "EUR", "19", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := Form{Currency: tc.currency, VatRate: tc.vatRate, Timezone: tc.timezone}
			got := f.Validate()
			if got != tc.valid {
				t.Errorf("currency=%q timezone=%q: Validate() = %v, want %v", tc.currency, tc.timezone, got, tc.valid)
			}
		})
	}
}
