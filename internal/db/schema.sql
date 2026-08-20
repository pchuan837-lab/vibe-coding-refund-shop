-- ============================================================
-- refund-shop schema.sql · orders / refunds 两表
-- 说明：本文件为「第一轮基线」单文件 schema，B3 坏味道 2 锚点
--       （B3 读者后续将拆成 migrations/001 + 002 + schema_version）
-- ============================================================

-- orders 订单表
CREATE TABLE IF NOT EXISTS orders (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    product_name TEXT    NOT NULL,
    -- 金额一律用「分」（整数），禁止浮点：NFR-6 红线
    amount       INTEGER NOT NULL CHECK (amount > 0),
    shipping     INTEGER NOT NULL DEFAULT 0 CHECK (shipping >= 0),
    coupon_used  INTEGER NOT NULL DEFAULT 0 CHECK (coupon_used >= 0),
    status       TEXT    NOT NULL DEFAULT 'paid'
                 CHECK (status IN ('paid','partial_refunded','fully_refunded')),
    created_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- refunds 退款单表
CREATE TABLE IF NOT EXISTS refunds (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id    INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount      INTEGER NOT NULL CHECK (amount > 0),
    status      TEXT    NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','approved','rejected')),
    reason      TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_orders_status  ON orders(status);
CREATE INDEX IF NOT EXISTS idx_refunds_order   ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status  ON refunds(status);