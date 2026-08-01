package orgs

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFormValidate(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantValid   bool
	}{
		{"valid name", "Acme Corp", true},
		{"empty name", "", false},
		{"whitespace only", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Form{DisplayName: strings.TrimSpace(tt.displayName)}
			if got := f.Validate(); got != tt.wantValid {
				t.Errorf("Validate() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		postName    string
		wantDisplay string
	}{
		{"trims whitespace", "  Acme  ", "Acme"},
		{"plain name", "Acme Corp", "Acme Corp"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := url.Values{"name": {tt.postName}}.Encode()
			r, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			form, err := Parse(r)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if form.DisplayName != tt.wantDisplay {
				t.Errorf("DisplayName = %q, want %q", form.DisplayName, tt.wantDisplay)
			}
		})
	}
}
