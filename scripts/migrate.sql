-- Drop existing table
DROP TABLE IF EXISTS inventory_transactions;

-- Create table with new schema
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(100) NOT NULL,
    warehouse_id INT NOT NULL,
    change INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 