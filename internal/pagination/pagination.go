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

// Package pagination provides types and helpers for paginating query results.
package pagination

// Filters holds pagination and sort parameters parsed from an HTTP request.
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// Metadata holds pagination state for a single page of results.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
	PrevPage     int `json:"prev_page,omitempty"`
	NextPage     int `json:"next_page,omitempty"`
}

// CalculateMetadata computes pagination metadata for the given total, page, and page size.
// Uses (totalRecords + pageSize - 1) / pageSize to round up the last page without floats.
func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	lastPage := (totalRecords + pageSize - 1) / pageSize
	nextPage := page + 1
	if lastPage == 0 || nextPage > lastPage {
		nextPage = max(1, lastPage)
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: totalRecords,
		PrevPage:     max(page-1, 1),
		NextPage:     nextPage,
	}
}

// Limit returns the maximum number of records to fetch for the current page.
func (f Filters) Limit() int {
	return f.PageSize
}

// Offset returns the number of records to skip for the current page.
func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}
