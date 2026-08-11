// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
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
