-- Create settings table
CREATE TABLE settings (
    setting_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service, key)
);

-- Insert default settings for all services
INSERT INTO settings (service, key, value, description) VALUES
-- General Database Settings
('General', 'DB_HOST', 'icecream_postgres', 'Database host address'),
('General', 'DB_PORT', '5432', 'Database port number'),
('General', 'DB_USER', 'postgres', 'Database username'),
('General', 'DB_PASSWORD', 'postgres123', 'Database password'),
('General', 'DB_NAME', 'icecream_store', 'Database name'),
('General', 'DB_SSL_MODE', 'disable', 'Database SSL mode'),

-- General Logging Settings
('General', 'LOG_LEVEL', 'info', 'Logging level (debug, info, warn, error)'),

-- JWT Settings
('General', 'JWT_SECRET', 'icecream-super-secret-jwt-key-change-in-production-2024', 'JWT signing secret'),
('General', 'JWT_EXPIRATION_TIME', '30m', 'JWT token expiration time'),

-- Business Settings
('General', 'DEFAULT_TAX_RATE', '13.0', 'Default tax rate percentage (Costa Rica IVA)'),
('General', 'DEFAULT_SERVICE_RATE', '10.0', 'Default service charge rate percentage'),
('General', 'ORDER_TIMEOUT', '30', 'Order timeout in minutes'),

-- Gateway Service Settings
('Gateway', 'GATEWAY_PORT', '8082', 'Port for the gateway service'),
('Gateway', 'SESSION_SERVICE_URL', 'http://icecream_session:8081', 'URL for session service'),
('Gateway', 'ORDERS_SERVICE_URL', 'http://icecream_orders:8083', 'URL for orders service'),
('Gateway', 'INVENTORY_SERVICE_URL', 'http://icecream_inventory:8084', 'URL for inventory service'),
('Gateway', 'INVOICE_SERVICE_URL', 'http://icecream_invoice:8085', 'URL for invoice service'),

-- Session Service Settings
('Session', 'SESSION_SERVER_PORT', '8081', 'Port for the session service'),

-- Orders Service Settings
('Orders', 'SERVER_PORT', '8083', 'Port for the orders service'),

-- Inventory Service Settings
('Inventory', 'INVENTORY_SERVER_PORT', '8084', 'Port for the inventory service'),

-- Invoice Service Settings
('Invoice', 'INVOICE_SERVER_PORT', '8085', 'Port for the invoice service'),
('Invoice', 'INVENTORY_SERVICE_URL', 'http://icecream_inventory:8084', 'URL for inventory service'),

-- Data Service Settings
('Data', 'DATA_SERVER_PORT', '8086', 'Port for the data service'),
('Data', 'DB_MAX_OPEN_CONNS', '25', 'Maximum number of open connections'),
('Data', 'DB_MAX_IDLE_CONNS', '5', 'Maximum number of idle connections'),
('Data', 'DB_CONN_MAX_LIFETIME', '5m', 'Maximum lifetime of connections'),
('Data', 'DB_CONN_MAX_IDLE_TIME', '5m', 'Maximum idle time of connections'),
('Data', 'DB_CONNECT_TIMEOUT', '10s', 'Database connection timeout'),
('Data', 'DB_QUERY_TIMEOUT', '30s', 'Database query timeout'),
('Data', 'DB_MAX_RETRIES', '3', 'Maximum number of retry attempts'),
('Data', 'DB_RETRY_INTERVAL', '1s', 'Interval between retry attempts'),

-- UI Service Settings
('UI', 'UI_PORT', '3000', 'Port for the UI service'),
('UI', 'GATEWAY_URL', 'http://icecream_gateway:8082', 'Gateway service URL');

-- Create indexes
CREATE INDEX idx_settings_service ON settings(service);
CREATE INDEX idx_settings_key ON settings(key);
CREATE INDEX idx_settings_service_key ON settings(service, key);
