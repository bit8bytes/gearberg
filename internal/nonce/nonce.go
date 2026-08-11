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

// Package nonce provides per-request CSP nonce generation and context propagation.
package nonce

import "context"

type contextKey string

const key contextKey = "nonce_id"

// NewContext returns a new context with the given trace ID stored.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// From returns the trace ID stored in ctx, or an empty string if none.
func From(ctx context.Context) Nonce {
	v, _ := ctx.Value(key).(string)
	return Nonce(v)
}

// Nonce is a per-request random value used as a CSP nonce.
type Nonce string

func (n *Nonce) String() string {
	return string(*n)
}
