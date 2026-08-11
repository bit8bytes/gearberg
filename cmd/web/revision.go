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
	"fmt"
	"time"
)

// revision is set at build time via -ldflags "-X main.revision=<value>".
// In Nix builds this is the flake revision; in local builds it defaults to
// "dev-<timestamp>" so that each process start busts the asset cache.
var revision = "dev"

var databaseVersion = "unknown"

func init() {
	if revision == "dev" {
		revision = fmt.Sprintf("dev-%d", time.Now().UnixMilli())
	}
}
