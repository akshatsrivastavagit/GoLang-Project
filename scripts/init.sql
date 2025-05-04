-- Warehouses
CREATE TABLE IF NOT EXISTS warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(255)
);

-- Products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL
);

-- Inventory (stock per warehouse)
CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    warehouse_id INT REFERENCES warehouses(id),
    quantity INT NOT NULL,
    UNIQUE(product_id, warehouse_id)
);

-- Stock Levels
CREATE TABLE IF NOT EXISTS stock_levels (
    sku VARCHAR(100) NOT NULL,
    warehouse_id INT NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (sku, warehouse_id)
);

-- Inventory Transactions
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(100) NOT NULL,
    warehouse_id INT NOT NULL,
    quantity INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 