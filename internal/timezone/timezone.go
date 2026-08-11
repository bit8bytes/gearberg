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

// Package timezone provides the Timezone domain type for IANA timezone identifiers.
package timezone

// IANA lists the IANA timezone identifiers exposed in the UI.
var IANA = []string{
	"Africa/Cairo", "Africa/Johannesburg", "Africa/Lagos", "Africa/Nairobi",
	"America/Anchorage", "America/Argentina/Buenos_Aires", "America/Bogota",
	"America/Chicago", "America/Denver", "America/Los_Angeles", "America/Mexico_City",
	"America/New_York", "America/Phoenix", "America/Sao_Paulo", "America/Toronto",
	"America/Vancouver",
	"Asia/Bangkok", "Asia/Colombo", "Asia/Dubai", "Asia/Hong_Kong", "Asia/Jakarta",
	"Asia/Karachi", "Asia/Kolkata", "Asia/Riyadh", "Asia/Seoul", "Asia/Shanghai",
	"Asia/Singapore", "Asia/Tokyo",
	"Atlantic/Azores",
	"Australia/Adelaide", "Australia/Brisbane", "Australia/Melbourne",
	"Australia/Perth", "Australia/Sydney",
	"Europe/Amsterdam", "Europe/Athens", "Europe/Berlin", "Europe/Brussels",
	"Europe/Budapest", "Europe/Dublin", "Europe/Helsinki", "Europe/Istanbul",
	"Europe/Lisbon", "Europe/London", "Europe/Madrid", "Europe/Moscow",
	"Europe/Oslo", "Europe/Paris", "Europe/Prague", "Europe/Rome",
	"Europe/Stockholm", "Europe/Vienna", "Europe/Warsaw", "Europe/Zurich",
	"Pacific/Auckland", "Pacific/Honolulu",
	"UTC",
}

// Timezone is an IANA timezone identifier (e.g. "Europe/Berlin").
type Timezone string

// String returns the timezone identifier as a plain string.
func (t Timezone) String() string { return string(t) }
