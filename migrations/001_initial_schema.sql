
CREATE TABLE shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    location VARCHAR(200),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id),
    name VARCHAR(100) NOT NULL,
    pin_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'cashier', -- cashier | manager | admin
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_hi VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES categories(id),
    name VARCHAR(200) NOT NULL,
    name_hi VARCHAR(200),
    price NUMERIC(10,2) NOT NULL,
    tax_rate NUMERIC(5,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE shop_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id),
    item_id UUID REFERENCES items(id),
    stock_qty INT DEFAULT 0,
    low_stock_threshold INT DEFAULT 10,
    UNIQUE(shop_id, item_id)
);


CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id),
    staff_id UUID REFERENCES staff(id),
    total_amount NUMERIC(10,2) NOT NULL,
    discount_amount NUMERIC(10,2) DEFAULT 0,
    payment_mode VARCHAR(10) NOT NULL, -- upi | cash
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending | confirmed | failed
    razorpay_order_id VARCHAR(100),
    razorpay_payment_id VARCHAR(100),
    cash_received NUMERIC(10,2),
    change_given NUMERIC(10,2),
    synced_from_terminal BOOLEAN DEFAULT false,
    terminal_txn_id VARCHAR(100), -- original SQLite ID from terminal
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE transaction_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id),
    item_id UUID REFERENCES items(id),
    quantity INT NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL,
    total_price NUMERIC(10,2) NOT NULL
);

CREATE TABLE terminal_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id),
    last_heartbeat TIMESTAMPTZ,
    last_sync_at TIMESTAMPTZ,
    pending_queue_size INT DEFAULT 0,
    is_online BOOLEAN DEFAULT false
);


CREATE TABLE restock_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id),
    item_id UUID REFERENCES items(id),
    requested_qty INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending | approved | fulfilled
    created_at TIMESTAMPTZ DEFAULT NOW()
);