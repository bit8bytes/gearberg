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

import "context"

// MemberEnforcer implements Enforcer using org membership as the sole criterion.
// A subject is allowed if and only if they are a member of the org (obj); act is ignored.
type MemberEnforcer struct {
	svc *Service
}

// NewMemberEnforcer returns a MemberEnforcer backed by the given Service.
func NewMemberEnforcer(svc *Service) *MemberEnforcer {
	return &MemberEnforcer{svc: svc}
}

// Enforce returns true when sub is a member of org obj.
func (e *MemberEnforcer) Enforce(ctx context.Context, sub, obj, _ string) (bool, error) {
	if err := e.svc.GetMember(ctx, obj, sub); err == nil {
		// GetMember returns sql.ErrNoRows for non-members, not a caller error
		return true, nil
	}
	return false, nil
}
