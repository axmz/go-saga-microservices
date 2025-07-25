CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    item_ids VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('Pending', 'AwaitingPayment', 'Paid', 'Failed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
