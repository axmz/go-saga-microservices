-- Create products table with status constraint
-- Status can only be: 'available', 'reserved', 'sold'
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'reserved', 'sold')),
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample products for POC
INSERT INTO products (name, sku, status, price) VALUES
('Widget A', 'WIDGET-A', 'available', 19.99),
('Widget B', 'WIDGET-B', 'available', 29.99),
('Widget C', 'WIDGET-C', 'available', 9.99),
('Widget R', 'WIDGET-R', 'reserved', 9.99),
('Widget S', 'WIDGET-S', 'sold', 9.99),
('Premium Widget', 'PREMIUM-WIDGET', 'available', 49.99),
('Basic Widget', 'BASIC-WIDGET', 'available', 5.99); 