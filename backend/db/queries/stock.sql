-- name: DecreaseStock :one
UPDATE products
SET stock = stock - sqlc.arg(qty)
WHERE id = sqlc.arg(id) AND stock >= sqlc.arg(qty)
RETURNING stock;

-- name: IncreaseStock :one
UPDATE products
SET stock = stock + sqlc.arg(qty)
WHERE id = sqlc.arg(id) AND stock <= 2147483647 - sqlc.arg(qty)
RETURNING stock;

-- name: GetProductStock :one
SELECT stock FROM products WHERE id = $1;

-- name: InsertStockMovement :one
INSERT INTO stock_movements (product_id, type, qty, note)
VALUES ($1, $2, $3, $4)
RETURNING id, product_id, type, qty, note, created_at;

-- name: ListStockMovements :many
SELECT id, product_id, type, qty, note, created_at
FROM stock_movements
WHERE product_id = $1
ORDER BY created_at DESC, id DESC;
