-- name: GetAccountIDByProviderSubject :one
SELECT account_id
FROM federated_identities
WHERE provider_id = ?
AND provider_subject = ?
LIMIT 1;

-- name: Create :exec
INSERT INTO federated_identities (
    account_id,
    provider_id,
    provider_subject
) VALUES (
    ?, ?, ?
);
