-- name: GetFirstByAccountID :one
SELECT o.id
FROM orgs o
JOIN org_members om ON om.org_id = o.id
WHERE om.account_id = ?
ORDER BY o.created_at
LIMIT 1;

-- name: GetRoleByName :one
SELECT
    id,
    name,
    rank
FROM org_roles
WHERE name = ?
LIMIT 1;

-- name: Create :one
INSERT INTO orgs (
    id,
    display_name
)
SELECT ?, ?
WHERE (
    SELECT COUNT(*) FROM org_members WHERE account_id = sqlc.arg(account_id)
) < CAST(sqlc.arg(max_orgs) AS INTEGER)
RETURNING
    id,
    display_name,
    created_at;

-- name: Get :one
SELECT
    id,
    display_name,
    updated_at,
    created_at
FROM orgs
WHERE id = ?
LIMIT 1;

-- name: List :many
SELECT
    o.id,
    o.display_name,
    o.updated_at,
    o.created_at
FROM orgs o
JOIN org_members om ON om.org_id = o.id
WHERE om.account_id = ?
ORDER BY o.created_at;

-- name: Update :exec
UPDATE orgs
SET
    display_name = ?,
    updated_at = unixepoch()
WHERE id = ?;

-- name: CreateMember :exec
INSERT INTO org_members (
    account_id,
    org_id,
    role_id
) VALUES (
    ?, ?, ?
);

-- name: GetMemberByAccountIDAndOrgID :one
SELECT
    id,
    account_id,
    org_id,
    role_id,
    updated_at,
    created_at
FROM org_members
WHERE account_id = ?
AND org_id = ?
LIMIT 1;

-- name: ListMembersByOrgID :many
SELECT
    id,
    account_id,
    org_id,
    role_id,
    updated_at,
    created_at
FROM org_members
WHERE org_id = ?;

-- name: UpdateMemberRole :exec
UPDATE org_members
SET
    role_id = ?,
    updated_at = unixepoch()
WHERE account_id = ?
AND org_id = ?;

-- name: DeleteMember :exec
DELETE FROM org_members
WHERE account_id = ?
AND org_id = ?;

-- name: DeleteOwnedByAccountID :exec
-- Deletes orgs where the given account is the sole owner.
DELETE FROM orgs
WHERE id IN (
    SELECT om.org_id
    FROM org_members om
    WHERE om.account_id = ?
      AND om.role_id = 1
      AND (
          SELECT COUNT(*)
          FROM org_members om2
          WHERE om2.org_id = om.org_id
            AND om2.role_id = 1
      ) = 1
);

-- name: Delete :exec
-- Deletes a single org only if the given account is its sole owner.
DELETE FROM orgs
WHERE orgs.id = sqlc.arg(org_id)
  AND (SELECT COUNT(*) FROM org_members WHERE org_members.org_id = sqlc.arg(org_id) AND org_members.role_id = 1) = 1
  AND EXISTS (
      SELECT 1 FROM org_members om
      WHERE om.org_id = sqlc.arg(org_id) AND om.account_id = sqlc.arg(account_id) AND om.role_id = 1
  );
