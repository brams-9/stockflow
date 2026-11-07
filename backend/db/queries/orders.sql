-- name: ListProductsBySKU :many
SELECT id, sku, name, price, stock, is_active, created_at
FROM products
WHERE sku = ANY(sqlc.arg(skus)::text[]);

-- name: DecreaseStockBySKU :one
UPDATE products
SET stock = stock - sqlc.arg(qty)
WHERE sku = sqlc.arg(sku) AND is_active AND stock >= sqlc.arg(qty)
RETURNING id, price;

-- name: GetProductStateBySKU :one
SELECT stock, is_active FROM products WHERE sku = $1;

-- name: IncreaseStockBySKU :one
UPDATE products
SET stock = stock + sqlc.arg(qty)
WHERE sku = sqlc.arg(sku)
RETURNING id;

-- name: InsertOrder :one
INSERT INTO sales_orders (customer, total)
VALUES ($1, $2)
RETURNING id, customer, status, total, created_at, cancelled_at;

-- name: InsertOrderItem :one
INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
VALUES ($1, $2, $3, $4)
RETURNING id, order_id, sku, qty, unit_price;

-- name: GetOrder :one
SELECT id, customer, status, total, created_at, cancelled_at
FROM sales_orders
WHERE id = $1;

-- name: GetOrderForUpdate :one
SELECT id, customer, status, total, created_at, cancelled_at
FROM sales_orders
WHERE id = $1
FOR UPDATE;

-- name: ListOrderItems :many
SELECT id, order_id, sku, qty, unit_price
FROM sales_order_items
WHERE order_id = $1
ORDER BY id;

-- name: CancelOrder :one
UPDATE sales_orders
SET status = 'CANCELLED', cancelled_at = now()
WHERE id = $1
RETURNING id, customer, status, total, created_at, cancelled_at;
