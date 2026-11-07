DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM products WHERE sku = 'ELC-0001') THEN
        RETURN;
    END IF;

    INSERT INTO products (sku, name, price, stock, is_active, created_at)
    SELECT v.sku, v.name, v.price, v.stock, v.is_active, now() - v.age
    FROM (VALUES
        ('ELC-0001', 'Wireless Barcode Scanner', 1250000, 40, TRUE, interval '21 days'),
        ('ELC-0002', 'Thermal Receipt Printer 58mm', 875000, 25, TRUE, interval '18 days'),
        ('ELC-0003', 'Handheld Inventory Terminal', 4500000, 6, TRUE, interval '2 hours 10 minutes'),
        ('ELC-0004', 'USB-C Charging Cable 2m', 45000, 150, TRUE, interval '15 days'),
        ('OFC-0001', 'A4 Copy Paper 80gsm Ream', 52000, 320, TRUE, interval '6 hours'),
        ('OFC-0002', 'Ballpoint Pen Box of 50', 78000, 90, TRUE, interval '3 days 4 hours'),
        ('OFC-0003', 'Thermal Label Roll 100x150mm', 36000, 14, TRUE, interval '12 days'),
        ('PKG-0001', 'Corrugated Box 40x30x30 Pack of 10', 95000, 200, TRUE, interval '16 days'),
        ('PKG-0002', 'Bubble Wrap Roll 50m', 135000, 35, TRUE, interval '14 days'),
        ('PKG-0003', 'Packing Tape 48mm x 100m', 15000, 500, TRUE, interval '13 days'),
        ('HND-0001', 'Cordless Drill 12V', 685000, 18, TRUE, interval '10 days'),
        ('HND-0002', 'Pallet Hand Truck 2.5T', 3250000, 4, TRUE, interval '45 minutes'),
        ('SAF-0001', 'Safety Helmet Yellow', 65000, 75, TRUE, interval '11 days'),
        ('SAF-0002', 'Safety Gloves Pair Size L', 28000, 120, FALSE, interval '25 days'),
        ('CLN-0001', 'Industrial Degreaser 5L', 185000, 30, TRUE, interval '8 days'),
        ('FRN-0001', 'Steel Shelving Unit 5-Tier', 1450000, 14, FALSE, interval '30 days')
    ) AS v(sku, name, price, stock, is_active, age)
    ON CONFLICT (sku) DO NOTHING;

    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT id, 'in', stock, 'opening stock', created_at + interval '10 minutes'
    FROM products
    WHERE sku IN (
        'ELC-0001', 'ELC-0002', 'ELC-0003', 'ELC-0004', 'OFC-0001', 'OFC-0002', 'OFC-0003', 'PKG-0001',
        'PKG-0002', 'PKG-0003', 'HND-0001', 'HND-0002', 'SAF-0001', 'SAF-0002', 'CLN-0001', 'FRN-0001'
    );

    WITH written_off AS (
        UPDATE products SET stock = 0 WHERE sku = 'CLN-0001' RETURNING id, created_at
    )
    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT id, 'out', 30, 'write-off: damaged in transit', now() - interval '5 days'
    FROM written_off;

    WITH lines(sku, qty) AS (
        VALUES ('ELC-0001', 5), ('ELC-0002', 3), ('ELC-0004', 20)
    ),
    priced AS (
        SELECT l.sku, l.qty, p.price FROM lines l JOIN products p ON p.sku = l.sku
    ),
    new_order AS (
        INSERT INTO sales_orders (customer, total, created_at)
        SELECT 'PT Maju Bersama', SUM(qty * price), now() - interval '3 days' FROM priced
        RETURNING id, created_at
    ),
    new_items AS (
        INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
        SELECT o.id, pr.sku, pr.qty, pr.price FROM new_order o CROSS JOIN priced pr
        RETURNING sku, qty
    ),
    decremented AS (
        UPDATE products p SET stock = p.stock - i.qty FROM new_items i WHERE p.sku = i.sku
        RETURNING p.id, i.qty
    )
    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT d.id, 'out', d.qty, 'order #' || o.id, o.created_at FROM decremented d CROSS JOIN new_order o;

    WITH lines(sku, qty) AS (
        VALUES ('PKG-0001', 30), ('PKG-0002', 6), ('PKG-0003', 50), ('OFC-0003', 6)
    ),
    priced AS (
        SELECT l.sku, l.qty, p.price FROM lines l JOIN products p ON p.sku = l.sku
    ),
    new_order AS (
        INSERT INTO sales_orders (customer, total, created_at)
        SELECT 'CV Sinar Logistik', SUM(qty * price), now() - interval '1 day 3 hours' FROM priced
        RETURNING id, created_at
    ),
    new_items AS (
        INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
        SELECT o.id, pr.sku, pr.qty, pr.price FROM new_order o CROSS JOIN priced pr
        RETURNING sku, qty
    ),
    decremented AS (
        UPDATE products p SET stock = p.stock - i.qty FROM new_items i WHERE p.sku = i.sku
        RETURNING p.id, i.qty
    )
    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT d.id, 'out', d.qty, 'order #' || o.id, o.created_at FROM decremented d CROSS JOIN new_order o;

    WITH lines(sku, qty) AS (
        VALUES ('HND-0001', 2), ('SAF-0001', 10)
    ),
    priced AS (
        SELECT l.sku, l.qty, p.price FROM lines l JOIN products p ON p.sku = l.sku
    ),
    new_order AS (
        INSERT INTO sales_orders (customer, total, created_at)
        SELECT 'Toko Berkah Abadi', SUM(qty * price), now() - interval '2 days' FROM priced
        RETURNING id, created_at
    ),
    new_items AS (
        INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
        SELECT o.id, pr.sku, pr.qty, pr.price FROM new_order o CROSS JOIN priced pr
        RETURNING sku, qty
    ),
    decremented AS (
        UPDATE products p SET stock = p.stock - i.qty FROM new_items i WHERE p.sku = i.sku
        RETURNING p.id, i.qty
    )
    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT d.id, 'out', d.qty, 'order #' || o.id, o.created_at FROM decremented d CROSS JOIN new_order o;

    WITH cancelled AS (
        UPDATE sales_orders
        SET status = 'CANCELLED', cancelled_at = created_at + interval '2 hours'
        WHERE customer = 'Toko Berkah Abadi'
        RETURNING id, cancelled_at
    ),
    restored AS (
        UPDATE products p SET stock = p.stock + i.qty
        FROM sales_order_items i
        JOIN cancelled c ON c.id = i.order_id
        WHERE p.sku = i.sku
        RETURNING p.id, i.qty, c.id AS order_id, c.cancelled_at
    )
    INSERT INTO stock_movements (product_id, type, qty, note, created_at)
    SELECT id, 'in', qty, 'cancel order #' || order_id, cancelled_at FROM restored;
END
$$;
