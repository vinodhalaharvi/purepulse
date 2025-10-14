-- This script runs automatically when the container first starts
-- It sets up required PostgreSQL extensions

\c purepulse

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Verify extensions
SELECT extname, extversion 
FROM pg_extension 
WHERE extname IN ('uuid-ossp', 'pg_trgm', 'btree_gin', 'pg_stat_statements');

-- Create application user (if different from postgres user)
-- This is optional - only if you want separate users for different purposes
-- CREATE USER app_readonly WITH PASSWORD 'readonly_password';
-- CREATE USER app_readwrite WITH PASSWORD 'readwrite_password';

-- Log success
SELECT 'Database initialized successfully' AS status;


