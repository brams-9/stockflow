-- name: ClaimIdempotencyKey :execrows
INSERT INTO idempotency_keys (key, request_hash)
VALUES ($1, $2)
ON CONFLICT (key) DO NOTHING;

-- name: GetIdempotencyKey :one
SELECT request_hash, order_id
FROM idempotency_keys
WHERE key = $1;

-- name: AttachIdempotencyKeyOrder :exec
UPDATE idempotency_keys
SET order_id = $2
WHERE key = $1;
