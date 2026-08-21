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
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
)

// testOptions returns an *options configured for the test database at dir.
func testOptions(dir string) *options {
	return &options{
		DbDsn:      "file:" + dir + "/test.db",
		StorageDSN: "file://" + dir,
		Limits: Limits{
			MaxOrgs:             10,
			MaxOrgCategories:    25,
			MaxOrgManufacturers: 100,
			MaxOrgLocations:     100,
			MaxStorageBytes:     1 << 30,
		},
	}
}

// Shared test credentials used across test files.
const (
	aliceEmail = "alice@example.com"
	bobEmail   = "bob@example.com"
	passwd1    = "R2f*ttkvdBA43vREwW7cdLxF3VzUPJ8U" //nolint:gosec // test credentials
	passwd2    = "Tg4Wkd8XebChc_R*D.H.92b268pBeC2m" //nolint:gosec // test credentials
)

// Define a custom testServer type which embeds a httptest.Server instance.
type testServer struct {
	*httptest.Server
}

func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	jar, _ := cookiejar.New(nil)
	ts.Client().Jar = jar

	// Prevent the client from following redirects so tests can assert on 3xx responses.
	ts.Client().CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) signout(t *testing.T) {
	if code, _, _ := ts.postForm(t, "/signout", url.Values{}); code != http.StatusSeeOther {
		t.Fatalf("signout failed: expected redirect, got status %d", code)
	}
}

func (ts *testServer) signin(t *testing.T, email, password string) {
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	if code, _, _ := ts.postForm(t, "/signin", form); code != http.StatusSeeOther {
		t.Fatalf("signin failed: expected redirect, got status %d", code)
	}
}

func (ts *testServer) get(t *testing.T, urlPath string) (code int, header http.Header, body []byte) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+urlPath, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rs.Body.Close() }()
	body, err = io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, body
}

func (ts *testServer) postForm(tb testing.TB, urlPath string, form url.Values) (code int, header http.Header, body []byte) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))
	if err != nil {
		tb.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rs, err := ts.Client().Do(req)
	if err != nil {
		tb.Fatal(err)
	}

	defer func() { _ = rs.Body.Close() }()
	body, err = io.ReadAll(rs.Body)
	if err != nil {
		tb.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, body
}

// testUser represents an authenticated test user with known credentials and org context.
type testUser struct {
	email    string
	password string
	orgID    string
	ts       *testServer
}

func (u *testUser) signin(t *testing.T) {
	t.Helper()
	// Always clear any active session first. The /signin route is wrapped with
	// the guest middleware, which silently redirects authenticated users to /home
	// with the same 303 as a successful login — making it impossible to detect
	// that signin was blocked. Signing out first guarantees a clean slate.
	u.ts.signout(t)
	u.ts.signin(t, u.email, u.password)
}

func (u *testUser) signout(t *testing.T) {
	t.Helper()
	u.ts.signout(t)
}

// postFile uploads a single file as a multipart/form-data POST request.
func (ts *testServer) postFile(t *testing.T, urlPath, fieldName, fileName string, content []byte) (code int, header http.Header, body []byte) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fieldName, fileName))
	h.Set("Content-Type", "text/csv; charset=utf-8")
	fw, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("postFile: create part: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("postFile: write content: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("postFile: close writer: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+urlPath, &buf)
	if err != nil {
		t.Fatalf("postFile: new request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("postFile: do request: %v", err)
	}
	defer func() { _ = rs.Body.Close() }()
	body, err = io.ReadAll(rs.Body)
	if err != nil {
		t.Fatalf("postFile: read body: %v", err)
	}
	body = bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, body
}
