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

// Package flash provides one-shot notification messages stored in the session
// and rendered as toasts on the next page load (post-redirect-get pattern).
package flash

import "context"

// Type controls the DaisyUI alert style of a notification.
type Type string

// Notification types mapped to DaisyUI alert styles.
const (
	Success Type = "success"
	Error   Type = "error"
)

// Flash is a one-shot notification message displayed as a toast.
type Flash struct {
	Message string
	Type    Type
}

// Manager stores and retrieves flash notifications across requests.
type Manager interface {
	// Put stores a flash notification to be shown on the next page load.
	Put(ctx context.Context, msg string, t Type)
	// Pop retrieves and removes the flash notification.
	// Returns nil if no flash is set.
	Pop(ctx context.Context) *Flash
}
