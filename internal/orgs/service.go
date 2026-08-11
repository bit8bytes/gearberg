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
	"database/sql"
	"fmt"
)

type orgRepository interface {
	Create(ctx context.Context, tx *sql.Tx, accountID, id, displayName string) (string, error)
	List(ctx context.Context, accountID string) ([]Org, error)
	Get(ctx context.Context, id string) (Org, error)
	Update(ctx context.Context, p UpdateParams) error
	Delete(ctx context.Context, orgID, accountID string) error
}

type memberRepository interface {
	CreateMember(ctx context.Context, tx *sql.Tx, accountID, orgID string, roleID int64) error
	GetMember(ctx context.Context, orgID, accountID string) error
}

// Service provides organization business logic.
type Service struct {
	db      *sql.DB
	orgs    orgRepository
	members memberRepository
}

// NewService creates a new organization service.
func NewService(db *sql.DB, orgs orgRepository, members memberRepository) *Service {
	return &Service{db: db, orgs: orgs, members: members}
}

// List returns organizations the given account is a member of.
func (s *Service) List(ctx context.Context, accountID string) ([]Org, error) {
	records, err := s.orgs.List(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("orgs.Service.List: %w", err)
	}
	return records, nil
}

// GetFirstByAccountID returns the ID of the first organization the account belongs to.
func (s *Service) GetFirstByAccountID(ctx context.Context, accountID string) (string, error) {
	orgs, err := s.orgs.List(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("orgs.Service.GetFirstByAccountID: %w", err)
	}
	if len(orgs) == 0 {
		return "", fmt.Errorf("orgs.Service.GetFirstByAccountID: no orgs found for account")
	}
	return orgs[0].ID, nil
}

// Get retrieves an organization by ID.
func (s *Service) Get(ctx context.Context, id string) (*Org, error) {
	record, err := s.orgs.Get(ctx, id)
	if err != nil {
		return &Org{}, fmt.Errorf("orgs.Service.Get: %w", err)
	}
	return &record, nil
}

// Update saves the organization fields.
func (s *Service) Update(ctx context.Context, p UpdateParams) error {
	if err := s.orgs.Update(ctx, p); err != nil {
		return fmt.Errorf("orgs.Service.Update: %w", err)
	}
	return nil
}

// Delete removes an org if accountID is its sole owner.
func (s *Service) Delete(ctx context.Context, orgID, accountID string) error {
	if err := s.orgs.Delete(ctx, orgID, accountID); err != nil {
		return fmt.Errorf("orgs.Service.Delete: %w", err)
	}
	return nil
}

// GetMember returns an error if accountID is not a member of orgID.
func (s *Service) GetMember(ctx context.Context, orgID, accountID string) error {
	if err := s.members.GetMember(ctx, orgID, accountID); err != nil {
		return fmt.Errorf("orgs.Service.GetMember: %w", err)
	}
	return nil
}

// CreateOrg holds the data required to create a new org.
type CreateOrg struct {
	ID          string
	DisplayName string
	AccountID   string
}

// Create creates a new org and adds the given account as owner, enforcing the configured MaxOrgs limit.
func (s *Service) Create(ctx context.Context, createOrg CreateOrg) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("orgs.Create: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	org, err := s.orgs.Create(ctx, tx, createOrg.AccountID, createOrg.ID, createOrg.DisplayName)
	if err != nil {
		return "", fmt.Errorf("failed to create org: %w", err)
	}

	if err := s.members.CreateMember(ctx, tx, createOrg.AccountID, org, int64(OwnerRole)); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return org, nil
}
