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

// Package templates holds the HTML templates embedded into the binary so the
// application can be deployed as a single self-contained executable without
// requiring template files to exist on the host filesystem.
package templates

import "embed"

// EmbedFS contains all template directories embedded at compile time.
//
//go:embed components layouts pages partials fragments
var EmbedFS embed.FS
