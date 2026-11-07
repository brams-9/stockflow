CREATE TABLE idempotency_keys (
    key          TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    order_id     INTEGER REFERENCES sales_orders(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
