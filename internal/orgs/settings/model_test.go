package settings

import (
	"testing"
)

func TestVatRate_Percent(t *testing.T) {
	cases := []struct {
		name string
		v    VatRate
		want string
	}{
		{"whole percent", 1900, "19"},
		{"zero", 0, "0"},
		{"fractional", 850, "8.50"},
		{"one basis point", 1, "0.01"},
		{"hundred percent", 10000, "100"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.v.Percent()
			if got != tc.want {
				t.Errorf("VatRate(%d).Percent() = %q, want %q", tc.v, got, tc.want)
			}
		})
	}
}

func TestVatRate_Display(t *testing.T) {
	cases := []struct {
		name string
		v    VatRate
		want string
	}{
		{"zero returns empty", 0, ""},
		{"whole percent", 1900, "19 %"},
		{"fractional", 850, "8.5 %"},
		{"one basis point", 1, "0.01 %"},
		{"hundred percent", 10000, "100 %"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.v.Display()
			if got != tc.want {
				t.Errorf("VatRate(%d).Display() = %q, want %q", tc.v, got, tc.want)
			}
		})
	}
}
