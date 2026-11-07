-- name: ListProducts :many
SELECT
    p.id,
    p.sku,
    p.name,
    p.price,
    p.stock,
    p.is_active,
    p.created_at,
    (SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id)::timestamptz AS last_movement_at
FROM products p
WHERE p.sku ILIKE sqlc.arg(pattern) OR p.name ILIKE sqlc.arg(pattern)
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountProducts :one
SELECT count(*)
FROM products
WHERE sku ILIKE sqlc.arg(pattern) OR name ILIKE sqlc.arg(pattern);

-- name: CreateProduct :one
INSERT INTO products (sku, name, price)
VALUES ($1, $2, $3)
RETURNING id, sku, name, price, stock, is_active, created_at;

-- name: GetProductByID :one
SELECT
    p.id,
    p.sku,
    p.name,
    p.price,
    p.stock,
    p.is_active,
    p.created_at,
    (SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id)::timestamptz AS last_movement_at
FROM products p
WHERE p.id = $1;

-- name: UpdateProduct :execrows
UPDATE products
SET sku = $2, name = $3, price = $4, is_active = $5
WHERE id = $1;

-- name: DeleteProduct :execrows
DELETE FROM products WHERE id = $1;

-- name: ProductHasHistory :one
SELECT (
    EXISTS (SELECT 1 FROM stock_movements WHERE product_id = sqlc.arg(id))
    OR EXISTS (
        SELECT 1
        FROM sales_order_items i
        JOIN products p ON p.sku = i.sku
        WHERE p.id = sqlc.arg(id)
    )
)::boolean AS has_history;
